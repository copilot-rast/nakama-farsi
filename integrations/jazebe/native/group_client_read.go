package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Read models of the main app's group screens (source group-read.service.ts:
// list, mine, detail and the creator-only history). Rank, honours and period
// stats are derived from the lottery tables at read time, as in the source.

const groupDefaultColor = "#2F6EBC"

type groupCard struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Members int    `json:"members"`
	Cap     int    `json:"cap"`
	Mode    string `json:"mode"`
	Rank    int    `json:"rank"`
	Sub     string `json:"sub,omitempty"`
	Desc    string `json:"desc,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Status  string `json:"status,omitempty"`
}

type groupRow struct {
	ID, Name, Mode, Status string
	Description, Icon      sql.NullString
	Color                  sql.NullString
	Members, Cap           int
	Created                time.Time
}

const groupRowColumns = `g.id,g.name,g.mode,g.status,g.description,g.icon,g.color,g.member_count,g.member_cap,g.created_at`

func scanGroupRow(row interface{ Scan(...any) error }) (groupRow, error) {
	var g groupRow
	err := row.Scan(&g.ID, &g.Name, &g.Mode, &g.Status, &g.Description, &g.Icon, &g.Color, &g.Members, &g.Cap, &g.Created)
	return g, err
}

func (g groupRow) card(rank int, withDescription bool) groupCard {
	c := groupCard{ID: g.ID, Name: g.Name, Color: groupDefaultColor, Members: g.Members, Cap: g.Cap, Mode: g.Mode, Rank: rank}
	if c.Rank < 1 {
		c.Rank = 1
	}
	if g.Color.Valid && g.Color.String != "" {
		c.Color = g.Color.String
	}
	if g.Description.Valid && g.Description.String != "" {
		c.Sub = g.Description.String
		if withDescription {
			c.Desc = g.Description.String
		}
	}
	if g.Icon.Valid {
		c.Icon = g.Icon.String
	}
	if g.Status != "active" {
		c.Status = g.Status
	}
	return c
}

type runningLottery struct {
	ID, Title string
	Starts    time.Time
}

// activeGroupPlanet requires an active planet of an active brand.
func activeGroupPlanet(ctx context.Context, db *sql.DB, planetID string) error {
	var ok bool
	err := db.QueryRowContext(ctx, `SELECT p.status='active' AND b.is_active FROM jazebe_planet p JOIN jazebe_brand b ON b.id=p.brand_id WHERE p.id=$1`, planetID).Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) || err == nil && !ok {
		return status.Error(codes.NotFound, "active planet not found")
	}
	if err != nil {
		return status.Error(codes.Internal, "planet unavailable")
	}
	return nil
}

func groupRunningLottery(ctx context.Context, db *sql.DB, planetID string, now time.Time) (*runningLottery, error) {
	var l runningLottery
	err := db.QueryRowContext(ctx, `SELECT id,title,starts_at FROM jazebe_lottery WHERE planet_id=$1 AND status='active' AND draw_at>$2
		ORDER BY draw_at ASC LIMIT 1`, planetID, now).Scan(&l.ID, &l.Title, &l.Starts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery unavailable")
	}
	return &l, nil
}

// groupRanks orders the planet's active groups by stored lottery rank, then
// ticket count, then member count, then age (source GroupReadService.ranks).
func groupRanks(ctx context.Context, db *sql.DB, planetID string, lottery *runningLottery) (map[string]int, error) {
	var lotteryID any
	if lottery != nil {
		lotteryID = lottery.ID
	}
	rows, err := db.QueryContext(ctx, `SELECT g.id FROM jazebe_group g
		LEFT JOIN jazebe_lottery_group_stat s ON s.group_id=g.id AND s.lottery_id=$2::uuid
		WHERE g.planet_id=$1 AND g.status='active'
		ORDER BY s.rank ASC NULLS LAST, COALESCE(s.ticket_count,-1) DESC, g.member_count DESC, g.created_at ASC, g.id ASC`, planetID, lotteryID)
	if err != nil {
		return nil, status.Error(codes.Internal, "group ranks unavailable")
	}
	defer rows.Close()
	ranks := map[string]int{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, status.Error(codes.Internal, "group ranks unavailable")
		}
		ranks[id] = len(ranks) + 1
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "group ranks unavailable")
	}
	return ranks, nil
}

func groupTickets(ctx context.Context, db *sql.DB, lottery *runningLottery, groupID string) (int, error) {
	if lottery == nil {
		return 0, nil
	}
	var n int
	err := db.QueryRowContext(ctx, `SELECT COALESCE((SELECT ticket_count FROM jazebe_lottery_group_stat WHERE lottery_id=$1 AND group_id=$2),0)`, lottery.ID, groupID).Scan(&n)
	if err != nil {
		return 0, status.Error(codes.Internal, "group tickets unavailable")
	}
	return n, nil
}

func groupHonors(ctx context.Context, db *sql.DB, groupID string) ([]map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT COALESCE(p.title,'جایزه'),COALESCE(l.title,'') FROM jazebe_lottery_winner w
		LEFT JOIN jazebe_lottery_prize p ON p.id=w.prize_id LEFT JOIN jazebe_lottery l ON l.id=w.lottery_id
		WHERE w.group_id=$1 ORDER BY w.recorded_at DESC LIMIT 10`, groupID)
	if err != nil {
		return nil, status.Error(codes.Internal, "group honours unavailable")
	}
	defer rows.Close()
	out := []map[string]string{}
	for rows.Next() {
		var title, subtitle string
		if err := rows.Scan(&title, &subtitle); err != nil {
			return nil, status.Error(codes.Internal, "group honours unavailable")
		}
		out = append(out, map[string]string{"title": title, "subtitle": subtitle})
	}
	return out, rows.Err()
}

