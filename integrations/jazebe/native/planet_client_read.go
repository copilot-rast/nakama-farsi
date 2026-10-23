package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Main app hub read models (source planet-overview.service.ts): the planet
// home (GET /planet), the constellation (GET /brands), the three Home-hub
// previews and the Sky discovery list. Read-only.

const (
	planetPreviewLimit  = 5
	planetDiscoverLimit = 50
	defaultBrandColor   = "#FFFFFF"
)

type clientPlanet struct {
	ID, Slug, Name, PointName, Status       string
	LatinName, ShortDescription, BrandColor sql.NullString
	Glyph, Icon, Logo, Cover, CategoryID    sql.NullString
	CategoryKey, CategoryTitle              sql.NullString
	Fans                                    int
	BrandActive                             bool
}

const clientPlanetColumns = `p.id,p.slug,p.name,p.point_name,p.status,p.latin_name,p.short_description,p.brand_color,p.glyph,
	p.icon_asset,p.logo_asset,p.cover_asset,p.category_id::text,c.key,c.title_fa,p.fans_count,b.is_active`

const clientPlanetFrom = ` FROM jazebe_planet p JOIN jazebe_brand b ON b.id=p.brand_id LEFT JOIN jazebe_planet_category c ON c.id=p.category_id`

func scanClientPlanet(row interface{ Scan(...any) error }) (clientPlanet, error) {
	var p clientPlanet
	err := row.Scan(&p.ID, &p.Slug, &p.Name, &p.PointName, &p.Status, &p.LatinName, &p.ShortDescription, &p.BrandColor, &p.Glyph,
		&p.Icon, &p.Logo, &p.Cover, &p.CategoryID, &p.CategoryKey, &p.CategoryTitle, &p.Fans, &p.BrandActive)
	return p, err
}

func (p clientPlanet) color() string {
	if p.BrandColor.Valid && p.BrandColor.String != "" {
		return p.BrandColor.String
	}
	return defaultBrandColor
}

func (p clientPlanet) latinOrSlug() string {
	if p.LatinName.Valid && p.LatinName.String != "" {
		return p.LatinName.String
	}
	return p.Slug
}

// glyph is the planet's glyph or the upper-cased first letter of its Latin
// (else Persian) name.
func (p clientPlanet) glyph() string {
	if p.Glyph.Valid && p.Glyph.String != "" {
		return p.Glyph.String
	}
	name := p.Name
	if p.LatinName.Valid && p.LatinName.String != "" {
		name = p.LatinName.String
	}
	r, _ := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError {
		return ""
	}
	return string(unicode.ToUpper(r))
}

func nullable(s sql.NullString) any {
	if s.Valid {
		return s.String
	}
	return nil
}

// availablePlanet enforces the client rule: unknown is not found, a planet
// that is not active (or whose brand is inactive) is unavailable.
func availablePlanet(p clientPlanet, err error) (clientPlanet, error) {
	if errors.Is(err, sql.ErrNoRows) {
		return p, status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return p, status.Error(codes.Internal, "planet unavailable")
	}
	if p.Status != "active" || !p.BrandActive {
		return p, status.Error(codes.PermissionDenied, "planet unavailable")
	}
	return p, nil
}

// planetByRef resolves an id, a slug or a Latin name (case-insensitive), as
// the Home-hub routes do.
func planetByRef(ctx context.Context, db *sql.DB, ref string) (clientPlanet, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" || len(ref) > 128 {
		return clientPlanet{}, status.Error(codes.InvalidArgument, "planet reference required")
	}
	if id, err := requiredUUID(ref); err == nil {
		return availablePlanet(scanClientPlanet(db.QueryRowContext(ctx, `SELECT `+clientPlanetColumns+clientPlanetFrom+` WHERE p.id=$1`, id)))
	}
	return availablePlanet(scanClientPlanet(db.QueryRowContext(ctx, `SELECT `+clientPlanetColumns+clientPlanetFrom+`
		WHERE lower(p.slug)=lower($1) OR lower(p.latin_name)=lower($1) ORDER BY p.created_at ASC LIMIT 1`, ref)))
}

func faGrouped(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 && r != '-' {
			b.WriteRune('٬')
		}
		b.WriteString(faDigits(string(r)))
	}
	return b.String()
}

func faDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune('۰' + (r - '0'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func tehran() *time.Location {
	if loc, err := time.LoadLocation("Asia/Tehran"); err == nil {
		return loc
	}
	return time.FixedZone("IR", 3*3600+1800)
}

// missionPeriodLabel is the mission "period" chip (source periodLabel).
func missionPeriodLabel(ends, now time.Time) string {
	days := int(math.Ceil(ends.Sub(now).Hours() / 24))
	switch {
	case days <= 0:
		return "امروز"
	case days == 1:
		return "۱ روز"
	case days < 7:
		return faNumber(days) + " روز"
	case days < 31:
		return "این ماه"
	}
	return faNumber(int(math.Round(float64(days)/30))) + " ماه"
}

// raffleEndsLabel is the raffle card's end line on the Tehran clock (source endsLabel).
func raffleEndsLabel(draw, now time.Time) string {
	loc := tehran()
	d, n := draw.In(loc), now.In(loc)
	clock := faDigits(d.Format("15:04"))
	sameDay := func(a, b time.Time) bool { return a.Year() == b.Year() && a.YearDay() == b.YearDay() }
	switch {
	case sameDay(d, n):
		return "پایان: امروز " + clock
	case sameDay(d, n.Add(24*time.Hour)):
		return "پایان: فردا " + clock
	}
	_, jm, jd := gregorianToJalali(d.Year(), int(d.Month()), d.Day())
	return "پایان: " + faNumber(jd) + " " + jalaliMonthNames[jm-1] + " " + clock
}

// planetLevelProgress mirrors the source levelProgress: the reached level
// (never below the stored one), names and the 0..1 progress to the next.
func planetLevelProgress(raw []byte, xp int64, stored int) map[string]any {
	names := []string{"تازه‌وارد", "عضو فعال", "کهنه‌کار"}
	ladder := []profileLevel{{1, "newcomer", names[0], 0}, {2, "active", names[1], 0}, {3, "veteran", names[2], 0}}
	var configured []profileLevel
	if json.Unmarshal(raw, &configured) == nil && len(configured) == 3 && configured[0].Level == 1 && configured[1].Level == 2 && configured[2].Level == 3 {
		ladder = configured
	}
	reached := 0
	for i := 1; i < len(ladder); i++ {
		if ladder[i].MinXp <= ladder[i-1].MinXp || xp < ladder[i].MinXp {
			break
		}
		reached = i
	}
	level := ladder[reached].Level
	if stored > level {
		level = stored
	}
	title := func(i int) string {
		if ladder[i].TitleFa != "" {
			return ladder[i].TitleFa
		}
		return names[i]
	}
	progress, nextName := 1.0, title(reached)
	if reached+1 < len(ladder) && ladder[reached+1].MinXp > ladder[reached].MinXp {
		span := float64(ladder[reached+1].MinXp - ladder[reached].MinXp)
		progress = math.Min(1, math.Max(0, float64(xp-ladder[reached].MinXp)/span))
		nextName = title(reached + 1)
	}
	return map[string]any{"level": level, "levelMax": len(ladder), "levelName": title(level - 1), "nextLevelName": nextName, "progress": progress}
}

type clientLottery struct {
	ID, Title, ShortDescription string
	Draw                        time.Time
	Price, Total                int
}

func nearestLottery(ctx context.Context, db *sql.DB, planetID string, now time.Time) (*clientLottery, error) {
	var l clientLottery
	var short sql.NullString
	err := db.QueryRowContext(ctx, `SELECT id,title,short_description,draw_at,ticket_price_points,total_ticket_count FROM jazebe_lottery
		WHERE planet_id=$1 AND status='active' AND draw_at>$2 ORDER BY draw_at ASC LIMIT 1`, planetID, now).Scan(&l.ID, &l.Title, &short, &l.Draw, &l.Price, &l.Total)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery unavailable")
	}
	l.ShortDescription = short.String
	return &l, nil
}

func topPrizeTitle(ctx context.Context, db *sql.DB, lotteryID string) (string, bool, error) {
	var title string
	err := db.QueryRowContext(ctx, `SELECT title FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority ASC LIMIT 1`, lotteryID).Scan(&title)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, status.Error(codes.Internal, "lottery prize unavailable")
	}
	return title, true, nil
}

