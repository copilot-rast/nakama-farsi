package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The notification centre and /settings endpoints of the main app (source
// notification-centre.service.ts and notification-settings.service.ts). Jazebe
// notices are Nakama notification rows with code 1001; their read and
// soft-delete state lives in jazebe_notification_state.

const jazebeNotificationCode = 1001

var notificationRewardSubjects = map[string]bool{"medal": true, "level": true, "reward": true, "lottery": true, "order": true}

var defaultCategoryPrefs = map[string]bool{"missions": true, "rewards": true, "friends": true, "news": true, "marketing": false}

type channelToggle struct {
	InApp bool `json:"in_app"`
	SMS   bool `json:"sms"`
}

func defaultChannelPrefs() map[string]channelToggle {
	return map[string]channelToggle{"group": {InApp: true}, "mission_game": {InApp: true}, "system": {InApp: true}}
}

// storedNotificationCategory normalizes a native notice category to the
// source NotificationCategory.
func storedNotificationCategory(content map[string]any) string {
	switch c, _ := content["category"].(string); c {
	case "critical", "system", "mission_game":
		return c
	case "group", "group_message":
		return "group"
	}
	return "system"
}

func centreCategory(content map[string]any) string {
	switch storedNotificationCategory(content) {
	case "critical":
		return "critical"
	case "group":
		return "group"
	case "mission_game":
		if subject, _ := content["subjectType"].(string); notificationRewardSubjects[subject] {
			return "reward"
		}
		return "game"
	}
	return "system"
}

// notificationPrefKey is the settings toggle a notice falls under; "" for
// critical, which no toggle mutes (source categoryPrefKey).
func notificationPrefKey(content map[string]any) string {
	switch storedNotificationCategory(content) {
	case "critical":
		return ""
	case "group":
		return "friends"
	case "mission_game":
		if subject, _ := content["subjectType"].(string); notificationRewardSubjects[subject] {
			return "rewards"
		}
		return "missions"
	}
	if source, _ := content["source"].(string); source == "brand" {
		return "marketing"
	}
	return "news"
}

// notificationDelivered applies the recipient's preferences (source
// resolveChannels for the in-app channel): critical always, otherwise both
// the category toggle and the category's in-app toggle must be on.
func notificationDelivered(ctx context.Context, tx *sql.Tx, userID string, content map[string]any) (bool, error) {
	key := notificationPrefKey(content)
	if key == "" {
		return true, nil
	}
	var categoryRaw, channelRaw []byte
	err := tx.QueryRowContext(ctx, `SELECT category_prefs,channel_prefs FROM jazebe_notification_preference WHERE user_id=$1`, userID).Scan(&categoryRaw, &channelRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultCategoryPrefs[key], nil
	}
	if err != nil {
		return false, err
	}
	categories := map[string]bool{}
	for k, v := range defaultCategoryPrefs {
		categories[k] = v
	}
	_ = json.Unmarshal(categoryRaw, &categories)
	if !categories[key] {
		return false, nil
	}
	channels := defaultChannelPrefs()
	_ = json.Unmarshal(channelRaw, &channels)
	return channels[storedNotificationCategory(content)].InApp, nil
}

type notificationScope struct {
	PlanetID string `json:"planetId"`
}

// inboxFilter narrows to the caller's undeleted Jazebe notices and, in
// brand-domain mode (a planetId is given), to that planet plus platform-wide.
const inboxFilter = `n.user_id=$1 AND n.code=1001 AND s.deleted_at IS NULL
	AND ($2::text='' OR n.content->>'planetId'=$2::text OR NOT (n.content ? 'planetId'))`

func notificationScopeFrom(payload string) (string, error) {
	var in notificationScope
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.PlanetID == "" {
		return "", nil
	}
	return requiredUUID(in.PlanetID)
}

func notificationUnread(ctx context.Context, db *sql.DB, user, planet string) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM notification n LEFT JOIN jazebe_notification_state s ON s.notification_id=n.id
		WHERE `+inboxFilter+` AND s.read_at IS NULL`, user, planet).Scan(&n)
	if err != nil {
		return 0, status.Error(codes.Internal, "notifications unavailable")
	}
	return n, nil
}

func notificationList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID string `json:"planetId"`
		Category string `json:"category"`
		Cursor   string `json:"cursor"`
		Limit    int    `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	planet := ""
	if in.PlanetID != "" {
		if planet, err = requiredUUID(in.PlanetID); err != nil {
			return "", err
		}
	}
	if in.Limit == 0 {
		in.Limit = 50
	}
	if in.Limit < 1 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "limit must be 1 to 100")
	}
	category := ""
	switch in.Category {
	case "":
	case "missions", "mission_game", "reward", "game":
		category = "mission_game"
	case "group", "system", "critical":
		category = in.Category
	default:
		return "", status.Error(codes.InvalidArgument, "invalid notification category")
	}
	var cursorAt any
	cursorID := ""
	if in.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(in.Cursor)
		parts := strings.SplitN(string(raw), "|", 2)
		if err != nil || len(parts) != 2 {
			return "", status.Error(codes.InvalidArgument, "invalid cursor")
		}
		at, err := time.Parse(time.RFC3339Nano, parts[0])
		if err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid cursor")
		}
		if cursorID, err = requiredUUID(parts[1]); err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid cursor")
		}
		cursorAt = at
	}
	rows, err := db.QueryContext(ctx, `SELECT n.id,n.subject,n.content,n.create_time,s.read_at IS NOT NULL,COALESCE(p.name,'')
		FROM notification n LEFT JOIN jazebe_notification_state s ON s.notification_id=n.id
		LEFT JOIN jazebe_planet p ON $2::text='' AND p.id::text=n.content->>'planetId'
		WHERE `+inboxFilter+`
		AND ($3::text='' OR CASE WHEN n.content->>'category' IN ('critical','system','mission_game') THEN n.content->>'category'
			WHEN n.content->>'category' IN ('group','group_message') THEN 'group' ELSE 'system' END = $3::text)
		AND ($4::timestamptz IS NULL OR (n.create_time,n.id) < ($4::timestamptz,$5::uuid))
		ORDER BY n.create_time DESC,n.id DESC LIMIT $6`, user, planet, category, cursorAt, nullUUID(cursorID), in.Limit+1)
	if err != nil {
		return "", status.Error(codes.Internal, "notifications unavailable")
	}
	defer rows.Close()
	now := time.Now().UTC()
	items := []map[string]any{}
	var lastAt time.Time
	var lastID string
	more := false
	for rows.Next() {
		var id, subject, planetName string
		var raw []byte
		var created time.Time
		var read bool
		if err := rows.Scan(&id, &subject, &raw, &created, &read, &planetName); err != nil {
			return "", status.Error(codes.Internal, "notifications unavailable")
		}
		if len(items) == in.Limit {
			more = true
			break
		}
		content := map[string]any{}
		_ = json.Unmarshal(raw, &content)
		body, _ := content["body"].(string)
		link, _ := content["actionLink"].(string)
		priority, _ := content["priority"].(string)
		if priority == "" {
			priority = "normal"
		}
		var planetLabel, linkValue any
		if planetName != "" {
			planetLabel = planetName
		}
		if link != "" {
			linkValue = link
		}
		items = append(items, map[string]any{"id": id, "cat": centreCategory(content), "title": subject, "body": body,
			"time": groupRelativeFa(created, now), "createdAt": created, "planet": planetLabel, "read": read,
			"action": link != "", "link": linkValue, "priority": priority})
		lastAt, lastID = created, id
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "notifications unavailable")
	}
	var next any
	if more {
		next = base64.RawURLEncoding.EncodeToString([]byte(lastAt.Format(time.RFC3339Nano) + "|" + lastID))
	}
	unread, err := notificationUnread(ctx, db, user, planet)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"items": items, "nextCursor": next, "unread": unread})
}

func nullUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}

func notificationUnreadCount(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	planet, err := notificationScopeFrom(payload)
	if err != nil {
		return "", err
	}
	n, err := notificationUnread(ctx, db, user, planet)
	if err != nil {
		return "", err
	}
	return encode(map[string]int{"unread": n})
}

// notificationMark sets read_at or deleted_at on the caller's notices.
func notificationMark(ctx context.Context, db *sql.DB, user, planet, id, column string) (int64, error) {
	res, err := db.ExecContext(ctx, `INSERT INTO jazebe_notification_state(notification_id,user_id,`+column+`)
		SELECT n.id,n.user_id,now() FROM notification n LEFT JOIN jazebe_notification_state s ON s.notification_id=n.id
		WHERE `+inboxFilter+` AND ($3::uuid IS NULL OR n.id=$3::uuid) AND s.`+column+` IS NULL
		ON CONFLICT (notification_id) DO UPDATE SET `+column+`=COALESCE(jazebe_notification_state.`+column+`,now()),updated_at=now()`,
		user, planet, nullUUID(id))
	if err != nil {
		return 0, status.Error(codes.Internal, "notification update failed")
	}
	return res.RowsAffected()
}

func notificationOwned(ctx context.Context, db *sql.DB, user, id string) (bool, error) {
	var ok bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM notification n LEFT JOIN jazebe_notification_state s ON s.notification_id=n.id
		WHERE n.id=$2 AND n.user_id=$1 AND n.code=1001 AND s.deleted_at IS NULL)`, user, id).Scan(&ok)
	if err != nil {
		return false, status.Error(codes.Internal, "notification unavailable")
	}
	return ok, nil
}

func notificationReadAll(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	planet, err := notificationScopeFrom(payload)
	if err != nil {
		return "", err
	}
	n, err := notificationMark(ctx, db, user, planet, "", "read_at")
	if err != nil {
		return "", err
	}
	return encode(map[string]int64{"updated": n})
}

func notificationIDFrom(payload string) (string, error) {
	var in struct {
		ID string `json:"id"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	return requiredUUID(in.ID)
}

