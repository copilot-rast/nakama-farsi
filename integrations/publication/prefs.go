package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const notifyCategoryJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'canOptOut',t.can_opt_out,'defaultOn',t.default_on,
	'isActive',t.is_active,'sort',t.sort,'optedOut',(SELECT count(*) FROM publication_notify_pref p WHERE p.category_id=t.id AND NOT p.enabled),
	'optedIn',(SELECT count(*) FROM publication_notify_pref p WHERE p.category_id=t.id AND p.enabled),
	'campaigns',(SELECT count(*) FROM publication_notification n WHERE n.category_id=t.id),'rules',(SELECT count(*) FROM publication_auto_rule r WHERE r.category_id=t.id),
	'createdAt',t.created_at,'updatedAt',t.updated_at)`

func notifyCategoryResource() resource {
	return resource{
		name:    "notify-categories",
		table:   "publication_notify_category",
		json:    notifyCategoryJSON,
		from:    ` FROM publication_notify_category t`,
		search:  []string{"t.key", translatedSearch("t.name")},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "key": "t.key = ?"},
		order:   "t.sort, t.key, t.id",
		save:    saveNotifyCategory,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			var used bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_notification WHERE category_id=$1) OR EXISTS (SELECT 1 FROM publication_auto_rule WHERE category_id=$1)`, id).Scan(&used); err != nil {
				return err
			}
			if used {
				return conflict("IN_USE", "campaigns or automatic rules use this category; switch it off instead")
			}
			return nil
		},
	}
}

func saveNotifyCategory(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string `json:"key"`
		Name        i18n   `json:"name"`
		Description i18n   `json:"description"`
		CanOptOut   *bool  `json:"canOptOut"`
		DefaultOn   *bool  `json:"defaultOn"`
		IsActive    *bool  `json:"isActive"`
		Sort        int    `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanBannerKey(ctx, tx, "publication_notify_category", id, create, in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 80)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 300)
	if err != nil {
		return err
	}
	canOut, defOn := boolOr(in.CanOptOut, true), boolOr(in.DefaultOn, true)
	if !canOut && !defOn {
		return badRequest("CATEGORY_ALWAYS_ON", "defaultOn", "a category players cannot turn off must be on by default")
	}
	return execSave(ctx, tx, "publication_notify_category", id, create, []string{"key", "name", "description", "can_opt_out", "default_on", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), canOut, defOn, boolOr(in.IsActive, true), in.Sort})
}

func categoryAllowsSQL(userCol, catParam string) string {
	return `(SELECT CASE WHEN NOT c.can_opt_out THEN true ELSE COALESCE((SELECT p.enabled FROM publication_notify_pref p WHERE p.user_id=` + userCol + ` AND p.category_id=c.id), c.default_on) END
		FROM publication_notify_category c WHERE c.id=` + catParam + `)`
}

func (s *service) optedOut(ctx context.Context, categoryID string, ids []string) (map[string]bool, error) {
	out := map[string]bool{}
	if categoryID == "" || len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT u::text FROM unnest($2::uuid[]) u WHERE NOT `+categoryAllowsSQL("u", "$1::uuid"), categoryID, uuidArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (s *service) prefsView(ctx context.Context, userID, requested string) (map[string]any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, requested)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT c.key, c.name::text, c.description::text, c.can_opt_out, CASE WHEN NOT c.can_opt_out THEN true ELSE COALESCE(p.enabled, c.default_on) END
		FROM publication_notify_category c LEFT JOIN publication_notify_pref p ON p.category_id=c.id AND p.user_id=$1 WHERE c.is_active ORDER BY c.sort, c.key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var key, name, desc string
		var canOut, on bool
		if err := rows.Scan(&key, &name, &desc, &canOut, &on); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"key": key, "name": langs.Resolve(parseI18n(name), lang), "description": langs.Resolve(parseI18n(desc), lang), "canOptOut": canOut, "enabled": on})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var devices int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM publication_push_device WHERE user_id=$1 AND is_active`, userID).Scan(&devices); err != nil {
		return nil, err
	}
	return map[string]any{"lang": lang, "items": items, "pushDevices": devices}, nil
}

func (s *service) rpcNotifyPrefs(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.prefsView(ctx, userID, in.Lang)
}

func (s *service) rpcNotifyPrefsSet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Prefs map[string]bool `json:"prefs"`
		Lang  string          `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if len(in.Prefs) == 0 {
		return nil, badRequest("PREFS_REQUIRED", "prefs", "send at least one category")
	}
	if len(in.Prefs) > 50 {
		return nil, badRequest("TOO_MANY_PREFS", "prefs", "at most 50 categories")
	}
	if err := s.bannerLimit(ctx, "push:p:"+userID, 30); err != nil {
		return nil, err
	}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		for key, on := range in.Prefs {
			var id string
			var canOut bool
			err := tx.QueryRowContext(ctx, `SELECT id::text, can_opt_out FROM publication_notify_category WHERE key=$1 AND is_active`, strings.TrimSpace(key)).Scan(&id, &canOut)
			if errors.Is(err, sql.ErrNoRows) {
				return badRequest("CATEGORY_NOT_FOUND", "prefs."+key, "unknown category")
			}
			if err != nil {
				return err
			}
			if !canOut && !on {
				return badRequest("CATEGORY_LOCKED", "prefs."+key, "this category cannot be turned off")
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO publication_notify_pref (user_id, category_id, enabled) VALUES ($1,$2,$3)
				ON CONFLICT (user_id, category_id) DO UPDATE SET enabled=EXCLUDED.enabled, updated_at=now()`, userID, id, on); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.prefsView(ctx, userID, in.Lang)
}

func (s *service) rpcNotificationOpened(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		NotificationID string `json:"notificationId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	id, err := cleanUUID("notificationId", in.NotificationID, true)
	if err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "push:o:"+userID, 60); err != nil {
		return nil, err
	}
	var counted bool
	err = s.db.QueryRowContext(ctx, `WITH r AS (UPDATE publication_notification_recipient SET opened_at=now(), seen_at=COALESCE(seen_at, now())
			WHERE notification_id=$1 AND user_id=$2 AND status='sent' AND opened_at IS NULL RETURNING notification_id),
		n AS (UPDATE publication_notification SET opened_count=opened_count+1 WHERE id IN (SELECT notification_id FROM r) RETURNING 1)
		SELECT EXISTS (SELECT 1 FROM n)`, *id, userID).Scan(&counted)
	if err != nil {
		return nil, err
	}
	if !counted {
		var exists bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_notification_recipient WHERE notification_id=$1 AND user_id=$2 AND status='sent')`, *id, userID).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, notFound("NOTIFICATION_NOT_FOUND")
		}
	}
	unread, err := s.unreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"counted": counted, "unread": unread}, nil
}
