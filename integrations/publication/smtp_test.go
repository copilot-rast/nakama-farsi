package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"io"
	"math/big"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strings"
	"sync"
	"testing"
	"time"
)

type smtpSession struct {
	auth, from, to string
	data           string
	tls            bool
}

type fakeSMTP struct {
	addr     string
	cert     tls.Certificate
	implicit bool
	starttls bool
	rejectTo string
	mu       sync.Mutex
	sessions []smtpSession
	listener net.Listener
}

func selfSigned(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "127.0.0.1"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := x509.ParseCertificate(der)
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}, pool
}

func startSMTP(t *testing.T, cert tls.Certificate, implicit, starttls bool) *fakeSMTP {
	t.Helper()
	var ln net.Listener
	var err error
	if implicit {
		ln, err = tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	} else {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeSMTP{addr: ln.Addr().String(), cert: cert, implicit: implicit, starttls: starttls, listener: ln}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(conn)
		}
	}()
	return f
}

func (f *fakeSMTP) serve(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	sess := smtpSession{tls: f.implicit}
	r := bufio.NewReader(conn)
	w := func(s string) { _, _ = io.WriteString(conn, s+"\r\n") }
	w("220 fake ESMTP")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		cmd := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			ext := []string{"250-fake", "250-8BITMIME"}
			if f.starttls && !sess.tls {
				ext = append(ext, "250-STARTTLS")
			}
			ext = append(ext, "250 AUTH PLAIN")
			for _, e := range ext {
				w(e)
			}
		case cmd == "STARTTLS":
			w("220 go ahead")
			tc := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{f.cert}})
			if err := tc.Handshake(); err != nil {
				return
			}
			conn = tc
			r = bufio.NewReader(conn)
			sess.tls = true
		case strings.HasPrefix(cmd, "AUTH PLAIN"):
			raw, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(line[len("AUTH PLAIN"):]))
			sess.auth = string(raw)
			if strings.HasSuffix(sess.auth, "\x00wrong") {
				w("535 5.7.8 bad credentials")
				continue
			}
			w("235 ok")
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			sess.from = angle(line)
			w("250 ok")
		case strings.HasPrefix(cmd, "RCPT TO:"):
			sess.to = angle(line)
			if f.rejectTo != "" && sess.to == f.rejectTo {
				w("550 5.1.1 no such user")
				continue
			}
			w("250 ok")
		case cmd == "DATA":
			w("354 send")
			var b strings.Builder
			for {
				l, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if l == ".\r\n" {
					break
				}
				b.WriteString(l)
			}
			sess.data = b.String()
			f.mu.Lock()
			f.sessions = append(f.sessions, sess)
			f.mu.Unlock()
			w("250 queued")
		case cmd == "QUIT":
			w("221 bye")
			return
		default:
			w("502 unknown")
		}
	}
}

func angle(line string) string {
	start, end := strings.IndexByte(line, '<'), strings.IndexByte(line, '>')
	if start < 0 || end < start {
		return ""
	}
	return line[start+1 : end]
}

func (f *fakeSMTP) last(t *testing.T) smtpSession {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sessions) == 0 {
		t.Fatal("no message delivered")
	}
	return f.sessions[len(f.sessions)-1]
}

func smtpConfig(addr, security string) providerConfig {
	host, port, _ := net.SplitHostPort(addr)
	return providerConfig{"host": host, "port": port, "security": security, "username": "mailer", "password": "pw", "from": "news@example.com", "fromName": "خبرنامه بازی"}
}

