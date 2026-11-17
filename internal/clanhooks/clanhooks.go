package clanhooks

import (
	"context"
	"sync"
)

type Membership struct {
	ClanID string
	Role   string
}

type Rival struct {
	WarID   string
	RoundID string
	Mode    string
	ClanID  string
	RivalID string
}

type (
	MemberReader func(ctx context.Context, userID string) (Membership, bool)
	MembersOf    func(ctx context.Context, clanID string) []string
	RivalCheck   func(ctx context.Context, userID, otherID string) (Rival, bool)
)

var (
	mu      sync.RWMutex
	member  MemberReader
	members MembersOf
	rival   RivalCheck
)

func Set(m MemberReader, all MembersOf, r RivalCheck) {
	mu.Lock()
	defer mu.Unlock()
	member, members, rival = m, all, r
}

func Available() bool {
	mu.RLock()
	defer mu.RUnlock()
	return member != nil
}

func ClanOf(ctx context.Context, userID string) (Membership, bool) {
	mu.RLock()
	fn := member
	mu.RUnlock()
	if fn == nil || userID == "" {
		return Membership{}, false
	}
	return fn(ctx, userID)
}

func Members(ctx context.Context, clanID string) []string {
	mu.RLock()
	fn := members
	mu.RUnlock()
	if fn == nil || clanID == "" {
		return nil
	}
	return fn(ctx, clanID)
}

func WarRival(ctx context.Context, userID, otherID string) (Rival, bool) {
	mu.RLock()
	fn := rival
	mu.RUnlock()
	if fn == nil || userID == "" || otherID == "" || userID == otherID {
		return Rival{}, false
	}
	return fn(ctx, userID, otherID)
}
