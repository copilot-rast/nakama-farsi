package server

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
)

func TestClusterLogoutBanDisconnectAndStreamJoinAcrossNodes(t *testing.T) {
	prefix := testPrefix()
	a := newClusterTestNode(t, "la", prefix)
	b := newClusterTestNode(t, "lb", prefix)
	waitCluster(t, a, b)
	time.Sleep(200 * time.Millisecond)
	user := uuid.Must(uuid.NewV4())
	exp := time.Now().Add(time.Hour).Unix()
	if !b.cache.IsValidSession(user, exp, "tok1") {
		t.Fatal("token must start valid")
	}
	a.cache.Remove(user, exp, "tok1", 0, "")
	clusterEventually(t, 2*time.Second, func() bool { return !b.cache.IsValidSession(user, exp, "tok1") }, "logout on a did not reach b")
	if !b.cache.IsValidSession(user, exp, "tok2") {
		t.Fatal("another token of the same user must stay valid")
	}
	other := uuid.Must(uuid.NewV4())
	a.cache.Ban([]uuid.UUID{other})
	clusterEventually(t, 2*time.Second, func() bool { return !b.cache.IsValidSession(other, time.Now().Add(-time.Second).Unix()+3600, "x") }, "ban on a did not reach b")

	late := newClusterTestNode(t, "lc", prefix)
	if late.cache.IsValidSession(user, exp, "tok1") {
		t.Fatal("a node started later must replay earlier logouts")
	}

	s := b.connect(user, "u", SessionFormatJson)
	stream := PresenceStream{Mode: 123, Label: "custom"}
	if ok, isNew, err := a.streams.UserJoin(stream, user, s.id, false, false, ""); err != nil || !ok || !isNew {
		t.Fatalf("remote stream join failed: %v %v %v", ok, isNew, err)
	}
	clusterEventually(t, 2*time.Second, func() bool { return a.tracker.CountByStream(stream) == 1 }, "remote join not visible")
	a.router.SendToStream(nil, stream, &rtapi.Envelope{Message: &rtapi.Envelope_StreamData{StreamData: &rtapi.StreamData{Data: "x"}}}, true)
	clusterEventually(t, 2*time.Second, func() bool { return s.count(isStreamData) == 1 }, "stream data not delivered to the remote member")
	if err := a.streams.UserLeave(stream, user, s.id); err != nil {
		t.Fatal(err)
	}
	clusterEventually(t, 2*time.Second, func() bool { return a.tracker.CountByStream(stream) == 0 }, "remote leave not visible")

	if err := a.sessions.Disconnect(context.Background(), s.id, true); err != nil {
		t.Fatal(err)
	}
	clusterEventually(t, 2*time.Second, func() bool {
		return s.count(func(e *rtapi.Envelope) bool {
			n := e.GetNotifications()
			return n != nil && len(n.Notifications) == 1 && n.Notifications[0].Code == NotificationCodeUserBanned
		}) == 1
	}, "ban disconnect did not reach the session on b")
}
