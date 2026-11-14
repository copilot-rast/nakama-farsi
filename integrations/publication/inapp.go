package main

import (
	"context"

	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) inappSender() inapp.Sender {
	return func(ctx context.Context, m inapp.Message) (string, error) {
		channel := m.Channel
		if channel != "push_in_app" {
			channel = "in_app"
		}
		in := notificationInput{Channel: channel, Title: localized.Text(m.Title), Body: localized.Text(m.Body), Target: target{Mode: "users", UserIDs: m.UserIDs}}
		if m.Action != nil {
			in.Action = &action{Type: m.Action.Type, Value: m.Action.Value, Label: localized.Text(m.Action.Label)}
		}
		createdBy := m.CreatedBy
		if createdBy == "" {
			createdBy = "system"
		}
		return s.createNotification(ctx, in, createdBy)
	}
}
