package server

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/rtapi"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type streamRouter struct {
	testMessageRouter
	sent []*rtapi.Envelope
}

func (s *streamRouter) SendToStream(_ *zap.Logger, _ PresenceStream, env *rtapi.Envelope, _ bool) {
	s.sent = append(s.sent, env)
}

func TestVirtualPresenceList(t *testing.T) {
	l := NewMatchPresenceList()
	human := &MatchPresence{Node: "n", UserID: uuid.Must(uuid.NewV4()), SessionID: uuid.Must(uuid.NewV4()), Username: "h"}
	l.Join([]*MatchPresence{human})
	if len(l.ListVisible()) != 1 || l.ListVisible()[0] != human {
		t.Fatal("no virtual: visible is the real list")
	}
	bot := &MatchPresence{Node: "n", UserID: uuid.Must(uuid.NewV4()), SessionID: uuid.Must(uuid.NewV4()), Username: "b"}
	if !l.JoinVirtual(bot) || l.JoinVirtual(bot) {
		t.Fatal("join once")
	}
	if l.Size() != 1 || len(l.ListPresences()) != 1 || len(l.ListPresenceIDs()) != 1 || len(l.ListVisible()) != 2 {
		t.Fatal("virtual presences never route or count")
	}
	if l.Contains(&PresenceID{Node: "n", SessionID: bot.SessionID}) {
		t.Fatal("not a session")
	}
	if l.LeaveVirtual(bot.UserID) == nil || l.LeaveVirtual(bot.UserID) != nil || len(l.ListVisible()) != 1 {
		t.Fatal("leave once")
	}
}

func TestRuntimeVirtualPresenceEvents(t *testing.T) {
	router := &streamRouter{}
	id := uuid.Must(uuid.NewV4())
	core := &RuntimeGoMatchCore{logger: zap.NewNop(), router: router, presenceList: NewMatchPresenceList(), id: id, node: "n", idStr: id.String() + ".n",
		stream: PresenceStream{Mode: StreamModeMatchAuthoritative, Subject: id, Label: "n"}, stopped: atomic.NewBool(false)}
	bot := uuid.Must(uuid.NewV4()).String()
	if err := core.VirtualPresenceJoin("x", "b"); err == nil {
		t.Fatal("bad id")
	}
	if err := core.VirtualPresenceJoin(bot, "sina_1392"); err != nil {
		t.Fatal(err)
	}
	_ = core.VirtualPresenceJoin(bot, "sina_1392")
	if len(router.sent) != 1 {
		t.Fatal("one join event")
	}
	ev := router.sent[0].GetMatchPresenceEvent()
	if ev == nil || ev.MatchId != id.String()+".n" || len(ev.Joins) != 1 || ev.Joins[0].UserId != bot || ev.Joins[0].Username != "sina_1392" || ev.Joins[0].SessionId == "" {
		t.Fatal(router.sent[0])
	}
	vis := core.presenceList.ListVisible()
	if len(vis) != 1 || vis[0].SessionID.String() != ev.Joins[0].SessionId {
		t.Fatal(vis)
	}
	if err := core.VirtualPresenceLeave(bot); err != nil {
		t.Fatal(err)
	}
	if len(router.sent) != 2 || len(router.sent[1].GetMatchPresenceEvent().GetLeaves()) != 1 {
		t.Fatal(router.sent)
	}
	core.stopped.Store(true)
	if err := core.VirtualPresenceJoin(bot, "x"); err == nil {
		t.Fatal("stopped match")
	}
}