func notificationMarkRead(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	id, err := notificationIDFrom(payload)
	if err != nil {
		return "", err
	}
	if ok, err := notificationOwned(ctx, db, user, id); err != nil || !ok {
		if err == nil {
			err = status.Error(codes.NotFound, "notification not found")
		}
		return "", err
	}
	if _, err := notificationMark(ctx, db, user, "", id, "read_at"); err != nil {
		return "", err
	}
	return encode(map[string]any{"id": id, "read": true})
}

// notificationRemove and notificationClear soft-delete, as the source does.
func notificationRemove(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	id, err := notificationIDFrom(payload)
	if err != nil {
		return "", err
	}
	if ok, err := notificationOwned(ctx, db, user, id); err != nil || !ok {
		if err == nil {
			err = status.Error(codes.NotFound, "notification not found")
		}
		return "", err
	}
	if _, err := notificationMark(ctx, db, user, "", id, "deleted_at"); err != nil {
		return "", err
	}
	return encode(map[string]any{"id": id, "deleted": true})
}

func notificationClear(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	planet, err := notificationScopeFrom(payload)
	if err != nil {
		return "", err
	}
	n, err := notificationMark(ctx, db, user, planet, "", "deleted_at")
	if err != nil {
		return "", err
	}
	return encode(map[string]int64{"cleared": n})
}

type appToggle struct {
	App *bool `json:"app"`
	SMS *bool `json:"sms"`
}

func preferencesView(categoryRaw, channelRaw []byte) map[string]any {
	categories := map[string]bool{}
	for k, v := range defaultCategoryPrefs {
		categories[k] = v
	}
	_ = json.Unmarshal(categoryRaw, &categories)
	channels := defaultChannelPrefs()
	_ = json.Unmarshal(channelRaw, &channels)
	app := func(t channelToggle) map[string]bool { return map[string]bool{"app": t.InApp, "sms": t.SMS} }
	return map[string]any{"categories": categories,
		"channels": map[string]any{"group": app(channels["group"]), "missions": app(channels["mission_game"]), "system": app(channels["system"])}}
}

func settingsPreferencesGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var categoryRaw, channelRaw []byte
	err = db.QueryRowContext(ctx, `SELECT category_prefs,channel_prefs FROM jazebe_notification_preference WHERE user_id=$1`, user).Scan(&categoryRaw, &channelRaw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "preferences unavailable")
	}
	return encode(preferencesView(categoryRaw, channelRaw))
}

func settingsPreferencesUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		Categories map[string]*bool     `json:"categories"`
		Channels   map[string]appToggle `json:"channels"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	for key := range in.Categories {
		if _, ok := defaultCategoryPrefs[key]; !ok {
			return "", status.Error(codes.InvalidArgument, "unknown preference category: "+key)
		}
	}
	appKeys := map[string]string{"group": "group", "missions": "mission_game", "system": "system"}
	for key := range in.Channels {
		if _, ok := appKeys[key]; !ok {
			return "", status.Error(codes.InvalidArgument, "unknown preference channel: "+key)
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "preferences unavailable")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_notification_preference(user_id) VALUES($1) ON CONFLICT (user_id) DO NOTHING`, user); err != nil {
		return "", status.Error(codes.Internal, "preferences unavailable")
	}
	var categoryRaw, channelRaw []byte
	if err := tx.QueryRowContext(ctx, `SELECT category_prefs,channel_prefs FROM jazebe_notification_preference WHERE user_id=$1 FOR UPDATE`, user).Scan(&categoryRaw, &channelRaw); err != nil {
		return "", status.Error(codes.Internal, "preferences unavailable")
	}
	categories := map[string]bool{}
	for k, v := range defaultCategoryPrefs {
		categories[k] = v
	}
	_ = json.Unmarshal(categoryRaw, &categories)
	for key, value := range in.Categories {
		if value != nil {
			categories[key] = *value
		}
	}
	channels := defaultChannelPrefs()
	_ = json.Unmarshal(channelRaw, &channels)
	for key, patch := range in.Channels {
		t := channels[appKeys[key]]
		if patch.App != nil {
			t.InApp = *patch.App
		}
		if patch.SMS != nil {
			t.SMS = *patch.SMS
		}
		channels[appKeys[key]] = t
	}
	newCategories, _ := json.Marshal(categories)
	newChannels, _ := json.Marshal(channels)
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_notification_preference SET category_prefs=$2::jsonb,channel_prefs=$3::jsonb,updated_at=now() WHERE user_id=$1`,
		user, string(newCategories), string(newChannels)); err != nil {
		return "", status.Error(codes.Internal, "preferences update failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "preferences update failed")
	}
	return encode(preferencesView(newCategories, newChannels))
}

type planetOrderItem struct {
	ID    string `json:"id"`
	Fa    string `json:"fa"`
	Color string `json:"color"`
}

// planetOrder is the constellation order: manual sort_order, then recent
// activity, then join time; only active planets.
func planetOrder(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, user string) ([]planetOrderItem, bool, error) {
	rows, err := q.QueryContext(ctx, `SELECT p.id,p.name,COALESCE(p.brand_color,'#FFFFFF'),m.sort_order IS NULL
		FROM jazebe_planet_membership m JOIN jazebe_planet p ON p.id=m.planet_id
		WHERE m.user_id=$1 AND m.is_active AND p.status='active'
		ORDER BY m.sort_order ASC NULLS LAST, m.last_activity_at DESC NULLS LAST, m.joined_at DESC`, user)
	if err != nil {
		return nil, false, status.Error(codes.Internal, "planet order unavailable")
	}
	defer rows.Close()
	items := []planetOrderItem{}
	auto := true
	for rows.Next() {
		var it planetOrderItem
		var unordered bool
		if err := rows.Scan(&it.ID, &it.Fa, &it.Color, &unordered); err != nil {
			return nil, false, status.Error(codes.Internal, "planet order unavailable")
		}
		auto = auto && unordered
		items = append(items, it)
	}
	return items, auto, rows.Err()
}

func settingsPlanetOrderGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	items, auto, err := planetOrder(ctx, db, user)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"planets": items, "auto": auto})
}

func settingsPlanetOrderUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		Planets []planetOrderItem `json:"planets"`
		Auto    bool              `json:"auto"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if len(in.Planets) > 200 {
		return "", status.Error(codes.InvalidArgument, "too many planets")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "planet order unavailable")
	}
	defer tx.Rollback()
	if in.Auto {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET sort_order=NULL WHERE user_id=$1`, user); err != nil {
			return "", status.Error(codes.Internal, "planet order update failed")
		}
	} else {
		current, _, err := planetOrder(ctx, tx, user)
		if err != nil {
			return "", err
		}
		byID, byName := map[string]string{}, map[string]string{}
		for _, p := range current {
			byID[p.ID] = p.ID
			byName[p.Fa] = p.ID
		}
		var ordered []string
		seen := map[string]bool{}
		for _, item := range in.Planets {
			hit := byID[item.ID]
			if hit == "" {
				hit = byName[item.Fa]
			}
			if hit == "" || seen[hit] {
				return "", status.Error(codes.InvalidArgument, "invalid planet order")
			}
			seen[hit] = true
			ordered = append(ordered, hit)
		}
		// Planets left out keep their place after the ordered ones.
		for _, p := range current {
			if !seen[p.ID] {
				ordered = append(ordered, p.ID)
			}
		}
		for i, planetID := range ordered {
			if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET sort_order=$3 WHERE user_id=$1 AND planet_id=$2`, user, planetID, i+1); err != nil {
				return "", status.Error(codes.Internal, "planet order update failed")
			}
		}
	}
	items, auto, err := planetOrder(ctx, tx, user)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "planet order update failed")
	}
	return encode(map[string]any{"planets": items, "auto": auto})
}
