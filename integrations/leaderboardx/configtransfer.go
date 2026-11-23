package main

import (
	"context"
	"errors"

	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/configtransfer"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) transferAreas() []configtransfer.Area {
	areas := configtransfer.LeaderboardAreas()
	for i := range areas {
		areas[i].After = s.afterTransfer
	}
	return areas
}

func (s *service) afterTransfer(ctx context.Context, res configtransfer.AreaResult) []string {
	invalidateBoards()
	settings.Invalidate()
	var warns []string
	for _, r := range res.Rows {
		if r.Table != "leaderboardx_board" || (r.Status != configtransfer.StatusNew && r.Status != configtransfer.StatusChanged) {
			continue
		}
		b, err := s.boardByKey(ctx, s.db, r.Key)
		if err != nil {
			warns = append(warns, "core_not_synced:"+r.Key)
			continue
		}
		if err := s.applyCore(ctx, b, false, false, "config-transfer"); err != nil {
			var ce *pluginkit.ConsoleError
			if errors.As(err, &ce) && ce.Code != "" {
				warns = append(warns, "core_"+ce.Code+":"+r.Key)
			} else {
				warns = append(warns, "core_not_synced:"+r.Key)
			}
			if s.logger != nil {
				s.logger.Warn("Config transfer: leaderboard %s core not synced: %v", r.Key, err)
			}
		}
	}
	return warns
}
