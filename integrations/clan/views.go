package main

import (
	"context"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

type texts struct {
	langs *localized.Set
	lang  string
}

func (s *service) texts(ctx context.Context, userID, lang string) texts {
	langs, l, _ := localized.ForUser(ctx, s.db, userID, lang)
	return texts{langs: langs, lang: l}
}

func (t texts) of(m map[string]string, fallback string) string {
	if t.langs != nil {
		if v := t.langs.Resolve(localized.Text(m), t.lang); v != "" {
			return v
		}
	}
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return fallback
}

func (s *service) clanView(ctx context.Context, c *clanRow, userID, role string) map[string]any {
	if c == nil {
		return nil
	}
	r := s.road(ctx)
	next := r.next(c.Level)
	var nextXP any
	if next != nil {
		nextXP = next.XP
	}
	var curXP int64
	if cur := r.get(c.Level); cur != nil {
		curXP = cur.XP
	}
	out := map[string]any{"clanId": c.ID, "name": c.Name, "description": c.Description, "avatarUrl": c.AvatarURL, "lang": c.Lang, "tag": c.Tag, "badge": c.Badge,
		"joinType": c.JoinType, "minTrophies": c.MinTrophies, "members": c.Members, "maxMembers": c.MaxMembers, "level": c.Level, "xp": c.XP, "levelXp": curXP, "nextLevelXp": nextXP,
		"trophies": c.Trophies, "warWins": c.WarWins, "warLosses": c.WarLosses, "warDraws": c.WarDraws, "warPoints": c.WarPoints, "createdAt": c.CreatedAt.UTC().Format(time.RFC3339),
		"channel": "3." + c.ID}
	if role != "" {
		out["myRole"] = role
	}
	return out
}

func (s *service) members(ctx context.Context, clanID string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.user_id::text, COALESCE(u.username,''), COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), m.role, m.joined_at, m.xp, m.war_points, m.wars,
		COALESCE(t.trophies,0), COALESCE(t.league_key,'')
		FROM clan_member m LEFT JOIN users u ON u.id=m.user_id LEFT JOIN trophy_player t ON t.user_id=m.user_id
		WHERE m.group_id::text=$1 ORDER BY CASE m.role WHEN 'leader' THEN 0 WHEN 'co_leader' THEN 1 WHEN 'elder' THEN 2 ELSE 3 END, COALESCE(t.trophies,0) DESC, m.joined_at`, clanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, uname, dname, avatar, role, league string
		var joined time.Time
		var xp, wp int64
		var wars, trophies int
		if err := rows.Scan(&id, &uname, &dname, &avatar, &role, &joined, &xp, &wp, &wars, &trophies, &league); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"userId": id, "username": uname, "displayName": dname, "avatarUrl": avatar, "role": role, "joinedAt": joined.UTC().Format(time.RFC3339),
			"xp": xp, "warPoints": wp, "wars": wars, "trophies": trophies, "league": league})
	}
	return out, rows.Err()
}

func (s *service) rpcGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ClanID string `json:"clanId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	c, err := s.clanByID(ctx, s.db, in.ClanID)
	if err != nil {
		return nil, err
	}
	role := ""
	if m, _ := s.memberOf(ctx, s.db, userID); m != nil && m.ClanID == c.ID {
		role = m.Role
	}
	members, err := s.members(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	var requested, invited bool
	if role == "" {
		requested = s.hasRequest(ctx, c.ID, userID)
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_request WHERE group_id::text=$1 AND user_id::text=$2 AND kind='invite' AND (expires_at IS NULL OR expires_at > now()))`, c.ID, userID).Scan(&invited)
	}
	return map[string]any{"clan": s.clanView(ctx, c, userID, role), "members": members, "requested": requested, "invited": invited}, nil
}

func (s *service) rpcMine(ctx context.Context, userID, payload string) (any, error) {
	cfg := s.settings(ctx)
	var invites int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan_request WHERE user_id::text=$1 AND kind='invite' AND (expires_at IS NULL OR expires_at > now())`, userID).Scan(&invites)
	var requested []string
	rows, err := s.db.QueryContext(ctx, `SELECT e.source_id::text FROM group_edge e JOIN clan c ON c.group_id=e.source_id WHERE e.destination_id::text=$1 AND e.state=3`, userID)
	if err == nil {
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				requested = append(requested, id)
			}
		}
		rows.Close()
	}
	if requested == nil {
		requested = []string{}
	}
	out := map[string]any{"enabled": cfg.Enabled, "clan": nil, "role": nil, "pendingRequests": 0, "invites": invites, "requestedClanIds": requested,
		"permissions": map[string]bool{}, "cooldownSeconds": 0}
	if err := s.cooldown(ctx, cfg, userID, ""); err != nil {
		if e, ok := apierr.As(err); ok {
			n, _ := e.Meta["retryAfter"].(int)
			out["cooldownSeconds"] = n
		}
	}
	m, err := s.memberOf(ctx, s.db, userID)
	if err != nil || m == nil {
		return out, err
	}
	c, err := s.clanByID(ctx, s.db, m.ClanID)
	if err != nil {
		return nil, err
	}
	out["clan"], out["role"] = s.clanView(ctx, c, userID, m.Role), m.Role
	out["permissions"] = map[string]bool{"accept": allowed(cfg.Roles.Accept, m.Role), "invite": allowed(cfg.Roles.Invite, m.Role), "kick": allowed(cfg.Roles.Kick, m.Role),
		"promote": allowed(cfg.Roles.Promote, m.Role), "edit": allowed(cfg.Roles.Edit, m.Role), "war": allowed(cfg.Roles.War, m.Role), "transfer": m.Role == settings.RoleLeader}
	if allowed(cfg.Roles.Accept, m.Role) {
		var n int
		_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM group_edge WHERE source_id::text=$1 AND state=3`, m.ClanID).Scan(&n)
		out["pendingRequests"] = n
	}
	return out, nil
}

func (s *service) rpcSearch(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Query      string `json:"q"`
		JoinType   string `json:"joinType"`
		Joinable   bool   `json:"joinable"`
		MinMembers int    `json:"minMembers"`
		Sort       string `json:"sort"`
		Limit      int    `json:"limit"`
		Offset     int    `json:"offset"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	limit := in.Limit
	if limit <= 0 || limit > cfg.Display.SearchLimit {
		limit = cfg.Display.SearchLimit
	}
	if in.Offset < 0 || in.Offset > 10000 {
		return nil, apierr.BadRequest("INVALID_OFFSET").WithField("offset")
	}
	args := &pluginkit.Args{}
	where := []string{"true"}
	q := strings.TrimSpace(in.Query)
	if q != "" {
		p := args.Add(pluginkit.LikePattern(q))
		where = append(where, "(g.name ILIKE "+p+" OR c.tag ILIKE "+args.Add(strings.ToUpper(strings.TrimPrefix(q, "#")))+")")
	}
	switch in.JoinType {
	case "":
	case settings.JoinOpen, settings.JoinClosed, settings.JoinInvite:
		where = append(where, "c.join_type = "+args.Add(in.JoinType))
	default:
		return nil, apierr.BadRequest("CLAN_JOIN_TYPE_INVALID").WithField("joinType")
	}
	if in.Joinable {
		where = append(where, "c.join_type <> 'invite' AND g.edge_count < g.max_count AND c.min_trophies <= "+args.Add(s.trophiesOf(ctx, userID)))
	}
	if in.MinMembers > 0 {
		where = append(where, "g.edge_count >= "+args.Add(in.MinMembers))
	}
	order := "c.trophies DESC, g.edge_count DESC"
	switch in.Sort {
	case "", "trophies":
	case "level":
		order = "c.level DESC, c.xp DESC"
	case "members":
		order = "g.edge_count DESC, c.trophies DESC"
	case "new":
		order = "c.created_at DESC"
	default:
		return nil, apierr.BadRequest("CLAN_SORT_INVALID").WithField("sort")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+clanCols+` FROM clan c JOIN groups g ON g.id=c.group_id WHERE `+strings.Join(where, " AND ")+` ORDER BY `+order+`, c.group_id LIMIT `+args.Add(limit+1)+` OFFSET `+args.Add(in.Offset), args.List...)
	if err != nil {
		return nil, err
	}
	list := []*clanRow{}
	for rows.Next() {
		c, err := scanClan(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, c)
	}
	rows.Close()
	more := len(list) > limit
	if more {
		list = list[:limit]
	}
	items := make([]map[string]any, 0, len(list))
	for _, c := range list {
		items = append(items, s.clanView(ctx, c, userID, ""))
	}
	return map[string]any{"items": items, "more": more, "nextOffset": in.Offset + len(items)}, nil
}

