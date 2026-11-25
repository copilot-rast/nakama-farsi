package privatemeta

import (
	"encoding/json"
	"strings"

	"github.com/heroiclabs/nakama-common/api"
)

var Keys = []string{"phone", "phoneNumber", "phone_number", "mobile", "email", "bot"}

func private(key string) bool {
	for _, k := range Keys {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

func Clean(raw string) string {
	if !strings.Contains(raw, "\"") {
		return raw
	}
	var md map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &md); err != nil {
		return raw
	}
	found := false
	for k := range md {
		if private(k) {
			delete(md, k)
			found = true
		}
	}
	if !found {
		return raw
	}
	out, err := json.Marshal(md)
	if err != nil {
		return "{}"
	}
	return string(out)
}

func Map(md map[string]any) map[string]any {
	for k := range md {
		if private(k) {
			delete(md, k)
		}
	}
	return md
}

func User(u *api.User) {
	if u != nil && u.Metadata != "" {
		u.Metadata = Clean(u.Metadata)
	}
}

func Users(users []*api.User) {
	for _, u := range users {
		User(u)
	}
}

func Friends(friends []*api.Friend) {
	for _, f := range friends {
		if f != nil {
			User(f.User)
		}
	}
}

func GroupUsers(users []*api.GroupUserList_GroupUser) {
	for _, g := range users {
		if g != nil {
			User(g.User)
		}
	}
}

func Account(a *api.Account) {
	if a != nil {
		User(a.User)
	}
}
