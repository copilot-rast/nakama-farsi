package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/clanhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
)

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type memberRow struct {
	ClanID   string
	Role     string
	JoinedAt time.Time
}

func (s *service) memberOf(ctx context.Context, q queryer, userID string) (*memberRow, error) {
	var m memberRow
	err := q.QueryRowContext(ctx, `SELECT group_id::text, role, joined_at FROM clan_member WHERE user_id::text=$1`, userID).Scan(&m.ClanID, &m.Role, &m.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *service) membership(ctx context.Context, userID string) (clanhooks.Membership, bool) {
	m, err := s.memberOf(ctx, s.db, userID)
	if err != nil || m == nil {
		return clanhooks.Membership{}, false
	}
	return clanhooks.Membership{ClanID: m.ClanID, Role: m.Role}, true
}

func (s *service) memberIDs(ctx context.Context, clanID string) []string {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text FROM clan_member WHERE group_id::text=$1`, clanID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out
}

type clanRow struct {
	ID          string
	Name        string
	Description string
	AvatarURL   string
	Lang        string
	Members     int
	MaxMembers  int
	Tag         string
	Badge       map[string]any
	JoinType    string
	MinTrophies int
	Level       int
	XP          int64
	Trophies    int64
	WarWins     int
	WarLosses   int
	WarDraws    int
	WarPoints   int64
	CreatedAt   time.Time
	Motto       string
	Color       string
	Country     string
	Tags        []string
	EmblemKey   string
	CoverKey    string
	CoverURL    string
	LookChangedAt *time.Time
}

const clanCols = `c.group_id::text, g.name, g.description, g.avatar_url, g.lang_tag, g.edge_count, g.max_count, c.tag, c.badge::text, c.join_type, c.min_trophies,
	c.level, c.xp, c.trophies, c.war_wins, c.war_losses, c.war_draws, c.war_points, c.created_at,
	c.motto, c.color, c.country, array_to_json(c.tags)::text, c.emblem_key, c.cover_key, c.cover_url, c.look_changed_at`

func scanClan(row interface{ Scan(...any) error }) (*clanRow, error) {
	c := &clanRow{}
	var badge, tags string
	if err := row.Scan(&c.ID, &c.Name, &c.Description, &c.AvatarURL, &c.Lang, &c.Members, &c.MaxMembers, &c.Tag, &badge, &c.JoinType, &c.MinTrophies,
		&c.Level, &c.XP, &c.Trophies, &c.WarWins, &c.WarLosses, &c.WarDraws, &c.WarPoints, &c.CreatedAt,
		&c.Motto, &c.Color, &c.Country, &tags, &c.EmblemKey, &c.CoverKey, &c.CoverURL, &c.LookChangedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(tags), &c.Tags)
	if c.Tags == nil {
		c.Tags = []string{}
	}
	_ = json.Unmarshal([]byte(badge), &c.Badge)
	if c.Badge == nil {
		c.Badge = map[string]any{}
	}
	return c, nil
}

var errNoClan = apierr.NotFound("CLAN_NOT_FOUND").WithField("clanId")

func (s *service) clanByID(ctx context.Context, q queryer, id string) (*clanRow, error) {
	if !uuidPattern.MatchString(strings.TrimSpace(id)) {
		return nil, apierr.BadRequest("CLAN_ID_REQUIRED").WithField("clanId")
	}
	c, err := scanClan(q.QueryRowContext(ctx, `SELECT `+clanCols+` FROM clan c JOIN groups g ON g.id=c.group_id WHERE c.group_id::text=$1`, strings.ToLower(strings.TrimSpace(id))))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNoClan
	}
	return c, err
}

func (s *service) trophiesOf(ctx context.Context, userID string) int {
	var t int
	_ = s.db.QueryRowContext(ctx, `SELECT trophies FROM trophy_player WHERE user_id::text=$1`, userID).Scan(&t)
	return t
}

func (s *service) username(ctx context.Context, userID string) string {
	var u string
	_ = s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id::text=$1`, userID).Scan(&u)
	return u
}

func (s *service) requireMember(ctx context.Context, userID string) (*memberRow, error) {
	m, err := s.memberOf(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, apierr.Forbidden("CLAN_NOT_MEMBER")
	}
	return m, nil
}

func (s *service) requireRole(ctx context.Context, userID string, roles []string) (*memberRow, error) {
	m, err := s.requireMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !allowed(roles, m.Role) {
		return nil, apierr.Forbidden("CLAN_ROLE_NOT_ALLOWED").With("role", m.Role)
	}
	return m, nil
}