// groupMemberOverview is the member home (source overviewFor).
func groupMemberOverview(ctx context.Context, db *sql.DB, g groupRow, planetID, viewer, role string, lottery *runningLottery, ranks map[string]int) (map[string]any, int, error) {
	honors, err := groupHonors(ctx, db, g.ID)
	if err != nil {
		return nil, 0, err
	}
	tickets, err := groupTickets(ctx, db, lottery, g.ID)
	if err != nil {
		return nil, 0, err
	}
	coins := 0
	if lottery != nil {
		if err := db.QueryRowContext(ctx, `SELECT COALESCE(sum(t.amount),0) FROM jazebe_point_transaction t
			WHERE t.planet_id=$1 AND t.direction='credit' AND t.created_at>=$3
			AND t.user_id IN (SELECT user_id FROM jazebe_group_membership WHERE group_id=$2 AND status='active')`,
			planetID, g.ID, lottery.Starts).Scan(&coins); err != nil {
			return nil, 0, status.Error(codes.Internal, "group coins unavailable")
		}
	}
	var unread int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_group_message m
		WHERE m.group_id=$1 AND m.deleted_at IS NULL AND m.seq>COALESCE((SELECT last_read_seq FROM jazebe_group_message_read WHERE group_id=$1 AND user_id=$2),0)
		AND (m.sender_user_id IS NULL OR m.sender_user_id<>$2)`, g.ID, viewer).Scan(&unread); err != nil {
		return nil, 0, status.Error(codes.Internal, "group messages unavailable")
	}
	var pinned any
	if lottery != nil {
		pinned = map[string]string{"id": lottery.ID, "title": lottery.Title}
	}
	var offer any
	var from, to string
	var expires time.Time
	err = db.QueryRowContext(ctx, `SELECT from_user_id,to_user_id,expires_at FROM jazebe_group_leadership_offer
		WHERE group_id=$1 AND status='pending' AND expires_at>now() AND (to_user_id=$2 OR from_user_id=$2) LIMIT 1`, g.ID, viewer).Scan(&from, &to, &expires)
	switch {
	case err == nil && to == viewer:
		offer = map[string]any{"direction": "incoming", "otherUserId": from, "expiresAt": expires}
	case err == nil:
		offer = map[string]any{"direction": "outgoing", "otherUserId": to, "expiresAt": expires}
	case !errors.Is(err, sql.ErrNoRows):
		return nil, 0, status.Error(codes.Internal, "leadership offer unavailable")
	}
	return map[string]any{
		"group":           g.card(ranks[g.ID], true),
		"viewerRole":      role,
		"honors":          honors,
		"stats":           map[string]int{"coins": coins, "tickets": tickets, "unreadMessages": unread},
		"pinnedRaffle":    pinned,
		"leadershipOffer": offer,
	}, tickets, nil
}

func groupList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := userID(ctx); err != nil {
		return "", err
	}
	var in struct {
		PlanetID string `json:"planetId"`
		Sort     string `json:"sort"`
		Search   string `json:"search"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	switch in.Sort {
	case "":
		in.Sort = "rank"
	case "rank", "members", "date":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid group sort")
	}
	in.Search = strings.TrimSpace(in.Search)
	if !utf8.ValidString(in.Search) || utf8.RuneCountInString(in.Search) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid group search")
	}
	if err := activeGroupPlanet(ctx, db, planetID); err != nil {
		return "", err
	}
	lottery, err := groupRunningLottery(ctx, db, planetID, time.Now().UTC())
	if err != nil {
		return "", err
	}
	ranks, err := groupRanks(ctx, db, planetID, lottery)
	if err != nil {
		return "", err
	}
	// Escape LIKE wildcards; the value itself is a bound parameter.
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(in.Search) + "%"
	rows, err := db.QueryContext(ctx, `SELECT `+groupRowColumns+` FROM jazebe_group g
		WHERE g.planet_id=$1 AND g.status='active' AND ($2='' OR g.name ILIKE $3)
		ORDER BY g.created_at DESC, g.id DESC LIMIT 500`, planetID, in.Search, pattern)
	if err != nil {
		return "", status.Error(codes.Internal, "groups unavailable")
	}
	defer rows.Close()
	cards := []groupCard{}
	for rows.Next() {
		g, err := scanGroupRow(rows)
		if err != nil {
			return "", status.Error(codes.Internal, "groups unavailable")
		}
		cards = append(cards, g.card(ranks[g.ID], false))
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "groups unavailable")
	}
	switch in.Sort {
	case "members":
		sortCards(cards, func(a, b groupCard) bool {
			return a.Members > b.Members || a.Members == b.Members && a.Rank < b.Rank
		})
	case "rank":
		sortCards(cards, func(a, b groupCard) bool {
			return a.Rank < b.Rank || a.Rank == b.Rank && a.Members > b.Members
		})
	}
	return encode(map[string]any{"items": cards})
}