func planetOverview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	planetID := in.PlanetID
	if planetID == "" {
		// Source resolution order without a host or header: latest membership.
		err := db.QueryRowContext(ctx, `SELECT planet_id FROM jazebe_planet_membership WHERE user_id=$1
			ORDER BY last_activity_at DESC NULLS LAST, joined_at DESC LIMIT 1`, user).Scan(&planetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "planet not found")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "planet unavailable")
		}
	}
	if planetID, err = requiredUUID(planetID); err != nil {
		return "", err
	}
	p, err := availablePlanet(scanClientPlanet(db.QueryRowContext(ctx, `SELECT `+clientPlanetColumns+clientPlanetFrom+` WHERE p.id=$1`, planetID)))
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	var (
		member, hasActivity        bool
		joinPath                   sql.NullString
		balance, xp                int64
		storedLevel, activeMission int
		activeGroups               int
		thresholds                 []byte
		groupID                    sql.NullString
	)
	err = db.QueryRowContext(ctx, `SELECT
		COALESCE((SELECT is_active FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2),false),
		(SELECT join_path FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2),
		COALESCE((SELECT balance FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2),0),
		COALESCE((SELECT total_xp FROM jazebe_user_planet_stats WHERE planet_id=$1 AND user_id=$2),0),
		COALESCE((SELECT level FROM jazebe_user_planet_stats WHERE planet_id=$1 AND user_id=$2),1),
		COALESCE((SELECT missions_completed>0 OR games_played>0 OR tickets_created>0 FROM jazebe_user_planet_stats WHERE planet_id=$1 AND user_id=$2),false),
		(SELECT count(*) FROM jazebe_mission WHERE planet_id=$1 AND status='active' AND starts_at<=$3 AND ends_at>=$3),
		(SELECT count(*) FROM jazebe_group WHERE planet_id=$1 AND status='active'),
		(SELECT xp_level_thresholds FROM jazebe_planet_settings WHERE planet_id=$1),
		(SELECT group_id::text FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active')`, planetID, user, now).
		Scan(&member, &joinPath, &balance, &xp, &storedLevel, &hasActivity, &activeMission, &activeGroups, &thresholds, &groupID)
	if err != nil {
		return "", status.Error(codes.Internal, "planet overview unavailable")
	}
	unreadNotifications, err := notificationUnread(ctx, db, user, planetID)
	if err != nil {
		return "", err
	}
	unreadMessages := 0
	if groupID.Valid {
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_group_message m WHERE m.group_id=$1 AND m.deleted_at IS NULL
			AND m.seq>COALESCE((SELECT last_read_seq FROM jazebe_group_message_read WHERE group_id=$1 AND user_id=$2),0)
			AND (m.sender_user_id IS NULL OR m.sender_user_id<>$2)`, groupID.String, user).Scan(&unreadMessages); err != nil {
			return "", status.Error(codes.Internal, "group messages unavailable")
		}
	}
	lottery, err := nearestLottery(ctx, db, planetID, now)
	if err != nil {
		return "", err
	}
	var teaser any
	if lottery != nil {
		prize, ok, err := topPrizeTitle(ctx, db, lottery.ID)
		if err != nil {
			return "", err
		}
		if !ok {
			prize = lottery.ShortDescription
		}
		teaser = map[string]any{"id": lottery.ID, "title": lottery.Title, "prize": prize, "endsInSeconds": int64(math.Max(0, math.Floor(lottery.Draw.Sub(now).Seconds())))}
	}
	stats := planetLevelProgress(thresholds, xp, storedLevel)
	stats["points"] = balance
	missions := map[string]any{"sub": "ماموریت فعالی نیست"}
	if activeMission > 0 {
		missions = map[string]any{"sub": faNumber(activeMission) + " ماموریت تازه", "badge": faNumber(activeMission) + " تازه", "hot": true}
	}
	groups := map[string]any{"sub": faNumber(activeGroups) + " گروه فعال"}
	if unreadMessages > 0 {
		groups["badge"] = faNumber(unreadMessages) + " پیام"
		groups["hot"] = true
	}
	raffle := map[string]any{"sub": "قرعه‌کشی فعالی نیست"}
	if lottery != nil {
		raffle = map[string]any{"sub": "فعال", "badge": "فعال", "hot": true}
	}
	notifications := map[string]any{"sub": "اعلان‌ها", "hot": unreadNotifications > 0}
	if unreadNotifications > 0 {
		notifications["sub"] = "اعلان‌های تازه"
	}
	category := ""
	if p.CategoryTitle.Valid {
		category = p.CategoryTitle.String
	}
	return encode(map[string]any{
		"brand": map[string]any{"id": p.ID, "name": p.latinOrSlug(), "fa": p.Name, "category": category, "pointName": p.PointName,
			"fans": p.Fans, "color": p.color(), "glyph": p.glyph(), "icon": nullable(p.Icon), "logo": nullable(p.Logo),
			"cover": nullable(p.Cover), "shortDescription": nullable(p.ShortDescription)},
		"stats":        stats,
		"raffleTeaser": teaser,
		"sections": map[string]any{"missions": missions, "games": map[string]any{"sub": "هر روز " + p.PointName + " بگیر", "badge": "روزانه"},
			"wallet": map[string]any{"sub": "موجودی " + p.PointName}, "store": map[string]any{"sub": "جوایز"}, "groups": groups,
			"raffle": raffle, "profile": map[string]any{"sub": "حساب و دستاوردها"}, "notifications": notifications},
		"unreadNotifications": unreadNotifications,
		"viewer":              map[string]any{"member": member, "joinPath": nullable(joinPath), "hasActivity": hasActivity},
	})
}

// planetConstellation is the caller's "منظومه": joined active planets in the
// manual order, then recent activity, then join time.
func planetConstellation(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT `+clientPlanetColumns+clientPlanetFrom+`
		JOIN jazebe_planet_membership m ON m.planet_id=p.id
		WHERE m.user_id=$1 AND m.is_active AND p.status='active' AND b.is_active
		ORDER BY m.sort_order ASC NULLS LAST, m.last_activity_at DESC NULLS LAST, m.joined_at DESC`, user)
	if err != nil {
		return "", status.Error(codes.Internal, "constellation unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		p, err := scanClientPlanet(rows)
		if err != nil {
			return "", status.Error(codes.Internal, "constellation unavailable")
		}
		logo := p.Logo.String
		if logo == "" {
			logo = p.Icon.String
		}
		items = append(items, map[string]any{"id": p.ID, "slug": p.Slug, "name": p.latinOrSlug(), "fa": p.Name,
			"category": p.CategoryTitle.String, "color": p.color(), "glyph": p.glyph(), "logo": logo, "fans": p.Fans, "pointName": p.PointName})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "constellation unavailable")
	}
	return encode(map[string]any{"items": items})
}

