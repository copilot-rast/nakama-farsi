package mongowire

import (
	"context"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const opMsg = 2013

type Options struct {
	Hosts      []string
	User       string
	Password   string
	AuthSource string
	TLS        bool
	TLSConfig  *tls.Config
	Timeout    time.Duration
	PoolSize   int
	Dialer     func(ctx context.Context, addr string) (net.Conn, error)
}

type Client struct {
	opt   Options
	mu    sync.Mutex
	idle  []*conn
	open  int
	reqID atomic.Int32
	wait  chan struct{}
}

type conn struct {
	c net.Conn
}

type CommandError struct {
	Code    int
	Message string
}

func (e *CommandError) Error() string { return fmt.Sprintf("mongodb error %d: %s", e.Code, e.Message) }

func New(opt Options) (*Client, error) {
	if len(opt.Hosts) == 0 {
		return nil, errors.New("mongowire: no hosts")
	}
	if opt.Timeout == 0 {
		opt.Timeout = 30 * time.Second
	}
	if opt.PoolSize <= 0 {
		opt.PoolSize = 4
	}
	if opt.AuthSource == "" {
		opt.AuthSource = "admin"
	}
	return &Client{opt: opt, wait: make(chan struct{}, opt.PoolSize)}, nil
}

func (cl *Client) Close() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	for _, c := range cl.idle {
		_ = c.c.Close()
	}
	cl.idle = nil
	return nil
}

func (cl *Client) dial(ctx context.Context, addr string) (net.Conn, error) {
	if cl.opt.Dialer != nil {
		return cl.opt.Dialer(ctx, addr)
	}
	d := net.Dialer{Timeout: cl.opt.Timeout}
	if cl.opt.TLS {
		cfg := cl.opt.TLSConfig
		if cfg == nil {
			host, _, _ := net.SplitHostPort(addr)
			cfg = &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		}
		return (&tls.Dialer{NetDialer: &d, Config: cfg}).DialContext(ctx, "tcp", addr)
	}
	return d.DialContext(ctx, "tcp", addr)
}

func (cl *Client) connect(ctx context.Context) (*conn, error) {
	var lastErr error
	tried := map[string]bool{}
	hosts := append([]string{}, cl.opt.Hosts...)
	for len(hosts) > 0 {
		addr := hosts[0]
		hosts = hosts[1:]
		if tried[addr] {
			continue
		}
		tried[addr] = true
		nc, err := cl.dial(ctx, addr)
		if err != nil {
			lastErr = err
			continue
		}
		c := &conn{c: nc}
		hello, err := cl.roundTrip(ctx, c, D{{"hello", int32(1)}, {"$db", "admin"}})
		if err != nil {
			_ = nc.Close()
			lastErr = err
			continue
		}
		if wp, _ := hello.Get("isWritablePrimary").(bool); !wp {
			if p, _ := hello.Get("primary").(string); p != "" && !tried[p] {
				_ = nc.Close()
				hosts = append([]string{p}, hosts...)
				continue
			}
		}
		if cl.opt.User != "" {
			if err := cl.auth(ctx, c); err != nil {
				_ = nc.Close()
				return nil, err
			}
		}
		return c, nil
	}
	if lastErr == nil {
		lastErr = errors.New("mongowire: no reachable primary")
	}
	return nil, lastErr
}