func (s *service) cooldown(ctx context.Context, cfg settings.All, userID, clanID string) error {
	now := s.now()
	if cfg.Cooldowns.JoinAfterLeaveMinutes > 0 {
		var at time.Time
		err := s.db.QueryRowContext(ctx, `SELECT max(created_at) FROM clan_cooldown WHERE user_id::text=$1 HAVING max(created_at) IS NOT NULL`, userID).Scan(&at)
		if err == nil {
			if left := at.Add(time.Duration(cfg.Cooldowns.JoinAfterLeaveMinutes) * time.Minute).Sub(now); left > 0 {
				return apierr.Conflict("CLAN_COOLDOWN").With("retryAfter", int(math.Ceil(left.Seconds())))
			}
		}
	}
	if clanID != "" && cfg.Cooldowns.RejoinAfterKickHours > 0 {
		var at time.Time
		err := s.db.QueryRowContext(ctx, `SELECT created_at FROM clan_cooldown WHERE user_id::text=$1 AND group_id::text=$2 AND reason='kicked'`, userID, clanID).Scan(&at)
		if err == nil {
			if left := at.Add(time.Duration(cfg.Cooldowns.RejoinAfterKickHours) * time.Hour).Sub(now); left > 0 {
				return apierr.Conflict("CLAN_KICK_COOLDOWN").With("retryAfter", int(math.Ceil(left.Seconds())))
			}
		}
	}
	return nil
}

const tagAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func newTag() string {
	b := make([]byte, 7)
	_, _ = rand.Read(b)
	out := make([]byte, len(b))
	for i, x := range b {
		out[i] = tagAlphabet[int(x)%len(tagAlphabet)]
	}
	return string(out)
}

func cleanBadge(cfg settings.All, in map[string]any) (map[string]any, error) {
	out := map[string]any{}
	for k, list := range map[string][]string{"emblem": cfg.Badge.Emblems, "color": cfg.Badge.Colors, "frame": cfg.Badge.Frames} {
		v := str(in[k])
		if v == "" {
			continue
		}
		if !settings.ShortPattern.MatchString(v) || (len(list) > 0 && !allowed(list, v)) {
			return nil, apierr.BadRequest("CLAN_BADGE_INVALID").WithField("badge." + k)
		}
		out[k] = v
	}
	return out, nil
}

type profileInput struct {
	lookInput
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	AvatarURL   *string        `json:"avatarUrl"`
	Lang        *string        `json:"lang"`
	Badge       map[string]any `json:"badge"`
	JoinType    *string        `json:"joinType"`
	MinTrophies *int           `json:"minTrophies"`
}

func cleanName(cfg settings.All, v string) (string, error) {
	v = strings.Join(strings.Fields(v), " ")
	n := utf8.RuneCountInString(v)
	if n < cfg.Create.NameMin || n > cfg.Create.NameMax {
		return "", apierr.BadRequest("CLAN_NAME_INVALID").WithField("name").With("min", cfg.Create.NameMin).With("max", cfg.Create.NameMax)
	}
	return v, nil
}