func planetRefPayload(payload string) (string, error) {
	var in struct {
		Planet string `json:"planet"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	return in.Planet, nil
}

func planetMissionsPreview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := planetRefPayload(payload)
	if err != nil {
		return "", err
	}
	p, err := planetByRef(ctx, db, ref)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	rows, err := db.QueryContext(ctx, `SELECT m.id,m.title,COALESCE(m.short_description,''),m.reward_points,m.ends_at,GREATEST(1,COALESCE(m.group_target_value,1)),
		(SELECT count(*) FROM jazebe_mission_completion c WHERE c.mission_id=m.id AND c.planet_id=m.planet_id AND c.user_id=$2 AND c.status IN ('approved','recorded'))
		FROM jazebe_mission m WHERE m.planet_id=$1 AND m.status='active' AND m.starts_at<=$3 AND m.ends_at>=$3
		ORDER BY m.sort_order ASC, m.ends_at ASC LIMIT $4`, p.ID, user, now, planetPreviewLimit)
	if err != nil {
		return "", status.Error(codes.Internal, "missions unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, title, desc string
		var points, total, done int
		var ends time.Time
		if err := rows.Scan(&id, &title, &desc, &points, &ends, &total, &done); err != nil {
			return "", status.Error(codes.Internal, "missions unavailable")
		}
		items = append(items, map[string]any{"id": id, "title": title, "desc": desc, "points": points,
			"period": missionPeriodLabel(ends, now), "progress": min(total, done), "total": total})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "missions unavailable")
	}
	return encode(map[string]any{"items": items})
}

func planetGroupsPreview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := planetRefPayload(payload)
	if err != nil {
		return "", err
	}
	p, err := planetByRef(ctx, db, ref)
	if err != nil {
		return "", err
	}
	lottery, err := nearestLottery(ctx, db, p.ID, time.Now().UTC())
	if err != nil {
		return "", err
	}
	var lotteryID any
	if lottery != nil {
		lotteryID = lottery.ID
	}
	rows, err := db.QueryContext(ctx, `SELECT g.id,g.name,g.member_count,COALESCE(s.ticket_count,0),m.role
		FROM jazebe_group g LEFT JOIN jazebe_lottery_group_stat s ON s.group_id=g.id AND s.lottery_id=$3::uuid
		LEFT JOIN jazebe_group_membership m ON m.group_id=g.id AND m.user_id=$2 AND m.status='active'
		WHERE g.planet_id=$1 AND g.status='active' ORDER BY g.member_count DESC, g.created_at ASC LIMIT $4`, p.ID, user, lotteryID, planetPreviewLimit)
	if err != nil {
		return "", status.Error(codes.Internal, "groups unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name string
		var members int
		var tickets int64
		var role sql.NullString
		if err := rows.Scan(&id, &name, &members, &tickets, &role); err != nil {
			return "", status.Error(codes.Internal, "groups unavailable")
		}
		var meta any
		switch role.String {
		case "creator":
			meta = "کاپیتان"
		case "admin":
			meta = "مدیر"
		}
		cta := "پیوستن"
		if role.Valid {
			cta = "باز کردن"
		}
		items = append(items, map[string]any{"id": id, "title": name, "members": members, "tokens": faGrouped(tickets), "meta": meta, "cta": cta, "joined": role.Valid})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "groups unavailable")
	}
	return encode(map[string]any{"items": items})
}

func planetRafflePreview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := planetRefPayload(payload)
	if err != nil {
		return "", err
	}
	p, err := planetByRef(ctx, db, ref)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	lottery, err := nearestLottery(ctx, db, p.ID, now)
	if err != nil || lottery == nil {
		if err == nil {
			return encode(map[string]any{"raffle": nil})
		}
		return "", err
	}
	prize, ok, err := topPrizeTitle(ctx, db, lottery.ID)
	if err != nil {
		return "", err
	}
	if !ok {
		prize = lottery.Title
	}
	var mine int64
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND creator_user_id=$2 AND status='active'`, lottery.ID, user).Scan(&mine); err != nil {
		return "", status.Error(codes.Internal, "tickets unavailable")
	}
	return encode(map[string]any{"raffle": map[string]any{
		"id":         lottery.ID,
		"headline":   map[string]string{"label": "جایزه بزرگ", "prize": prize, "ends": raffleEndsLabel(lottery.Draw, now)},
		"stats":      []map[string]string{{"label": "کل بلیط‌ها", "value": faGrouped(int64(lottery.Total))}, {"label": "بلیط من", "value": faGrouped(mine)}},
		"ticketCost": max(1, lottery.Price),
	}})
}

