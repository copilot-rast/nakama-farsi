package server

import (
	"context"
	"sort"
	"time"

	"github.com/heroiclabs/nakama/v3/console"
	"google.golang.org/protobuf/proto"
)

type ClusterStatusHandler struct {
	c     *Cluster
	local StatusHandler
}

var _ StatusHandler = (*ClusterStatusHandler)(nil)

func (s *ClusterStatusHandler) GetStatus(ctx context.Context) ([]*console.StatusList_Status, error) {
	own, err := s.local.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	out := append([]*console.StatusList_Status{}, own...)
	for _, r := range s.c.fanout(callCtx, callNodeStatus, nil, false) {
		if r.err != nil {
			out = append(out, &console.StatusList_Status{Name: r.node, Health: console.StatusHealth_STATUS_HEALTH_ERROR})
			continue
		}
		list := &console.StatusList{}
		if err := proto.Unmarshal(r.out, list); err != nil {
			out = append(out, &console.StatusList_Status{Name: r.node, Health: console.StatusHealth_STATUS_HEALTH_ERROR})
			continue
		}
		out = append(out, list.Nodes...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *ClusterStatusHandler) handleCall(ctx context.Context) ([]byte, error) {
	own, err := s.local.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&console.StatusList{Nodes: own})
}
