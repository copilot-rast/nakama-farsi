package main

import (
	"context"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/eventcontext"
)

const (
	SourceStage   = "campaign.stage"
	SourceChapter = "campaign.chapter"
	SourceItem    = "campaign.item"
	SourceAttempt = "campaign.attempt"

	maxContextSpans = 500
)

var contextLabels = map[string][2]string{
	SourceStage:   {"کلید مرحلهٔ تلاش باز", "Stage key of the open attempt"},
	SourceChapter: {"کلید فصل تلاش باز", "Chapter key of the open attempt"},
	SourceItem:    {"آیتم تلاش باز (یا آیتم مجهز)", "Item of the open attempt (or the equipped item)"},
	SourceAttempt: {"شناسهٔ تلاش باز", "Open attempt id"},
}

func contextSources() []eventcontext.Source {
	out := []eventcontext.Source{}
	for _, k := range []string{SourceStage, SourceChapter, SourceItem, SourceAttempt} {
		l := contextLabels[k]
		out = append(out, eventcontext.Source{Key: k, Owner: "campaign", LabelFa: l[0], LabelEn: l[1]})
	}
	for _, k := range campaignhooks.AttemptItemKeys() {
		out = append(out, eventcontext.Source{Key: SourceItem + "." + k, Owner: "campaign", LabelFa: "آیتم تلاش باز: " + k, LabelEn: "Open attempt item: " + k})
	}
	return out
}

func (s *service) contextSpans(ctx context.Context, userID string, from, to time.Time) ([]eventcontext.Span, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.id::text, a.started_at, LEAST(a.completed_at, a.closed_at, a.expires_at), s.key, COALESCE(c.key,''), a.item_key
		FROM campaign_attempt a JOIN campaign_stage s ON s.id=a.stage_id LEFT JOIN campaign_chapter c ON c.id=s.chapter_id
		WHERE a.user_id=$1 AND a.started_at <= $3 AND LEAST(a.completed_at, a.closed_at, a.expires_at) >= $2
		ORDER BY a.started_at DESC LIMIT $4`, userID, from, to, maxContextSpans)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []eventcontext.Span{}
	for rows.Next() {
		var sp eventcontext.Span
		var stage, chapter, item string
		if err := rows.Scan(&sp.ID, &sp.From, &sp.To, &stage, &chapter, &item); err != nil {
			return nil, err
		}
		sp.Values = map[string]string{SourceStage: stage, SourceChapter: chapter, SourceAttempt: sp.ID, SourceItem: item}
		out = append(out, sp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	cache := map[string]campaignhooks.ItemInfo{}
	for i := range out {
		key := out[i].Values[SourceItem]
		info, ok := cache[key]
		if !ok {
			info, err = campaignhooks.ResolveAttemptItem(ctx, userID, key)
			if err != nil {
				s.logger.Debug("Campaign attempt item for %s not resolved: %v", userID, err)
			}
			cache[key] = info
		}
		out[i].Values[SourceItem] = info.Key
		for k, v := range info.Values {
			out[i].Values[SourceItem+"."+k] = v
		}
	}
	return out, nil
}