// planetDiscover is the Sky (PDD Page-6): active planets of active brands with
// the prize value of their running lotteries, active group count and the
// viewer's membership.
func planetDiscover(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		Search   string `json:"search"`
		Category string `json:"category"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	in.Search, in.Category = strings.TrimSpace(in.Search), strings.TrimSpace(in.Category)
	if !utf8.ValidString(in.Search) || utf8.RuneCountInString(in.Search) > 64 || len(in.Category) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid discovery filter")
	}
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(in.Search) + "%"
	now := time.Now().UTC()
	rows, err := db.QueryContext(ctx, `SELECT `+clientPlanetColumns+`,
		COALESCE((SELECT sum(pr.value_rial) FROM jazebe_lottery_prize pr JOIN jazebe_lottery l ON l.id=pr.lottery_id
			WHERE pr.planet_id=p.id AND pr.deleted_at IS NULL AND l.status='active' AND l.draw_at>$4),0)::bigint,
		(SELECT count(*) FROM jazebe_group g WHERE g.planet_id=p.id AND g.status='active'),
		EXISTS(SELECT 1 FROM jazebe_planet_membership m WHERE m.planet_id=p.id AND m.user_id=$1 AND m.is_active)`+clientPlanetFrom+`
		WHERE p.status='active' AND b.is_active
		AND ($2::text='' OR p.name ILIKE $3 OR p.latin_name ILIKE $3 OR p.slug ILIKE $3 OR b.name ILIKE $3)
		AND ($5::text='' OR c.id::text=$5::text OR c.key=$5::text)
		ORDER BY p.fans_count DESC, p.name ASC LIMIT $6`, user, in.Search, pattern, now, in.Category, planetDiscoverLimit)
	if err != nil {
		return "", status.Error(codes.Internal, "discovery unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var p clientPlanet
		var prizes int64
		var groups int
		var isMember bool
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.PointName, &p.Status, &p.LatinName, &p.ShortDescription, &p.BrandColor, &p.Glyph,
			&p.Icon, &p.Logo, &p.Cover, &p.CategoryID, &p.CategoryKey, &p.CategoryTitle, &p.Fans, &p.BrandActive, &prizes, &groups, &isMember); err != nil {
			return "", status.Error(codes.Internal, "discovery unavailable")
		}
		var category any
		if p.CategoryKey.Valid {
			category = map[string]string{"key": p.CategoryKey.String, "titleFa": p.CategoryTitle.String}
		}
		items = append(items, map[string]any{"id": p.ID, "slug": p.Slug, "name": p.Name, "latinName": nullable(p.LatinName),
			"icon": nullable(p.Icon), "glyph": p.glyph(), "color": p.color(), "shortDescription": nullable(p.ShortDescription),
			"category": category, "totalActivePrizeValue": prizes, "activeGroupCount": groups, "fans": p.Fans, "isMember": isMember})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "discovery unavailable")
	}
	return encode(map[string]any{"items": items})
}