func checkMessage(t *testing.T, sess smtpSession) {
	t.Helper()
	if sess.from != "news@example.com" || sess.to != "u@example.com" || sess.auth != "\x00mailer\x00pw" {
		t.Fatalf("session = %+v", sess)
	}
	msg, err := mail.ReadMessage(strings.NewReader(sess.data))
	if err != nil {
		t.Fatal(err)
	}
	dec := new(mime.WordDecoder)
	subject, err := dec.DecodeHeader(msg.Header.Get("Subject"))
	if err != nil || subject != "سلام دنیا" {
		t.Fatalf("subject = %q %v (raw %q)", subject, err, msg.Header.Get("Subject"))
	}
	if !strings.HasPrefix(msg.Header.Get("Subject"), "=?UTF-8?b?") {
		t.Fatalf("subject must be RFC 2047 encoded: %q", msg.Header.Get("Subject"))
	}
	from, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil || from.Name != "خبرنامه بازی" || from.Address != "news@example.com" {
		t.Fatalf("from = %v %v", from, err)
	}
	mediaType, params, _ := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if mediaType != "multipart/alternative" {
		t.Fatalf("content type = %s", mediaType)
	}
	mr := multipart.NewReader(msg.Body, params["boundary"])
	parts := map[string]string{}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(base64.NewDecoder(base64.StdEncoding, p))
		ct, ctParams, _ := mime.ParseMediaType(p.Header.Get("Content-Type"))
		if ctParams["charset"] != "UTF-8" {
			t.Fatalf("part %s charset = %s", ct, ctParams["charset"])
		}
		parts[ct] = string(raw)
	}
	if parts["text/plain"] != "متن پررنگ" {
		t.Fatalf("text part = %q", parts["text/plain"])
	}
	if !strings.Contains(parts["text/html"], "<strong>پررنگ</strong>") {
		t.Fatalf("html part = %q", parts["text/html"])
	}
}

func smtpMessage() outMessage {
	return outMessage{To: "u@example.com", Subject: "سلام دنیا", Text: markdownText("متن **پررنگ**"), HTML: emailHTML("سلام دنیا", renderMarkdown("متن **پررنگ**"), "fa")}
}

func TestSMTPStartTLS(t *testing.T) {
	cert, pool := selfSigned(t)
	srv := startSMTP(t, cert, false, true)
	s := testService()
	s.tls = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(srv.addr, "starttls")}, smtpMessage()); err != nil {
		t.Fatal(err)
	}
	sess := srv.last(t)
	if !sess.tls {
		t.Fatal("STARTTLS was not used")
	}
	checkMessage(t, sess)

	s.tls = nil
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(srv.addr, "starttls")}, smtpMessage()); err == nil {
		t.Fatal("an untrusted certificate must be refused")
	}
}

func TestSMTPImplicitTLS(t *testing.T) {
	cert, pool := selfSigned(t)
	srv := startSMTP(t, cert, true, false)
	s := testService()
	s.tls = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(srv.addr, "tls")}, smtpMessage()); err != nil {
		t.Fatal(err)
	}
	checkMessage(t, srv.last(t))
}

func TestSMTPPlainAndErrors(t *testing.T) {
	cert, _ := selfSigned(t)
	srv := startSMTP(t, cert, false, false)
	s := testService()
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(srv.addr, "none")}, smtpMessage()); err != nil {
		t.Fatal(err)
	}
	if sess := srv.last(t); sess.tls {
		t.Fatal("plain session must not be TLS")
	} else {
		checkMessage(t, sess)
	}
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(srv.addr, "starttls")}, smtpMessage()); err == nil || !isPermanent(err) {
		t.Fatalf("missing STARTTLS must fail permanently: %v", err)
	}
	srv.rejectTo = "u@example.com"
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(srv.addr, "none")}, smtpMessage()); err == nil || !isPermanent(err) || !strings.Contains(err.Error(), "550") {
		t.Fatalf("550 must be permanent: %v", err)
	}
	srv.rejectTo = ""
	cfg := smtpConfig(srv.addr, "none")
	cfg["password"] = "wrong"
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: cfg}, smtpMessage()); err == nil || !isPermanent(err) {
		t.Fatalf("535 must be permanent: %v", err)
	}
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	dead := ln.Addr().String()
	ln.Close()
	if _, err := s.deliver(context.Background(), provider{Kind: "smtp", Config: smtpConfig(dead, "none")}, smtpMessage()); err == nil || isPermanent(err) {
		t.Fatalf("connect failure must be retried: %v", err)
	}
}

func TestBuildMIMEStripsHeaderInjection(t *testing.T) {
	raw := string(buildMIME("a@example.com", "b@example.com", "hi\r\nBcc: evil@example.com", "t", "<p>t</p>", time.Now(), "example.com"))
	if strings.Contains(raw, "\r\nBcc:") {
		t.Fatalf("header injection: %q", raw)
	}
}
