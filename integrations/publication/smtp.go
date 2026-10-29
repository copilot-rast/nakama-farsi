package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

func buildMIME(from, to, subject, text, htmlBody string, now time.Time, domain string) []byte {
	var b bytes.Buffer
	boundaryBytes := make([]byte, 12)
	_, _ = rand.Read(boundaryBytes)
	boundary := "pub-" + hex.EncodeToString(boundaryBytes)
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	header := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	header("From", from)
	header("To", to)
	subject = strings.Join(strings.Fields(subject), " ")
	header("Subject", mime.BEncoding.Encode("UTF-8", subject))
	header("Date", now.Format(time.RFC1123Z))
	header("Message-ID", "<"+hex.EncodeToString(idBytes)+"@"+domain+">")
	header("MIME-Version", "1.0")
	header("Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	b.WriteString("\r\n")
	part := func(ctype, body string) {
		b.WriteString("--" + boundary + "\r\n")
		header("Content-Type", ctype+"; charset=UTF-8")
		header("Content-Transfer-Encoding", "base64")
		b.WriteString("\r\n")
		enc := base64.StdEncoding.EncodeToString([]byte(body))
		for len(enc) > 76 {
			b.WriteString(enc[:76] + "\r\n")
			enc = enc[76:]
		}
		b.WriteString(enc + "\r\n")
	}
	part("text/plain", text)
	part("text/html", htmlBody)
	b.WriteString("--" + boundary + "--\r\n")
	return b.Bytes()
}

func (s *service) tlsFor(host string) *tls.Config {
	if s.tls != nil {
		c := s.tls.Clone()
		if c.ServerName == "" {
			c.ServerName = host
		}
		return c
	}
	return &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
}

func smtpErr(stage string, err error) error {
	var tp *textproto.Error
	if errors.As(err, &tp) {
		msg := fmt.Sprintf("smtp %s: %d %s", stage, tp.Code, tp.Msg)
		if tp.Code >= 500 {
			return permanentErr("%s", msg)
		}
		return transientErr("%s", msg)
	}
	return transientErr("smtp %s: %v", stage, err)
}

func (s *service) sendSMTP(ctx context.Context, cfg providerConfig, msg outMessage) (string, error) {
	host := cfg.str("host")
	port := cfg.str("port")
	if port == "" {
		port = "587"
	}
	security := cfg.str("security")
	if security == "" {
		security = "starttls"
	}
	addr := net.JoinHostPort(host, port)
	deadline := time.Now().Add(30 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	var conn net.Conn
	var err error
	if security == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, s.tlsFor(host))
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return "", transientErr("smtp connect: %v", err)
	}
	_ = conn.SetDeadline(deadline)
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return "", smtpErr("greeting", err)
	}
	defer c.Close()
	if err := c.Hello("localhost"); err != nil {
		return "", smtpErr("EHLO", err)
	}
	if security == "starttls" {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return "", permanentErr("smtp server does not offer STARTTLS; pick security tls or none")
		}
		if err := c.StartTLS(s.tlsFor(host)); err != nil {
			return "", transientErr("smtp STARTTLS: %v", err)
		}
	}
	if user := cfg.str("username"); user != "" {
		if ok, _ := c.Extension("AUTH"); !ok {
			return "", permanentErr("smtp server does not offer AUTH")
		}
		if err := c.Auth(smtp.PlainAuth("", user, cfg.str("password"), host)); err != nil {
			if strings.Contains(err.Error(), "unencrypted") {
				return "", permanentErr("smtp refuses to send the password over an unencrypted connection; use starttls or tls")
			}
			return "", smtpErr("AUTH", err)
		}
	}
	from := cfg.str("from")
	if err := c.Mail(from); err != nil {
		return "", smtpErr("MAIL FROM", err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return "", smtpErr("RCPT TO", err)
	}
	wc, err := c.Data()
	if err != nil {
		return "", smtpErr("DATA", err)
	}
	domain := from[strings.LastIndex(from, "@")+1:]
	if _, err := wc.Write(buildMIME(fromHeader(cfg), msg.To, msg.Subject, msg.Text, msg.HTML, s.now(), domain)); err != nil {
		return "", smtpErr("DATA", err)
	}
	if err := wc.Close(); err != nil {
		return "", smtpErr("DATA", err)
	}
	_ = c.Quit()
	return "smtp accepted by " + host, nil
}
