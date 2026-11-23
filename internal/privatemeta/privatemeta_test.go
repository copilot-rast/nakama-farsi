package privatemeta

import (
	"encoding/json"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
)

func TestCleanDropsPrivateKeysAndKeepsTheRest(t *testing.T) {
	got := Clean(`{"phone":"+989121234567","Mobile":"0912","email":"a@b.c","phone_number":1,"country":"IR","avatar":{"x":1}}`)
	var md map[string]any
	if err := json.Unmarshal([]byte(got), &md); err != nil {
		t.Fatal(err)
	}
	if len(md) != 2 || md["country"] != "IR" || md["avatar"] == nil {
		t.Fatalf("%s", got)
	}
	for _, raw := range []string{"", "{}", `{"country":"IR"}`, "not json", "[1,2]"} {
		if Clean(raw) != raw {
			t.Fatalf("%q changed to %q", raw, Clean(raw))
		}
	}
}

func TestStripsUsersFriendsAccountsAndMaps(t *testing.T) {
	u := &api.User{Id: "a", Metadata: `{"phone":"+989121234567","bot":true}`}
	f := &api.Friend{User: &api.User{Metadata: `{"phone":"x"}`}}
	g := &api.GroupUserList_GroupUser{User: &api.User{Metadata: `{"email":"x"}`}}
	a := &api.Account{User: &api.User{Metadata: `{"phoneNumber":"x"}`}}
	Users([]*api.User{u, nil})
	Friends([]*api.Friend{f, nil, {}})
	GroupUsers([]*api.GroupUserList_GroupUser{g, nil})
	Account(a)
	Account(nil)
	if u.Metadata != `{"bot":true}` || f.User.Metadata != "{}" || g.User.Metadata != "{}" || a.User.Metadata != "{}" {
		t.Fatal(u.Metadata, f.User.Metadata, g.User.Metadata, a.User.Metadata)
	}
	m := Map(map[string]any{"phone": "x", "level": 3})
	if len(m) != 1 || m["level"] != 3 {
		t.Fatal(m)
	}
}