func (cl *Client) get(ctx context.Context) (*conn, error) {
	select {
	case cl.wait <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	cl.mu.Lock()
	if n := len(cl.idle); n > 0 {
		c := cl.idle[n-1]
		cl.idle = cl.idle[:n-1]
		cl.mu.Unlock()
		return c, nil
	}
	cl.mu.Unlock()
	c, err := cl.connect(ctx)
	if err != nil {
		<-cl.wait
		return nil, err
	}
	return c, nil
}

func (cl *Client) put(c *conn, broken bool) {
	if broken {
		_ = c.c.Close()
	} else {
		cl.mu.Lock()
		cl.idle = append(cl.idle, c)
		cl.mu.Unlock()
	}
	<-cl.wait
}

func (cl *Client) roundTrip(ctx context.Context, c *conn, cmd D) (D, error) {
	body, err := Marshal(cmd)
	if err != nil {
		return nil, err
	}
	id := cl.reqID.Add(1)
	msg := make([]byte, 16+4+1, 16+4+1+len(body))
	binary.LittleEndian.PutUint32(msg[4:], uint32(id))
	binary.LittleEndian.PutUint32(msg[12:], opMsg)
	msg = append(msg, body...)
	binary.LittleEndian.PutUint32(msg[0:], uint32(len(msg)))
	deadline := time.Now().Add(cl.opt.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = c.c.SetDeadline(deadline)
	if _, err := c.c.Write(msg); err != nil {
		return nil, err
	}
	return readReply(c.c, id)
}

func readReply(r io.Reader, id int32) (D, error) {
	var head [16]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return nil, err
	}
	size := int(binary.LittleEndian.Uint32(head[0:]))
	if size < 21 || size > 48<<20 {
		return nil, errors.New("mongowire: bad reply size")
	}
	if got := int32(binary.LittleEndian.Uint32(head[8:])); id != 0 && got != id {
		return nil, errors.New("mongowire: reply to another request")
	}
	if binary.LittleEndian.Uint32(head[12:]) != opMsg {
		return nil, errors.New("mongowire: reply is not OP_MSG")
	}
	rest := make([]byte, size-16)
	if _, err := io.ReadFull(r, rest); err != nil {
		return nil, err
	}
	if rest[4] != 0 {
		return nil, errors.New("mongowire: unexpected section kind")
	}
	d, _, err := readDoc(rest[5:])
	return d, err
}

func ReadRequest(r io.Reader) (int32, D, error) {
	var head [16]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return 0, nil, err
	}
	size := int(binary.LittleEndian.Uint32(head[0:]))
	if size < 21 || size > 48<<20 {
		return 0, nil, errors.New("mongowire: bad request size")
	}
	rest := make([]byte, size-16)
	if _, err := io.ReadFull(r, rest); err != nil {
		return 0, nil, err
	}
	d, _, err := readDoc(rest[5:])
	return int32(binary.LittleEndian.Uint32(head[4:])), d, err
}

func WriteReply(w io.Writer, to int32, doc D) error {
	body, err := Marshal(doc)
	if err != nil {
		return err
	}
	msg := make([]byte, 21, 21+len(body))
	binary.LittleEndian.PutUint32(msg[8:], uint32(to))
	binary.LittleEndian.PutUint32(msg[12:], opMsg)
	msg = append(msg, body...)
	binary.LittleEndian.PutUint32(msg[0:], uint32(len(msg)))
	_, err = w.Write(msg)
	return err
}

func okDoc(d D) error {
	ok, _ := Num(d.Get("ok"))
	if ok == 1 {
		return nil
	}
	code, _ := Num(d.Get("code"))
	msg, _ := d.Get("errmsg").(string)
	return &CommandError{Code: int(code), Message: msg}
}

func (cl *Client) Command(ctx context.Context, db string, cmd D) (D, error) {
	c, err := cl.get(ctx)
	if err != nil {
		return nil, err
	}
	out, err := cl.roundTrip(ctx, c, append(cmd, E{"$db", db}))
	var ce *CommandError
	cl.put(c, err != nil && !errors.As(err, &ce))
	if err != nil {
		return nil, err
	}
	return out, okDoc(out)
}

func (cl *Client) auth(ctx context.Context, c *conn) error {
	nonce := make([]byte, 18)
	_, _ = rand.Read(nonce)
	sc := newScram(cl.opt.User, cl.opt.Password, base64.StdEncoding.EncodeToString(nonce))
	first := sc.clientFirst()
	out, err := cl.roundTrip(ctx, c, D{{"saslStart", int32(1)}, {"mechanism", "SCRAM-SHA-256"}, {"payload", Binary{Data: []byte(first)}}, {"options", D{{"skipEmptyExchange", true}}}, {"$db", cl.opt.AuthSource}})
	if err == nil {
		err = okDoc(out)
	}
	if err != nil {
		return err
	}
	conv := out.Get("conversationId")
	payload, _ := out.Get("payload").(Binary)
	final, err := sc.clientFinal(string(payload.Data))
	if err != nil {
		return err
	}
	out, err = cl.roundTrip(ctx, c, D{{"saslContinue", int32(1)}, {"conversationId", conv}, {"payload", Binary{Data: []byte(final)}}, {"$db", cl.opt.AuthSource}})
	if err == nil {
		err = okDoc(out)
	}
	if err != nil {
		return err
	}
	payload, _ = out.Get("payload").(Binary)
	if err := sc.verify(string(payload.Data)); err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		if done, _ := out.Get("done").(bool); done {
			return nil
		}
		out, err = cl.roundTrip(ctx, c, D{{"saslContinue", int32(1)}, {"conversationId", conv}, {"payload", Binary{Data: []byte{}}}, {"$db", cl.opt.AuthSource}})
		if err == nil {
			err = okDoc(out)
		}
		if err != nil {
			return err
		}
	}
	return errors.New("mongowire: authentication did not finish")
}

type scram struct {
	user, pass, nonce string
	firstBare         string
	serverSig         []byte
}