func (p profileInput) check(cfg settings.All) error {
	if p.Description != nil && utf8.RuneCountInString(*p.Description) > cfg.Create.DescriptionMax {
		return apierr.BadRequest("CLAN_DESCRIPTION_TOO_LONG").WithField("description").With("max", cfg.Create.DescriptionMax)
	}
	if p.AvatarURL != nil && len(*p.AvatarURL) > 512 {
		return apierr.BadRequest("CLAN_AVATAR_INVALID").WithField("avatarUrl")
	}
	if p.Lang != nil && len(*p.Lang) > 18 {
		return apierr.BadRequest("CLAN_LANG_INVALID").WithField("lang")
	}
	if p.JoinType != nil && *p.JoinType != settings.JoinOpen && *p.JoinType != settings.JoinClosed && *p.JoinType != settings.JoinInvite {
		return apierr.BadRequest("CLAN_JOIN_TYPE_INVALID").WithField("joinType")
	}
	if p.MinTrophies != nil && (*p.MinTrophies < 0 || *p.MinTrophies > cfg.Create.MaxMinTrophies) {
		return apierr.BadRequest("CLAN_MIN_TROPHIES_INVALID").WithField("minTrophies").With("max", cfg.Create.MaxMinTrophies)
	}
	return nil
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func (s *service) enabled(cfg settings.All) error {
	if !cfg.Enabled {
		return apierr.Forbidden("CLAN_DISABLED")
	}
	return nil
}

func (s *service) rpcCreate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		profileInput
		RequestID string `json:"requestId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if err := s.enabled(cfg); err != nil {
		return nil, err
	}
	if !cfg.Create.Enabled {
		return nil, apierr.Forbidden("CLAN_CREATE_DISABLED")
	}
	if m, err := s.memberOf(ctx, s.db, userID); err != nil || m != nil {
		if err != nil {
			return nil, err
		}
		return nil, apierr.Conflict("CLAN_ALREADY_MEMBER").With("clanId", m.ClanID)
	}
	if err := s.cooldown(ctx, cfg, userID, ""); err != nil {
		return nil, err
	}
	if t := s.trophiesOf(ctx, userID); t < cfg.Create.MinTrophies {
		return nil, apierr.Forbidden("CLAN_NOT_ENOUGH_TROPHIES").With("need", cfg.Create.MinTrophies).With("have", t)
	}
	name, err := cleanName(cfg, deref(in.Name))
	if err != nil {
		return nil, err
	}
	if err := in.check(cfg); err != nil {
		return nil, err
	}
	badge, err := cleanBadge(cfg, in.Badge)
	if err != nil {
		return nil, err
	}
	joinType := cfg.Create.DefaultJoinType
	if in.JoinType != nil {
		joinType = *in.JoinType
	}
	minT := 0
	if in.MinTrophies != nil {
		minT = *in.MinTrophies
	}
	req := strings.TrimSpace(in.RequestID)
	if req == "" {
		req = uuid.Must(uuid.NewV4()).String()
	}
	if len(req) > 64 {
		return nil, apierr.BadRequest("INVALID_REQUEST_ID").WithField("requestId")
	}
	if in.AvatarURL != nil {
		return nil, apierr.BadRequest("CLAN_AVATAR_USE_GALLERY").WithField("avatarUrl")
	}
	f := s.wordFilter(ctx, cfg)
	if blocked(f, name) {
		return nil, apierr.BadRequest("CLAN_NAME_BLOCKED").WithField("name")
	}
	var taken bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM groups WHERE lower(name)=lower($1))`, name).Scan(&taken)
	if taken {
		return nil, apierr.Conflict("CLAN_NAME_TAKEN").WithField("name")
	}
	if in.EmblemKey == nil || *in.EmblemKey == "" {
		if d := s.defaultArt(ctx, artEmblem); d != nil {
			k := d.Key
			in.EmblemKey = &k
		}
	}
	if in.CoverKey == nil || *in.CoverKey == "" {
		if d := s.defaultArt(ctx, artCover); d != nil {
			k := d.Key
			in.CoverKey = &k
		}
	}
	if cfg.Profile.RequireEmblem && (in.EmblemKey == nil || *in.EmblemKey == "") {
		return nil, apierr.BadRequest("CLAN_EMBLEM_REQUIRED").WithField("emblemKey")
	}
	probe := &clanRow{Level: 1, Tags: []string{}, Name: name}
	plan, err := s.planProfile(ctx, cfg, probe, profileInput{lookInput: in.lookInput, Lang: in.Lang, Description: in.Description}, editor{userID: userID})
	if err != nil {
		return nil, err
	}
	cap := s.road(ctx).memberCap(1, 30)
	id, err := s.groups.Create(ctx, userID, name, deref(in.Lang), deref(in.Description), "", joinType == settings.JoinOpen, cap)
	if err != nil {
		if errors.Is(err, runtime.ErrGroupNameInUse) || strings.Contains(err.Error(), "name") && strings.Contains(err.Error(), "use") {
			return nil, apierr.Conflict("CLAN_NAME_TAKEN").WithField("name")
		}
		return nil, err
	}
	undo := func() { _ = s.groups.Delete(context.Background(), id) }
	raw, _ := json.Marshal(badge)
	err = pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		var ierr error
		for i := 0; i < 5; i++ {
			_, ierr = tx.ExecContext(ctx, `INSERT INTO clan (group_id, tag, badge, join_type, min_trophies, created_by) VALUES ($1,$2,$3,$4,$5,$6)`, id, newTag(), raw, joinType, minT, userID)
			var pg *pgconn.PgError
			if !(errors.As(ierr, &pg) && pg.Code == "23505" && strings.Contains(pg.ConstraintName, "tag")) {
				break
			}
		}
		if ierr != nil {
			return ierr
		}
		_, ierr = tx.ExecContext(ctx, `INSERT INTO clan_member (group_id, user_id, role) VALUES ($1,$2,'leader')`, id, userID)
		return ierr
	})
	if err != nil {
		undo()
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return nil, apierr.Conflict("CLAN_ALREADY_MEMBER")
		}
		return nil, err
	}
	if err := s.debit(ctx, userID, "clan-create:"+userID+":"+req, cfg.Create.Cost); err != nil {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM clan WHERE group_id=$1`, id)
		undo()
		var ins *pluginkit.ErrInsufficient
		if errors.As(err, &ins) {
			return nil, apierr.Conflict("CLAN_NOT_ENOUGH_CURRENCY").With("currency", ins.Path)
		}
		return nil, err
	}
	if c, err := s.clanByID(ctx, s.db, id); err == nil {
		plan.changed = false
		if err := s.applyPlan(ctx, c, joinType, plan, editor{userID: userID}); err != nil {
			_, _ = s.db.ExecContext(ctx, `DELETE FROM clan WHERE group_id=$1`, id)
			undo()
			return nil, err
		}
	}
	s.refreshTrophies(ctx, id)
	s.say(ctx, cfg, id, "created", userID, "")
	c, err := s.clanByID(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"clan": s.clanView(ctx, c, userID, ""), "role": settings.RoleLeader}, nil
}

func (s *service) rpcUpdate(ctx context.Context, userID, payload string) (any, error) {
	var in profileInput
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if err := s.enabled(cfg); err != nil {
		return nil, err
	}
	m, err := s.requireRole(ctx, userID, cfg.Roles.Edit)
	if err != nil {
		return nil, err
	}
	if err := in.check(cfg); err != nil {
		return nil, err
	}
	c, err := s.clanByID(ctx, s.db, m.ClanID)
	if err != nil {
		return nil, err
	}
	if err := s.applyProfile(ctx, cfg, c, in, editor{userID: userID}); err != nil {
		return nil, err
	}
	s.say(ctx, cfg, c.ID, "updated", userID, "")
	c, err = s.clanByID(ctx, s.db, m.ClanID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"clan": s.clanView(ctx, c, userID, m.Role)}, nil
}

func (s *service) applyProfile(ctx context.Context, cfg settings.All, c *clanRow, in profileInput, who editor) error {
	plan, err := s.planProfile(ctx, cfg, c, in, who)
	if err != nil {
		return err
	}
	joinType := c.JoinType
	if in.JoinType != nil {
		joinType = *in.JoinType
	}
	badge := c.Badge
	if in.Badge != nil {
		b, err := cleanBadge(cfg, in.Badge)
		if err != nil {
			return err
		}
		badge = b
	}
	minT := c.MinTrophies
	if in.MinTrophies != nil {
		minT = *in.MinTrophies
	}
	if err := s.applyPlan(ctx, c, joinType, plan, who); err != nil {
		return err
	}
	raw, _ := json.Marshal(badge)
	_, err = s.db.ExecContext(ctx, `UPDATE clan SET badge=$2, join_type=$3, min_trophies=$4, updated_at=now() WHERE group_id::text=$1`, c.ID, raw, joinType, minT)
	return err
}

func (s *service) isFull(ctx context.Context, clanID string) bool {
	var count, max int
	if err := s.db.QueryRowContext(ctx, `SELECT edge_count, max_count FROM groups WHERE id::text=$1`, clanID).Scan(&count, &max); err != nil {
		return false
	}
	return max > 0 && count >= max
}

func (s *service) addMember(ctx context.Context, cfg settings.All, c *clanRow, userID string, native func() error) error {
	if s.isFull(ctx, c.ID) {
		return apierr.Conflict("CLAN_FULL").With("max", c.MaxMembers)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO clan_member (group_id, user_id, role) VALUES ($1,$2,'member')`, c.ID, userID); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return apierr.Conflict("CLAN_ALREADY_MEMBER")
		}
		return err
	}
	if err := native(); err != nil {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_member WHERE group_id=$1 AND user_id=$2`, c.ID, userID)
		if errors.Is(err, runtime.ErrGroupFull) {
			return apierr.Conflict("CLAN_FULL").With("max", c.MaxMembers)
		}
		return err
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_request WHERE user_id=$1`, userID)
	s.dropOtherRequests(ctx, userID, c.ID)
	s.refreshTrophies(ctx, c.ID)
	return nil
}

