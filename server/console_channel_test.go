package server

import (
	"context"
	"testing"

	"github.com/heroiclabs/nakama/v3/console"
	"go.uber.org/zap"
)

func TestListChannelMessagesWithoutTargetIsEmpty(t *testing.T) {
	s := &ConsoleServer{logger: zap.NewNop()}
	for _, in := range []*console.ListChannelMessagesRequest{
		{Type: console.ListChannelMessagesRequest_ROOM},
		{Type: console.ListChannelMessagesRequest_GROUP},
		{Type: console.ListChannelMessagesRequest_DIRECT},
	} {
		list, err := s.ListChannelMessages(context.Background(), in)
		if err != nil {
			t.Fatalf("type %v: %v", in.Type, err)
		}
		if len(list.Messages) != 0 {
			t.Fatalf("type %v: expected no messages", in.Type)
		}
	}
	if _, err := s.ListChannelMessages(context.Background(), &console.ListChannelMessagesRequest{Type: console.ListChannelMessagesRequest_GROUP, GroupId: "bad"}); err == nil {
		t.Fatal("expected an error for a malformed group id")
	}
}
