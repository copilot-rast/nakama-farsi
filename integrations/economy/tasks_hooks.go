package main

import (
	"context"

	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func asRefusal(err error) error {
	if app, ok := asApp(err); ok {
		return &economyhooks.Refusal{Status: app.Status, Code: app.Code, Message: app.Message, Meta: app.Meta}
	}
	return err
}

func (s *service) listTasksHook(ctx context.Context, userID string, f economyhooks.TaskFilter) ([]map[string]any, error) {
	if _, err := s.refreshGroups(ctx, userID, nil, ""); err != nil {
		return nil, asRefusal(err)
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, f.Lang)
	if err != nil {
		return nil, asRefusal(err)
	}
	v, err := s.newTaskViewer(ctx, userID, langs, lang, false)
	if err != nil {
		return nil, asRefusal(err)
	}
	keys := map[string]bool{}
	for _, k := range f.Keys {
		keys[k] = true
	}
	top := []*taskDef{}
	for _, t := range v.e.tasks {
		if t == nil || t.ParentID != "" || !t.Active {
			continue
		}
		if f.Category != "" && v.cats[t.CategoryID] != f.Category {
			continue
		}
		if len(keys) > 0 && !keys[t.Key] {
			continue
		}
		top = append(top, t)
	}
	sortTasks(top)
	out := []map[string]any{}
	for _, t := range top {
		tv, err := v.view(ctx, t)
		if err != nil {
			return nil, asRefusal(err)
		}
		if tv["state"] == "hidden" {
			continue
		}
		tv["lang"] = lang
		tv["startsAt"], tv["endsAt"] = timeOrNil(t.StartsAt), timeOrNil(t.EndsAt)
		if t.PrizeID != "" {
			tv["prizeId"] = t.PrizeID
		}
		out = append(out, tv)
	}
	return out, nil
}

func (s *service) claimTaskHook(ctx context.Context, userID, idOrKey, lang string) (map[string]any, error) {
	t, err := s.findTask(ctx, idOrKey)
	if err != nil {
		return nil, asRefusal(err)
	}
	res, err := s.claimTask(ctx, userID, t.ID, "", "player")
	if err != nil {
		return nil, asRefusal(err)
	}
	out, err := s.taskClaimResponse(ctx, userID, lang, res)
	if err != nil {
		return nil, asRefusal(err)
	}
	return out, nil
}