func (s *service) dropOtherRequests(ctx context.Context, userID, keep string) {
	rows, err := s.db.QueryContext(ctx, `SELECT e.source_id::text FROM group_edge e JOIN clan c ON c.group_id=e.source_id WHERE e.destination_id::text=$1 AND e.state=3 AND e.source_id::text<>$2`, userID, keep)
	if err != nil {
		return
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		_ = s.groups.Kick(ctx, id, []string{userID})
	}
}

func (s *service) rpcJoin(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ClanID string `json:"clanId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if err := s.enabled(cfg); err != nil {
		return nil, err
	}
	c, err := s.clanByID(ctx, s.db, in.ClanID)
	if err != nil {
		return nil, err
	}
	if m, err := s.memberOf(ctx, s.db, userID); err != nil || m != nil {
		if err != nil {
			return nil, err
		}
		return nil, apierr.Conflict("CLAN_ALREADY_MEMBER").With("clanId", m.ClanID)
	}
	if err := s.cooldown(ctx, cfg, userID, c.ID); err != nil {
		return nil, err
	}
	var invited bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_request WHERE group_id=$1 AND user_id=$2 AND kind='invite' AND (expires_at IS NULL OR expires_at > now()))`, c.ID, userID).Scan(&invited)
	if invited {
		return s.joinInvited(ctx, cfg, c, userID)
	}
	if t := s.trophiesOf(ctx, userID); t < c.MinTrophies {
		return nil, apierr.Forbidden("CLAN_NOT_ENOUGH_TROPHIES").With("need", c.MinTrophies).With("have", t)
	}
	uname := s.username(ctx, userID)
	switch c.JoinType {
	case settings.JoinInvite:
		return nil, apierr.Forbidden("CLAN_INVITE_ONLY")
	case settings.JoinClosed:
		var pending int
		_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM group_edge e JOIN clan c ON c.group_id=e.source_id WHERE e.destination_id::text=$1 AND e.state=3`, userID).Scan(&pending)
		var already bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM group_edge WHERE source_id::text=$1 AND destination_id::text=$2 AND state=3)`, c.ID, userID).Scan(&already)
		if already {
			return map[string]any{"status": "requested", "clanId": c.ID, "replayed": true}, nil
		}
		if pending >= cfg.Cooldowns.RequestsPerUser {
			return nil, apierr.Conflict("CLAN_TOO_MANY_REQUESTS").With("max", cfg.Cooldowns.RequestsPerUser)
		}
		if s.isFull(ctx, c.ID) {
			return nil, apierr.Conflict("CLAN_FULL").With("max", c.MaxMembers)
		}
		if err := s.groups.Join(ctx, c.ID, userID, uname); err != nil {
			if errors.Is(err, runtime.ErrGroupFull) {
				return nil, apierr.Conflict("CLAN_FULL").With("max", c.MaxMembers)
			}
			return nil, err
		}
		s.notify(ctx, s.withRoles(ctx, c.ID, cfg.Roles.Accept), "clan_join_request", map[string]any{"clanId": c.ID, "userId": userID, "username": uname}, true)
		return map[string]any{"status": "requested", "clanId": c.ID, "replayed": false}, nil
	}
	if err := s.addMember(ctx, cfg, c, userID, func() error { return s.groups.Join(ctx, c.ID, userID, uname) }); err != nil {
		return nil, err
	}
	s.say(ctx, cfg, c.ID, "joined", userID, "")
	s.notify(ctx, s.except(s.memberIDs(ctx, c.ID), userID), "clan_member_joined", map[string]any{"clanId": c.ID, "userId": userID, "username": uname}, false)
	c, _ = s.clanByID(ctx, s.db, c.ID)
	return map[string]any{"status": "joined", "clanId": c.ID, "clan": s.clanView(ctx, c, userID, settings.RoleMember)}, nil
}