func newScram(user, pass, nonce string) *scram { return &scram{user: user, pass: pass, nonce: nonce} }

func saslName(s string) string { return strings.NewReplacer("=", "=3D", ",", "=2C").Replace(s) }

func (s *scram) clientFirst() string {
	s.firstBare = "n=" + saslName(s.user) + ",r=" + s.nonce
	return "n,," + s.firstBare
}

func hmacSum(key, data []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(data)
	return m.Sum(nil)
}

func (s *scram) clientFinal(serverFirst string) (string, error) {
	fields := map[string]string{}
	for _, p := range strings.Split(serverFirst, ",") {
		if k, v, ok := strings.Cut(p, "="); ok {
			fields[k] = v
		}
	}
	r, salt64, iter := fields["r"], fields["s"], fields["i"]
	if !strings.HasPrefix(r, s.nonce) || salt64 == "" {
		return "", errors.New("mongowire: bad SCRAM server nonce")
	}
	salt, err := base64.StdEncoding.DecodeString(salt64)
	if err != nil {
		return "", err
	}
	n := 0
	for _, c := range iter {
		if c < '0' || c > '9' {
			return "", errors.New("mongowire: bad SCRAM iteration count")
		}
		n = n*10 + int(c-'0')
	}
	if n < 4096 && n != 0 {
		return "", errors.New("mongowire: SCRAM iteration count too low")
	}
	salted, err := pbkdf2.Key(sha256.New, s.pass, salt, n, 32)
	if err != nil {
		return "", err
	}
	clientKey := hmacSum(salted, []byte("Client Key"))
	stored := sha256.Sum256(clientKey)
	without := "c=biws,r=" + r
	authMsg := s.firstBare + "," + serverFirst + "," + without
	sig := hmacSum(stored[:], []byte(authMsg))
	proof := make([]byte, len(clientKey))
	for i := range clientKey {
		proof[i] = clientKey[i] ^ sig[i]
	}
	serverKey := hmacSum(salted, []byte("Server Key"))
	s.serverSig = hmacSum(serverKey, []byte(authMsg))
	return without + ",p=" + base64.StdEncoding.EncodeToString(proof), nil
}

func (s *scram) verify(serverFinal string) error {
	v, ok := strings.CutPrefix(serverFinal, "v=")
	if !ok {
		return errors.New("mongowire: SCRAM server signature missing")
	}
	got, err := base64.StdEncoding.DecodeString(v)
	if err != nil || !hmac.Equal(got, s.serverSig) {
		return errors.New("mongowire: SCRAM server signature mismatch")
	}
	return nil
}

func (cl *Client) InsertMany(ctx context.Context, db, coll string, docs []D) (int, error) {
	if len(docs) == 0 {
		return 0, nil
	}
	arr := make(A, len(docs))
	for i, d := range docs {
		arr[i] = d
	}
	out, err := cl.Command(ctx, db, D{{"insert", coll}, {"documents", arr}, {"ordered", false}})
	if err != nil {
		return 0, err
	}
	dups := 0
	if we, ok := out.Get("writeErrors").(A); ok {
		for _, item := range we {
			d, _ := item.(D)
			code, _ := Num(d.Get("code"))
			if code != 11000 {
				msg, _ := d.Get("errmsg").(string)
				return 0, &CommandError{Code: int(code), Message: msg}
			}
			dups++
		}
	}
	return dups, nil
}

func (cl *Client) Aggregate(ctx context.Context, db, coll string, pipeline []D, max int) ([]D, error) {
	stages := make(A, len(pipeline))
	for i, p := range pipeline {
		stages[i] = p
	}
	out, err := cl.Command(ctx, db, D{{"aggregate", coll}, {"pipeline", stages}, {"cursor", D{{"batchSize", int32(1000)}}}, {"allowDiskUse", true}})
	if err != nil {
		return nil, err
	}
	res := []D{}
	for {
		cur, _ := out.Get("cursor").(D)
		batch, _ := cur.Get("firstBatch").(A)
		if batch == nil {
			batch, _ = cur.Get("nextBatch").(A)
		}
		for _, item := range batch {
			if d, ok := item.(D); ok {
				res = append(res, d)
			}
		}
		if max > 0 && len(res) > max {
			return nil, fmt.Errorf("mongowire: more than %d results", max)
		}
		id, _ := cur.Get("id").(int64)
		if id == 0 {
			return res, nil
		}
		out, err = cl.Command(ctx, db, D{{"getMore", id}, {"collection", coll}, {"batchSize", int32(1000)}})
		if err != nil {
			return nil, err
		}
	}
}
