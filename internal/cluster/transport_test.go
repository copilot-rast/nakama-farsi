package cluster

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

type recorder struct {
	mu     sync.Mutex
	frames map[string][]string
	resets map[string]int
}

func newRecorder() *recorder {
	return &recorder{frames: map[string][]string{}, resets: map[string]int{}}
}

func (r *recorder) HandleFrame(from string, kind byte, payload []byte) {
	r.mu.Lock()
	if kind == FrameResync {
		r.resets[from+"/resync"]++
		r.mu.Unlock()
		return
	}
	r.frames[from] = append(r.frames[from], fmt.Sprintf("%d:%s", kind, payload))
	r.mu.Unlock()
}

func (r *recorder) HandleCall(_ context.Context, from string, kind byte, payload []byte) ([]byte, error) {
	if kind == 9 {
		return nil, errors.New("boom")
	}
	return []byte(fmt.Sprintf("%s->%d:%s", from, kind, payload)), nil
}

func (r *recorder) PeerReset(node string) {
	r.mu.Lock()
	r.resets[node]++
	r.mu.Unlock()
}

func (r *recorder) count(from string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.frames[from])
}

func TestFrameRoundTrip(t *testing.T) {
	e := NewEnc(16).U8(7).Bool(true).Uvarint(300).Varint(-5).Str("سلام").Bytes([]byte{1, 2}).Raw16([16]byte{9})
	batch := AppendFrame(nil, 3, e.Out())
	batch = AppendFrame(batch, 4, []byte("x"))
	var kinds []byte
	var first []byte
	if err := SplitFrames(batch, func(k byte, p []byte) {
		kinds = append(kinds, k)
		if first == nil {
			first = p
		}
	}); err != nil {
		t.Fatal(err)
	}
	if len(kinds) != 2 || kinds[0] != 3 || kinds[1] != 4 {
		t.Fatalf("kinds %v", kinds)
	}
	d := NewDec(first)
	if d.U8() != 7 || !d.Bool() || d.Uvarint() != 300 || d.Varint() != -5 || d.Str() != "سلام" || len(d.Bytes()) != 2 || d.Raw16()[0] != 9 || d.Err() != nil || d.Len() != 0 {
		t.Fatal("decode mismatch")
	}
	if d.U8(); d.Err() == nil {
		t.Fatal("expected short frame")
	}
	if err := SplitFrames([]byte{1, 5, 0}, func(byte, []byte) {}); err == nil {
		t.Fatal("expected error for truncated batch")
	}
}

func startPair(t *testing.T, secretB string) (*Transport, *recorder, *Transport, *recorder) {
	t.Helper()
	ra, rb := newRecorder(), newRecorder()
	a := NewTransport(zap.NewNop(), "a", "s3cret", 1024, time.Second, ra)
	b := NewTransport(zap.NewNop(), "b", secretB, 1024, time.Second, rb)
	if err := a.Listen("127.0.0.1", 0); err != nil {
		t.Fatal(err)
	}
	if err := b.Listen("127.0.0.1", 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Stop(); b.Stop() })
	if err := a.AddPeer("b", fmt.Sprintf("127.0.0.1:%d", b.Port())); err != nil {
		t.Fatal(err)
	}
	if err := b.AddPeer("a", fmt.Sprintf("127.0.0.1:%d", a.Port())); err != nil {
		t.Fatal(err)
	}
	return a, ra, b, rb
}

func TestTransportOrderedFramesAndCalls(t *testing.T) {
	a, _, b, rb := startPair(t, "s3cret")
	const n = 5000
	for i := 0; i < n; i++ {
		if !a.Send("b", 2, []byte(fmt.Sprint(i))) {
			i--
			time.Sleep(time.Millisecond)
		}
	}
	eventually(t, 5*time.Second, func() bool { return rb.count("a") == n }, "frames missing")
	rb.mu.Lock()
	for i, f := range rb.frames["a"] {
		if f != fmt.Sprintf("2:%d", i) {
			rb.mu.Unlock()
			t.Fatalf("frame %d out of order: %s", i, f)
		}
	}
	rb.mu.Unlock()
	if st := a.Stats(); st.BatchesSent >= n {
		t.Fatalf("frames were not batched: %+v", st)
	}
	out, err := a.Call(context.Background(), "b", 5, []byte("hi"))
	if err != nil || string(out) != "a->5:hi" {
		t.Fatalf("call: %q %v", out, err)
	}
	if _, err := a.Call(context.Background(), "b", 9, nil); err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom, got %v", err)
	}
	if _, err := a.Call(context.Background(), "zz", 5, nil); !errors.Is(err, ErrUnknownPeer) {
		t.Fatalf("expected unknown peer, got %v", err)
	}
	b.Broadcast(3, []byte("all"))
	_ = b
}

func TestTransportRejectsWrongSecret(t *testing.T) {
	a, _, _, rb := startPair(t, "other")
	a.Send("b", 2, []byte("x"))
	if _, err := a.Call(context.Background(), "b", 5, nil); err == nil {
		t.Fatal("call with a wrong secret must fail")
	}
	time.Sleep(200 * time.Millisecond)
	if rb.count("a") != 0 {
		t.Fatal("frames with a wrong secret must be refused")
	}
}

func TestTransportReconnectSignalsReset(t *testing.T) {
	ra := newRecorder()
	rb := newRecorder()
	a := NewTransport(zap.NewNop(), "a", "k", 1024, time.Second, ra)
	defer a.Stop()
	b := NewTransport(zap.NewNop(), "b", "k", 1024, time.Second, rb)
	if err := b.Listen("127.0.0.1", 0); err != nil {
		t.Fatal(err)
	}
	port := b.Port()
	if err := a.AddPeer("b", fmt.Sprintf("127.0.0.1:%d", port)); err != nil {
		t.Fatal(err)
	}
	a.Send("b", 2, []byte("one"))
	eventually(t, 3*time.Second, func() bool { return rb.count("a") == 1 }, "first frame not received")
	b.Stop()
	time.Sleep(100 * time.Millisecond)
	b2 := NewTransport(zap.NewNop(), "b", "k", 1024, time.Second, rb)
	if err := b2.Listen("127.0.0.1", port); err != nil {
		t.Fatal(err)
	}
	defer b2.Stop()
	eventually(t, 10*time.Second, func() bool {
		a.Send("b", 2, []byte("again"))
		time.Sleep(50 * time.Millisecond)
		return rb.count("a") >= 2
	}, "frames not delivered after reconnect")
	ra.mu.Lock()
	resets := ra.resets["b"]
	ra.mu.Unlock()
	if resets == 0 {
		t.Fatal("sender must signal a reset after reconnecting")
	}
}

func TestTransportQueueFullDropsAndAsksResync(t *testing.T) {
	r := newRecorder()
	a := NewTransport(zap.NewNop(), "a", "k", 4, time.Second, r)
	defer a.Stop()
	if err := a.AddPeer("b", "127.0.0.1:1"); err != nil {
		t.Fatal(err)
	}
	sent := 0
	for i := 0; i < 50; i++ {
		if a.Send("b", 2, []byte("x")) {
			sent++
		}
	}
	if sent > 5 || a.Stats().FramesDropped == 0 {
		t.Fatalf("expected drops, sent=%d stats=%+v", sent, a.Stats())
	}
}