func (s *service) joinInvited(ctx context.Context, cfg settings.All, c *clanRow, userID string) (any, error) {
	if err := s.addMember(ctx, cfg, c, userID, func() error { return s.groups.Add(ctx, c.ID, []string{userID}) }); err != nil {
		return nil, err
	}
	uname := s.username(ctx, userID)
	s.say(ctx, cfg, c.ID, "joined", userID, "")
	s.notify(ctx, s.except(s.memberIDs(ctx, c.ID), userID), "clan_member_joined", map[string]any{"clanId": c.ID, "userId": userID, "username": uname}, false)
	c, _ = s.clanByID(ctx, s.db, c.ID)
	return map[string]any{"status": "joined", "clanId": c.ID, "clan": s.clanView(ctx, c, userID, settings.RoleMember)}, nil
}

func (s *service) except(ids []string, drop string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != drop {
			out = append(out, id)
		}
	}
	return out
}

func (s *service) withRoles(ctx context.Context, clanID string, roles []string) []string {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text FROM clan_member WHERE group_id::text=$1 AND role = ANY($2)`, clanID, roles)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out
}

func (s *service) successor(ctx context.Context, clanID, leaving string) string {
	var id string
	_ = s.db.QueryRowContext(ctx, `SELECT user_id::text FROM clan_member WHERE group_id::text=$1 AND user_id::text<>$2
		ORDER BY CASE role WHEN 'co_leader' THEN 0 WHEN 'elder' THEN 1 ELSE 2 END, joined_at LIMIT 1`, clanID, leaving).Scan(&id)
	return id
}

func (s *service) rpcLeave(ctx context.Context, userID, payload string) (any, error) {
	cfg := s.settings(ctx)
	m, err := s.requireMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cfg.Cooldowns.LeaveAfterJoinMinutes > 0 {
		if left := m.JoinedAt.Add(time.Duration(cfg.Cooldowns.LeaveAfterJoinMinutes) * time.Minute).Sub(s.now()); left > 0 {
			return nil, apierr.Conflict("CLAN_LEAVE_COOLDOWN").With("retryAfter", int(math.Ceil(left.Seconds())))
		}
	}
	uname := s.username(ctx, userID)
	deleted := false
	newLeader := ""
	if m.Role == settings.RoleLeader {
		next := s.successor(ctx, m.ClanID, userID)
		if next == "" {
			if err := s.groups.Delete(ctx, m.ClanID); err != nil && !errors.Is(err, runtime.ErrGroupNotFound) {
				return nil, err
			}
			_, _ = s.db.ExecContext(ctx, `DELETE FROM clan WHERE group_id::text=$1`, m.ClanID)
			deleted = true
		} else {
			if cfg.Roles.LeaderLeave == settings.LeaveBlock {
				return nil, apierr.Conflict("CLAN_LEADER_MUST_TRANSFER")
			}
			if err := s.makeLeader(ctx, m.ClanID, next); err != nil {
				return nil, err
			}
			newLeader = next
		}
	}
	if !deleted {
		if err := s.groups.Leave(ctx, m.ClanID, userID, uname); err != nil && !errors.Is(err, runtime.ErrGroupUserNotFound) {
			return nil, err
		}
		_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_member WHERE group_id::text=$1 AND user_id::text=$2`, m.ClanID, userID)
		s.refreshTrophies(ctx, m.ClanID)
		s.say(ctx, cfg, m.ClanID, "left", userID, "")
		if newLeader != "" {
			s.say(ctx, cfg, m.ClanID, "leader", newLeader, "")
			s.notify(ctx, []string{newLeader}, "clan_role_changed", map[string]any{"clanId": m.ClanID, "role": settings.RoleLeader}, true)
		}
	}
	_, _ = s.db.ExecContext(ctx, `INSERT INTO clan_cooldown (user_id, group_id, reason, created_at) VALUES ($1,$2,'left',now()) ON CONFLICT (user_id, group_id) DO UPDATE SET reason='left', created_at=now()`, userID, m.ClanID)
	return map[string]any{"left": true, "clanId": m.ClanID, "deleted": deleted, "newLeaderId": newLeader}, nil
}