func (s *service) rpcLevels(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	tx := s.texts(ctx, userID, in.Lang)
	r := s.road(ctx)
	cur, xp := 0, int64(0)
	clanID := ""
	if m, _ := s.memberOf(ctx, s.db, userID); m != nil {
		clanID = m.ClanID
		_ = s.db.QueryRowContext(ctx, `SELECT level, xp FROM clan WHERE group_id::text=$1`, m.ClanID).Scan(&cur, &xp)
	}
	items := make([]map[string]any, 0, len(r))
	for _, l := range r {
		items = append(items, map[string]any{"level": l.Level, "xp": l.XP, "name": tx.of(l.Name, ""), "iconUrl": l.IconURL, "memberCap": l.MemberCap, "prizeKey": l.PrizeKey,
			"perks": l.Perks, "reached": clanID != "" && l.Level <= cur})
	}
	cfg := s.settings(ctx)
	sources := make([]map[string]any, 0, len(cfg.XP.Sources))
	for _, src := range cfg.XP.Sources {
		sources = append(sources, map[string]any{"type": src.Type, "event": src.Event, "amount": src.Amount, "perValue": src.PerValue})
	}
	var clan any
	if clanID != "" {
		clan = map[string]any{"clanId": clanID, "level": cur, "xp": xp}
	}
	return map[string]any{"levels": items, "clan": clan, "sources": sources, "war": map[string]any{"win": cfg.XP.WarWin, "draw": cfg.XP.WarDraw, "loss": cfg.XP.WarLoss},
		"memberDailyCap": cfg.XP.MemberDailyCap}, nil
}

func (s *service) rpcXPHistory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit int `json:"limit"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	m, err := s.requireMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	limit := in.Limit
	if limit <= 0 || limit > cfg.Display.HistoryLimit {
		limit = cfg.Display.HistoryLimit
	}
	rows, err := s.db.QueryContext(ctx, `SELECT l.source, l.amount, COALESCE(l.user_id::text,''), COALESCE(u.username,''), l.created_at FROM clan_xp_log l LEFT JOIN users u ON u.id=l.user_id
		WHERE l.group_id::text=$1 ORDER BY l.id DESC LIMIT $2`, m.ClanID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var src, uid, uname string
		var amount int64
		var at time.Time
		if err := rows.Scan(&src, &amount, &uid, &uname, &at); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"source": src, "amount": amount, "userId": uid, "username": uname, "createdAt": at.UTC().Format(time.RFC3339)})
	}
	return map[string]any{"clanId": m.ClanID, "items": items}, nil
}