func sortCards(cards []groupCard, less func(a, b groupCard) bool) {
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0 && less(cards[j], cards[j-1]); j-- {
			cards[j], cards[j-1] = cards[j-1], cards[j]
		}
	}
}

func groupMine(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err := activeGroupPlanet(ctx, db, planetID); err != nil {
		return "", err
	}
	var role string
	g, err := scanGroupRow(db.QueryRowContext(ctx, `SELECT `+groupRowColumns+` FROM jazebe_group_membership m
		JOIN jazebe_group g ON g.id=m.group_id WHERE m.planet_id=$1 AND m.user_id=$2 AND m.status='active'`, planetID, user))
	if errors.Is(err, sql.ErrNoRows) {
		return encode(map[string]any{"group": nil})
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	if err := db.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active'`, g.ID, user).Scan(&role); err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	lottery, err := groupRunningLottery(ctx, db, planetID, time.Now().UTC())
	if err != nil {
		return "", err
	}
	ranks, err := groupRanks(ctx, db, planetID, lottery)
	if err != nil {
		return "", err
	}
	overview, _, err := groupMemberOverview(ctx, db, g, planetID, user, role, lottery, ranks)
	if err != nil {
		return "", err
	}
	return encode(overview)
}

func groupDetail(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	if err := activeGroupPlanet(ctx, db, ref.PlanetID); err != nil {
		return "", err
	}
	g, err := scanGroupRow(db.QueryRowContext(ctx, `SELECT `+groupRowColumns+` FROM jazebe_group g WHERE g.id=$1 AND g.planet_id=$2`, ref.GroupID, ref.PlanetID))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	lottery, err := groupRunningLottery(ctx, db, ref.PlanetID, time.Now().UTC())
	if err != nil {
		return "", err
	}
	ranks, err := groupRanks(ctx, db, ref.PlanetID, lottery)
	if err != nil {
		return "", err
	}
	var role string
	err = db.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active'`, g.ID, user).Scan(&role)
	if err == nil {
		overview, tickets, err := groupMemberOverview(ctx, db, g, ref.PlanetID, user, role, lottery, ranks)
		if err != nil {
			return "", err
		}
		overview["view"] = "member"
		overview["tickets"] = tickets
		return encode(overview)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "group membership unavailable")
	}
	honors, err := groupHonors(ctx, db, g.ID)
	if err != nil {
		return "", err
	}
	tickets, err := groupTickets(ctx, db, lottery, g.ID)
	if err != nil {
		return "", err
	}
	var requested bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_join_request WHERE group_id=$1 AND user_id=$2 AND status='pending')`, g.ID, user).Scan(&requested); err != nil {
		return "", status.Error(codes.Internal, "group request unavailable")
	}
	return encode(map[string]any{"view": "profile", "group": g.card(ranks[g.ID], false), "honors": honors, "tickets": tickets, "requested": requested})
}

// groupHistory is the creator's view of what managers did (source
// GroupLedgerService.history over jz_group_audit). Native group audits live in
// jazebe_audit_log; this maps them to the source actions and Persian lines.
func groupHistory(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	if err := requireActiveGroupViewer(ctx, db, ref, user, true); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT a.id,a.action,a.actor_type,a.actor_id,a.params,a.created_at,
		CASE WHEN a.action='group.request.decision' THEN r.user_id::text END
		FROM jazebe_audit_log a LEFT JOIN jazebe_group_join_request r ON a.action='group.request.decision' AND r.id=a.target_id
		WHERE a.planet_id=$2 AND a.action<>'group.create' AND (
			(a.target_type='group' AND a.target_id=$1) OR
			(a.target_type IN ('group_request','group_message') AND a.params->>'groupId'=$1::text))
		ORDER BY a.created_at DESC, a.id DESC LIMIT 50`, ref.GroupID, ref.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "group history unavailable")
	}
	defer rows.Close()
	type entry struct {
		id, action, actorType, actor, requester string
		params                                  map[string]any
		at                                      time.Time
	}
	var entries []entry
	ids := map[string]bool{}
	for rows.Next() {
		var e entry
		var actor, requester sql.NullString
		var raw []byte
		if err := rows.Scan(&e.id, &e.action, &e.actorType, &actor, &raw, &e.at, &requester); err != nil {
			return "", status.Error(codes.Internal, "group history unavailable")
		}
		e.actor, e.requester = actor.String, requester.String
		_ = json.Unmarshal(raw, &e.params)
		entries = append(entries, e)
		for _, id := range []string{e.actor, e.requester, paramString(e.params, "userId"), paramString(e.params, "to"), paramString(e.params, "from")} {
			if id != "" {
				ids[id] = true
			}
		}
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "group history unavailable")
	}
	names, err := displayNames(ctx, db, ids)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	out := []map[string]any{}
	for _, e := range entries {
		action, kind, actorID, targetID := groupHistoryAction(e.action, e.params)
		if action == "" {
			continue
		}
		if actorID == "" {
			actorID = e.actor
		}
		if e.action == "group.request.decision" {
			targetID = e.requester
		}
		actorName := nameOr(names, actorID)
		var target any
		targetName := ""
		if targetID != "" {
			targetName = nameOr(names, targetID)
			target = map[string]string{"id": targetID, "name": targetName}
		}
		out = append(out, map[string]any{"id": e.id, "kind": kind, "action": action,
			"text": groupHistoryText(action, actorName, targetName, e.actorType, e.params),
			"time": groupRelativeFa(e.at, now), "at": e.at, "actor": map[string]string{"id": actorID, "name": actorName},
			"target": target, "payload": e.params})
	}
	return encode(map[string]any{"items": out})
}