func (s *service) makeLeader(ctx context.Context, clanID, userID string) error {
	var state int
	if err := s.db.QueryRowContext(ctx, `SELECT state FROM group_edge WHERE source_id::text=$1 AND destination_id::text=$2`, clanID, userID).Scan(&state); err != nil {
		return err
	}
	for ; state > 0; state-- {
		if err := s.groups.Promote(ctx, clanID, []string{userID}); err != nil {
			return err
		}
	}
	_, err := s.db.ExecContext(ctx, `UPDATE clan_member SET role='leader' WHERE group_id::text=$1 AND user_id::text=$2`, clanID, userID)
	return err
}

type targetInput struct {
	UserID string `json:"userId"`
	ClanID string `json:"clanId"`
}

func (s *service) target(ctx context.Context, payload string) (string, error) {
	var in targetInput
	if err := apierr.Decode(payload, &in); err != nil {
		return "", err
	}
	id := strings.ToLower(strings.TrimSpace(in.UserID))
	if !uuidPattern.MatchString(id) {
		return "", apierr.BadRequest("INVALID_USER_ID").WithField("userId")
	}
	return id, nil
}

func (s *service) rpcRequests(ctx context.Context, userID, payload string) (any, error) {
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Accept)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT u.id::text, u.username, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), e.update_time, COALESCE(t.trophies,0)
		FROM group_edge e JOIN users u ON u.id=e.destination_id LEFT JOIN trophy_player t ON t.user_id=u.id
		WHERE e.source_id::text=$1 AND e.state=3 ORDER BY e.update_time LIMIT 200`, m.ClanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, uname, dname, avatar string
		var at time.Time
		var t int
		if err := rows.Scan(&id, &uname, &dname, &avatar, &at, &t); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"userId": id, "username": uname, "displayName": dname, "avatarUrl": avatar, "trophies": t, "requestedAt": at.UTC().Format(time.RFC3339)})
	}
	return map[string]any{"clanId": m.ClanID, "items": items}, nil
}

func (s *service) hasRequest(ctx context.Context, clanID, userID string) bool {
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM group_edge WHERE source_id::text=$1 AND destination_id::text=$2 AND state=3)`, clanID, userID).Scan(&ok)
	return ok
}

func (s *service) rpcAccept(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Accept)
	if err != nil {
		return nil, err
	}
	if !s.hasRequest(ctx, m.ClanID, target) {
		return nil, apierr.NotFound("CLAN_REQUEST_NOT_FOUND").WithField("userId")
	}
	c, err := s.clanByID(ctx, s.db, m.ClanID)
	if err != nil {
		return nil, err
	}
	if other, _ := s.memberOf(ctx, s.db, target); other != nil {
		_ = s.groups.Kick(ctx, c.ID, []string{target})
		return nil, apierr.Conflict("CLAN_ALREADY_MEMBER").WithField("userId")
	}
	if err := s.addMember(ctx, cfg, c, target, func() error { return s.groups.Add(ctx, c.ID, []string{target}) }); err != nil {
		return nil, err
	}
	s.say(ctx, cfg, c.ID, "joined", target, userID)
	s.notify(ctx, []string{target}, "clan_request_accepted", map[string]any{"clanId": c.ID, "clanName": c.Name, "byUserId": userID}, true)
	return map[string]any{"accepted": true, "clanId": c.ID, "userId": target}, nil
}

func (s *service) rpcReject(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Accept)
	if err != nil {
		return nil, err
	}
	if !s.hasRequest(ctx, m.ClanID, target) {
		return nil, apierr.NotFound("CLAN_REQUEST_NOT_FOUND").WithField("userId")
	}
	if err := s.groups.Kick(ctx, m.ClanID, []string{target}); err != nil {
		return nil, err
	}
	s.notify(ctx, []string{target}, "clan_request_rejected", map[string]any{"clanId": m.ClanID}, true)
	return map[string]any{"rejected": true, "clanId": m.ClanID, "userId": target}, nil
}

