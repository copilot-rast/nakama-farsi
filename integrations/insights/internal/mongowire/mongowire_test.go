package mongowire

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBSONRoundTrip(t *testing.T) {
	at := time.UnixMilli(1793500000123).UTC()
	v := 1.5
	doc := D{{"s", "سلام"}, {"f", 2.5}, {"p", &v}, {"n", nil}, {"b", true}, {"i", int32(7)}, {"l", int64(1) << 40}, {"t", at},
		{"d", D{{"x", "y"}}}, {"a", A{"a", int32(1), D{{"z", false}}}}, {"m", map[string]string{"car": "c1", "a": "b"}}, {"bin", Binary{Data: []byte{1, 2}}}}
	raw, err := Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Unmarshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if back.Get("s") != "سلام" || back.Get("f") != 2.5 || back.Get("p") != 1.5 || back.Get("n") != nil || back.Get("b") != true || back.Get("i") != int32(7) || back.Get("l") != int64(1)<<40 {
		t.Fatalf("scalars %+v", back)
	}
	if !back.Get("t").(time.Time).Equal(at) || back.Get("d").(D).Get("x") != "y" || back.Get("a").(A)[2].(D).Get("z") != false {
		t.Fatalf("nested %+v", back)
	}
	m := back.Get("m").(D)
	if m[0].Key != "a" || m.Get("car") != "c1" || back.Get("bin").(Binary).Data[1] != 2 {
		t.Fatalf("map/binary %+v", back)
	}
	if _, err := Unmarshal(raw[:len(raw)-2]); err == nil {
		t.Fatal("truncated accepted")
	}
	if _, err := Marshal(D{{"bad\x00key", 1}}); err == nil {
		t.Fatal("nul key accepted")
	}
}

func TestScramRFC7677Vector(t *testing.T) {
	s := newScram("user", "pencil", "rOprNGfwEbeRWgbNEkqO")
	if s.clientFirst() != "n,,n=user,r=rOprNGfwEbeRWgbNEkqO" {
		t.Fatal(s.clientFirst())
	}
	final, err := s.clientFinal("r=rOprNGfwEbeRWgbNEkqO%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,s=W22ZaJ0SNY7soEsUEjb6gQ==,i=4096")
	if err != nil {
		t.Fatal(err)
	}
	if final != "c=biws,r=rOprNGfwEbeRWgbNEkqO%hvYDpWUa2RaTCAfuxFIlj)hNlF$k0,p=dHzbZapWIk4jUhN+Ute9ytag9zjfMHgsqmmiz7AndVQ=" {
		t.Fatalf("final %s", final)
	}
	if err := s.verify("v=6rriTRBi23WpRR/wtup+mMhUZUn/dB5nLTJRsjl95G4="); err != nil {
		t.Fatal(err)
	}
	if err := s.verify("v=AAAA"); err == nil {
		t.Fatal("bad signature accepted")
	}
	if _, err := newScram("u", "p", "abc").clientFinal("r=zzz,s=AAAA,i=4096"); err == nil {
		t.Fatal("foreign nonce accepted")
	}
}

func TestClientAuthInsertAggregateAgainstFake(t *testing.T) {
	srv, err := NewFakeServer("analytics", "s3cret", func(cmd D) D {
		switch cmd[0].Key {
		case "insert":
			docs := cmd.Get("documents").(A)
			if len(docs) == 3 {
				return D{{"n", int32(2)}, {"writeErrors", A{D{{"index", int32(2)}, {"code", int32(11000)}, {"errmsg", "E11000 duplicate key"}}}}, {"ok", 1.0}}
			}
			return D{{"n", int32(len(docs))}, {"ok", 1.0}}
		case "aggregate":
			return D{{"cursor", D{{"id", int64(5)}, {"ns", "db.c"}, {"firstBatch", A{D{{"_id", "a"}, {"n", int32(1)}}}}}}, {"ok", 1.0}}
		case "getMore":
			return D{{"cursor", D{{"id", int64(0)}, {"ns", "db.c"}, {"nextBatch", A{D{{"_id", "b"}, {"n", int64(2)}}}}}}, {"ok", 1.0}}
		case "boom":
			return D{{"ok", 0.0}, {"code", int32(59)}, {"errmsg", "no such command"}}
		}
		return D{{"ok", 1.0}}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cl, _ := New(Options{Hosts: []string{"127.0.0.1:1", srv.Addr()}, User: "analytics", Password: "s3cret", Timeout: 2 * time.Second})
	defer cl.Close()
	ctx := context.Background()
	dups, err := cl.InsertMany(ctx, "db", "c", []D{{{"_id", "1"}}, {{"_id", "2"}}, {{"_id", "2"}}})
	if err != nil || dups != 1 {
		t.Fatalf("insert %d %v", dups, err)
	}
	last, _ := srv.Last()
	if last.Get("$db") != "db" || last.Get("ordered") != false {
		t.Fatalf("insert cmd %+v", last)
	}
	rows, err := cl.Aggregate(ctx, "db", "c", []D{{{"$match", D{}}}}, 0)
	if err != nil || len(rows) != 2 || rows[1].Get("_id") != "b" {
		t.Fatalf("aggregate %+v %v", rows, err)
	}
	if _, err := cl.Aggregate(ctx, "db", "c", nil, 1); err == nil {
		t.Fatal("max not enforced")
	}
	_, err = cl.Command(ctx, "db", D{{"boom", int32(1)}})
	var ce *CommandError
	if !errors.As(err, &ce) || ce.Code != 59 {
		t.Fatalf("command error %v", err)
	}
	bad, _ := New(Options{Hosts: []string{srv.Addr()}, User: "analytics", Password: "wrong", Timeout: 2 * time.Second})
	if _, err := bad.Command(ctx, "db", D{{"ping", int32(1)}}); err == nil {
		t.Fatal("wrong password accepted")
	}
	anon, _ := New(Options{Hosts: []string{srv.Addr()}, Timeout: 2 * time.Second})
	if _, err := anon.Command(ctx, "db", D{{"ping", int32(1)}}); err == nil {
		t.Fatal("unauthenticated command accepted")
	}
}
