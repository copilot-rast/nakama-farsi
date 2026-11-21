package mongowire

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"strings"
	"sync"
)

type Handler func(cmd D) D

type FakeServer struct {
	Listener net.Listener
	User     string
	Password string
	Handle   Handler
	mu       sync.Mutex
	Commands []D
	authed   map[net.Conn]bool
}

func NewFakeServer(user, password string, h Handler) (*FakeServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &FakeServer{Listener: ln, User: user, Password: password, Handle: h, authed: map[net.Conn]bool{}}
	go s.serve()
	return s, nil
}

func (s *FakeServer) Addr() string { return s.Listener.Addr().String() }

func (s *FakeServer) Close() error { return s.Listener.Close() }

func (s *FakeServer) serve() {
	for {
		c, err := s.Listener.Accept()
		if err != nil {
			return
		}
		go s.conn(c)
	}
}

func (s *FakeServer) conn(c net.Conn) {
	defer c.Close()
	var sc struct {
		first, serverFirst, nonce string
		salted                    []byte
	}
	authed := s.User == ""
	for {
		id, cmd, err := ReadRequest(c)
		if err != nil {
			return
		}
		name := ""
		if len(cmd) > 0 {
			name = cmd[0].Key
		}
		var reply D
		switch name {
		case "hello":
			reply = D{{"isWritablePrimary", true}, {"maxWireVersion", int32(21)}, {"ok", 1.0}}
		case "saslStart":
			p, _ := cmd.Get("payload").(Binary)
			msg := string(p.Data)
			bare := strings.TrimPrefix(msg, "n,,")
			fields := map[string]string{}
			for _, kv := range strings.Split(bare, ",") {
				if k, v, ok := strings.Cut(kv, "="); ok {
					fields[k] = v
				}
			}
			sc.first, sc.nonce = bare, fields["r"]+"srvnonce"
			salt := []byte("fixed-salt-16byt")
			sc.salted, _ = pbkdf2.Key(sha256.New, s.Password, salt, 4096, 32)
			sc.serverFirst = "r=" + sc.nonce + ",s=" + base64.StdEncoding.EncodeToString(salt) + ",i=4096"
			if fields["n"] != s.User {
				reply = D{{"ok", 0.0}, {"code", int32(18)}, {"errmsg", "Authentication failed."}}
			} else {
				reply = D{{"conversationId", int32(1)}, {"done", false}, {"payload", Binary{Data: []byte(sc.serverFirst)}}, {"ok", 1.0}}
			}
		case "saslContinue":
			p, _ := cmd.Get("payload").(Binary)
			msg := string(p.Data)
			if msg == "" {
				reply = D{{"conversationId", int32(1)}, {"done", true}, {"payload", Binary{}}, {"ok", 1.0}}
				break
			}
			without, proof64, _ := strings.Cut(msg, ",p=")
			authMsg := sc.first + "," + sc.serverFirst + "," + without
			clientKey := hmacSum(sc.salted, []byte("Client Key"))
			stored := sha256.Sum256(clientKey)
			sig := hmacSum(stored[:], []byte(authMsg))
			proof, _ := base64.StdEncoding.DecodeString(proof64)
			okProof := len(proof) == len(sig)
			for i := range sig {
				if okProof && proof[i]^sig[i] != clientKey[i] {
					okProof = false
				}
			}
			if !okProof {
				reply = D{{"ok", 0.0}, {"code", int32(18)}, {"errmsg", "Authentication failed."}}
				break
			}
			authed = true
			serverSig := hmacSum(hmacSum(sc.salted, []byte("Server Key")), []byte(authMsg))
			reply = D{{"conversationId", int32(1)}, {"done", false}, {"payload", Binary{Data: []byte("v=" + base64.StdEncoding.EncodeToString(serverSig))}}, {"ok", 1.0}}
		default:
			if !authed {
				reply = D{{"ok", 0.0}, {"code", int32(13)}, {"errmsg", "command requires authentication"}}
				break
			}
			s.mu.Lock()
			s.Commands = append(s.Commands, cmd)
			s.mu.Unlock()
			reply = s.Handle(cmd)
		}
		if err := WriteReply(c, id, reply); err != nil {
			return
		}
	}
}

func (s *FakeServer) Last() (D, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Commands) == 0 {
		return nil, errors.New("no commands")
	}
	return s.Commands[len(s.Commands)-1], nil
}

func (s *FakeServer) All() []D {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]D{}, s.Commands...)
}