func (s *service) rpcInvite(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Invite)
	if err != nil {
		return nil, err
	}
	if target == userID {
		return nil, apierr.BadRequest("CLAN_CANNOT_INVITE_SELF").WithField("userId")
	}
	var exists bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id::text=$1)`, target).Scan(&exists)
	if !exists {
		return nil, apierr.NotFound("USER_NOT_FOUND").WithField("userId")
	}
	if other, _ := s.memberOf(ctx, s.db, target); other != nil {
		return nil, apierr.Conflict("CLAN_ALREADY_MEMBER").WithField("userId")
	}
	expires := s.now().Add(time.Duration(cfg.Cooldowns.InviteDays) * 24 * time.Hour)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO clan_request (group_id, user_id, kind, by_user, expires_at) VALUES ($1,$2,'invite',$3,$4)
		ON CONFLICT (group_id, user_id, kind) DO UPDATE SET by_user=EXCLUDED.by_user, expires_at=EXCLUDED.expires_at, created_at=now()`, m.ClanID, target, userID, expires); err != nil {
		return nil, err
	}
	c, _ := s.clanByID(ctx, s.db, m.ClanID)
	name := ""
	if c != nil {
		name = c.Name
	}
	s.notify(ctx, []string{target}, "clan_invite", map[string]any{"clanId": m.ClanID, "clanName": name, "byUserId": userID, "expiresAt": expires.UTC().Format(time.RFC3339)}, true)
	return map[string]any{"invited": true, "clanId": m.ClanID, "userId": target, "expiresAt": expires.UTC().Format(time.RFC3339)}, nil
}

func (s *service) rpcInvites(ctx context.Context, userID, payload string) (any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.group_id::text, COALESCE(r.by_user::text,''), r.created_at, r.expires_at FROM clan_request r
		WHERE r.user_id::text=$1 AND r.kind='invite' AND (r.expires_at IS NULL OR r.expires_at > now()) ORDER BY r.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	type inv struct {
		id, by  string
		at      time.Time
		expires *time.Time
	}
	list := []inv{}
	for rows.Next() {
		var x inv
		if err := rows.Scan(&x.id, &x.by, &x.at, &x.expires); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, x)
	}
	rows.Close()
	items := []map[string]any{}
	for _, x := range list {
		c, err := s.clanByID(ctx, s.db, x.id)
		if err != nil {
			continue
		}
		var exp any
		if x.expires != nil {
			exp = x.expires.UTC().Format(time.RFC3339)
		}
		items = append(items, map[string]any{"clan": s.clanView(ctx, c, userID, ""), "byUserId": x.by, "invitedAt": x.at.UTC().Format(time.RFC3339), "expiresAt": exp})
	}
	return map[string]any{"items": items}, nil
}

func (s *service) rpcInviteAccept(ctx context.Context, userID, payload string) (any, error) {
	var in targetInput
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if err := s.enabled(cfg); err != nil {
		return nil, err
	}
	c, err := s.clanByID(ctx, s.db, in.ClanID)
	if err != nil {
		return nil, err
	}
	var invited bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_request WHERE group_id=$1 AND user_id=$2 AND kind='invite' AND (expires_at IS NULL OR expires_at > now()))`, c.ID, userID).Scan(&invited)
	if !invited {
		return nil, apierr.NotFound("CLAN_INVITE_NOT_FOUND").WithField("clanId")
	}
	if m, _ := s.memberOf(ctx, s.db, userID); m != nil {
		return nil, apierr.Conflict("CLAN_ALREADY_MEMBER").With("clanId", m.ClanID)
	}
	if err := s.cooldown(ctx, cfg, userID, c.ID); err != nil {
		return nil, err
	}
	return s.joinInvited(ctx, cfg, c, userID)
}

func (s *service) rpcInviteDecline(ctx context.Context, userID, payload string) (any, error) {
	var in targetInput
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM clan_request WHERE group_id::text=$1 AND user_id::text=$2 AND kind='invite'`, strings.ToLower(strings.TrimSpace(in.ClanID)), userID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, apierr.NotFound("CLAN_INVITE_NOT_FOUND").WithField("clanId")
	}
	return map[string]any{"declined": true, "clanId": in.ClanID}, nil
}

func (s *service) peer(ctx context.Context, actor *memberRow, target string) (*memberRow, error) {
	if target == "" {
		return nil, apierr.BadRequest("INVALID_USER_ID").WithField("userId")
	}
	t, err := s.memberOf(ctx, s.db, target)
	if err != nil {
		return nil, err
	}
	if t == nil || t.ClanID != actor.ClanID {
		return nil, apierr.NotFound("CLAN_MEMBER_NOT_FOUND").WithField("userId")
	}
	return t, nil
}

