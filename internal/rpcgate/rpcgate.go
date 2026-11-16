package rpcgate

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
)

type Gate func(ctx context.Context, rpcID, userID string) (codes.Code, error)

var (
	mu   sync.RWMutex
	gate Gate
)

func Set(g Gate) {
	mu.Lock()
	gate = g
	mu.Unlock()
}

func Active() bool {
	mu.RLock()
	defer mu.RUnlock()
	return gate != nil
}

func Check(ctx context.Context, rpcID, userID string) (codes.Code, error) {
	mu.RLock()
	g := gate
	mu.RUnlock()
	if g == nil || userID == "" {
		return codes.OK, nil
	}
	return g(ctx, rpcID, userID)
}
