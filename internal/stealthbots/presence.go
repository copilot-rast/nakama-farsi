package stealthbots

import "github.com/heroiclabs/nakama-common/runtime"

type virtualPresences interface {
	VirtualPresenceJoin(userID, username string) error
	VirtualPresenceLeave(userID string) error
}

func ShowPresence(d runtime.MatchDispatcher, userID, username string) bool {
	v, ok := d.(virtualPresences)
	if !ok || userID == "" {
		return false
	}
	return v.VirtualPresenceJoin(userID, username) == nil
}

func HidePresence(d runtime.MatchDispatcher, userID string) bool {
	v, ok := d.(virtualPresences)
	if !ok || userID == "" {
		return false
	}
	return v.VirtualPresenceLeave(userID) == nil
}