// groupHistoryAction maps a native audit action to the source GroupAuditAction,
// its history kind, and the actor/target the source line names.
func groupHistoryAction(action string, params map[string]any) (sourceAction, kind, actor, target string) {
	switch action {
	case "group.request.decision":
		if paramString(params, "decision") == "reject" {
			return "reject", "reject", "", ""
		}
		return "approve", "approve", "", ""
	case "group.member.remove":
		return "remove", "remove", "", paramString(params, "userId")
	case "group.member.promote":
		return "promote", "promote", "", paramString(params, "userId")
	case "group.member.demote":
		return "demote", "promote", "", paramString(params, "userId")
	case "group.transfer":
		return "transfer", "promote", paramString(params, "from"), paramString(params, "to")
	case "group.transfer.offer":
		return "transfer_offer", "promote", "", paramString(params, "to")
	case "group.transfer.decline":
		return "transfer_decline", "reject", "", ""
	case "group.message.delete":
		return "delete_message", "remove", "", ""
	case "group.settings.update":
		return "edit_settings", "approve", "", ""
	case "group.disband":
		return "disband", "remove", "", ""
	}
	return "", "", "", ""
}

func groupHistoryText(action, actor, target, actorType string, params map[string]any) string {
	t := target
	if t == "" {
		t = "ناشناس"
	}
	switch action {
	case "approve":
		return actor + " درخواست «" + t + "» را تأیید کرد"
	case "reject":
		return actor + " درخواست «" + t + "» را رد کرد"
	case "remove":
		return actor + " «" + t + "» را از گروه حذف کرد"
	case "promote":
		return actor + " «" + t + "» را به مدیر ارتقا داد"
	case "demote":
		return actor + " «" + t + "» را به عضو عادی تغییر داد"
	case "transfer":
		if v, _ := params["automatic"].(bool); v {
			return "رهبری گروه پس از خروج " + actor + " به «" + t + "» رسید"
		}
		return actor + " رهبری گروه را به «" + t + "» منتقل کرد"
	case "transfer_offer":
		return actor + " رهبری گروه را به «" + t + "» پیشنهاد داد"
	case "transfer_decline":
		return actor + " پیشنهاد رهبری گروه را نپذیرفت"
	case "delete_message":
		switch actorType {
		case "operator":
			return actor + " (اپراتور) یک پیام را حذف کرد"
		case "brand_admin":
			return actor + " (مدیر برند) یک پیام را حذف کرد"
		}
		return actor + " یک پیام نامناسب را حذف کرد"
	case "edit_settings":
		if fields, ok := params["fields"].([]any); ok && len(fields) > 0 {
			return actor + " تنظیمات گروه را تغییر داد (" + faNumber(len(fields)) + " مورد)"
		}
		return actor + " تنظیمات گروه را تغییر داد"
	case "disband":
		if actorType == "operator" {
			return "گروه توسط اپراتور منحل شد"
		}
		return actor + " گروه را منحل کرد"
	}
	return actor + " " + action
}

func paramString(params map[string]any, key string) string {
	s, _ := params[key].(string)
	return s
}

func displayNames(ctx context.Context, db *sql.DB, ids map[string]bool) (map[string]string, error) {
	names := map[string]string{}
	if len(ids) == 0 {
		return names, nil
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		if _, err := requiredUUID(id); err == nil {
			list = append(list, id)
		}
	}
	rows, err := db.QueryContext(ctx, `SELECT id::text,COALESCE(NULLIF(trim(display_name),''),'کاربر ناشناس') FROM users WHERE id::text = ANY($1::text[])`, list)
	if err != nil {
		return nil, status.Error(codes.Internal, "profiles unavailable")
	}
	defer rows.Close()
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, status.Error(codes.Internal, "profiles unavailable")
		}
		names[id] = name
	}
	return names, rows.Err()
}

func nameOr(names map[string]string, id string) string {
	if n, ok := names[id]; ok {
		return n
	}
	return "کاربر ناشناس"
}