func (s *service) rpcKick(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Kick)
	if err != nil {
		return nil, err
	}
	if target == userID {
		return nil, apierr.BadRequest("CLAN_CANNOT_KICK_SELF").WithField("userId")
	}
	t, err := s.peer(ctx, m, target)
	if err != nil {
		return nil, err
	}
	if !outranks(m.Role, t.Role) {
		return nil, apierr.Forbidden("CLAN_ROLE_TOO_LOW").With("role", t.Role)
	}
	if err := s.groups.Kick(ctx, m.ClanID, []string{target}); err != nil {
		return nil, err
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_member WHERE group_id::text=$1 AND user_id::text=$2`, m.ClanID, target)
	_, _ = s.db.ExecContext(ctx, `INSERT INTO clan_cooldown (user_id, group_id, reason, created_at) VALUES ($1,$2,'kicked',now()) ON CONFLICT (user_id, group_id) DO UPDATE SET reason='kicked', created_at=now()`, target, m.ClanID)
	s.refreshTrophies(ctx, m.ClanID)
	s.say(ctx, cfg, m.ClanID, "kicked", target, userID)
	s.notify(ctx, []string{target}, "clan_kicked", map[string]any{"clanId": m.ClanID, "byUserId": userID}, true)
	return map[string]any{"kicked": true, "clanId": m.ClanID, "userId": target}, nil
}

func (s *service) rpcPromote(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Promote)
	if err != nil {
		return nil, err
	}
	t, err := s.peer(ctx, m, target)
	if err != nil {
		return nil, err
	}
	to, ok := canPromote(m.Role, t.Role)
	if !ok {
		return nil, apierr.Forbidden("CLAN_ROLE_TOO_LOW").With("role", t.Role)
	}
	if to == settings.RoleCoLeader {
		var n int
		_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan_member WHERE group_id::text=$1 AND role='co_leader'`, m.ClanID).Scan(&n)
		if n >= cfg.Roles.MaxCoLeaders {
			return nil, apierr.Conflict("CLAN_TOO_MANY_CO_LEADERS").With("max", cfg.Roles.MaxCoLeaders)
		}
		if err := s.groups.Promote(ctx, m.ClanID, []string{target}); err != nil {
			return nil, err
		}
	}
	return s.setRole(ctx, cfg, m, target, t.Role, to)
}

func (s *service) rpcDemote(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, cfg.Roles.Promote)
	if err != nil {
		return nil, err
	}
	t, err := s.peer(ctx, m, target)
	if err != nil {
		return nil, err
	}
	to := demoteTo(t.Role)
	if to == "" || !outranks(m.Role, t.Role) {
		return nil, apierr.Forbidden("CLAN_ROLE_TOO_LOW").With("role", t.Role)
	}
	if t.Role == settings.RoleCoLeader {
		if err := s.groups.Demote(ctx, m.ClanID, []string{target}); err != nil {
			return nil, err
		}
	}
	return s.setRole(ctx, cfg, m, target, t.Role, to)
}

func (s *service) setRole(ctx context.Context, cfg settings.All, m *memberRow, target, from, to string) (any, error) {
	if _, err := s.db.ExecContext(ctx, `UPDATE clan_member SET role=$3 WHERE group_id::text=$1 AND user_id::text=$2`, m.ClanID, target, to); err != nil {
		return nil, err
	}
	s.say(ctx, cfg, m.ClanID, "role", target, to)
	s.notify(ctx, []string{target}, "clan_role_changed", map[string]any{"clanId": m.ClanID, "role": to, "from": from}, true)
	return map[string]any{"clanId": m.ClanID, "userId": target, "role": to, "from": from}, nil
}

func (s *service) rpcTransfer(ctx context.Context, userID, payload string) (any, error) {
	target, err := s.target(ctx, payload)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	m, err := s.requireRole(ctx, userID, []string{settings.RoleLeader})
	if err != nil {
		return nil, err
	}
	if target == userID {
		return nil, apierr.BadRequest("CLAN_ALREADY_LEADER").WithField("userId")
	}
	t, err := s.peer(ctx, m, target)
	if err != nil {
		return nil, err
	}
	if err := s.makeLeader(ctx, m.ClanID, target); err != nil {
		return nil, err
	}
	if err := s.groups.Demote(ctx, m.ClanID, []string{userID}); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE clan_member SET role='co_leader' WHERE group_id::text=$1 AND user_id::text=$2`, m.ClanID, userID); err != nil {
		return nil, err
	}
	s.say(ctx, cfg, m.ClanID, "leader", target, userID)
	s.notify(ctx, []string{target}, "clan_role_changed", map[string]any{"clanId": m.ClanID, "role": settings.RoleLeader, "from": t.Role}, true)
	return map[string]any{"clanId": m.ClanID, "leaderId": target, "role": settings.RoleCoLeader}, nil
}

func (s *service) say(ctx context.Context, cfg settings.All, clanID, kind, userID, extra string) {
	if !cfg.Chat.SystemMessages {
		return
	}
	content := map[string]any{"type": "clan_system", "event": kind, "userId": userID, "username": s.username(ctx, userID)}
	if extra != "" {
		content["extra"] = extra
	}
	if err := s.groups.Say(ctx, clanID, content); err != nil && s.logger != nil {
		s.logger.Debug("Clan %s system message not sent: %v", clanID, err)
	}
}

func (s *service) refreshTrophies(ctx context.Context, clanID string) {
	var total int64
	if err := s.db.QueryRowContext(ctx, `UPDATE clan SET trophies=(SELECT COALESCE(SUM(p.trophies),0) FROM clan_member m JOIN trophy_player p ON p.user_id=m.user_id WHERE m.group_id=clan.group_id)
		WHERE group_id::text=$1 RETURNING trophies`, clanID).Scan(&total); err != nil {
		return
	}
	cfg := s.settings(ctx)
	if cfg.Events.Publish {
		gameevents.Publish(ctx, gameevents.Event{ID: owner + ":t:" + uuid.Must(uuid.NewV4()).String(), UserID: clanID, Name: "clan_trophies", Source: owner,
			Fields: map[string]any{"trophies": float64(total)}})
	}
}
