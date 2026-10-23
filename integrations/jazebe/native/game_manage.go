package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gameManageInput struct {
	PlanetID             string  `json:"planetId"`
	GameID               string  `json:"gameId"`
	TemplateID           string  `json:"templateId"`
	Name                 *string `json:"name"`
	Description          *string `json:"description"`
	Icon                 *string `json:"icon"`
	Accent               *string `json:"accent"`
	Status               string  `json:"status"`
	EntryModel           *string `json:"entryModel"`
	PointCost            *int    `json:"pointCost"`
	FreeIndividualCount  *int    `json:"freeIndividualCount"`
	FreeIndividualPeriod *string `json:"freeIndividualPeriod"`
	FreeGroupCount       *int    `json:"freeGroupCount"`
	FreeGroupPeriod      *string `json:"freeGroupPeriod"`
	PrizeTicketMin       *int    `json:"prizeTicketMin"`
	PrizeTicketMax       *int    `json:"prizeTicketMax"`
	RewardXp             *int    `json:"rewardXp"`
	TargetLotteryID      *string `json:"targetLotteryId"`
	DurationSec          *int    `json:"durationSec"`
	Search               string  `json:"search"`
	Page                 int     `json:"page"`
	Limit                int     `json:"limit"`
}

const brandGameSelect = `SELECT jsonb_build_object(
	'id',g.id,'planetId',g.planet_id,'brandId',p.brand_id,'templateId',g.template_id,'kind',g.kind,
	'name',g.name,'description',g.description,'icon',g.icon,'accent',g.accent,
	'engine',COALESCE(t.engine,g.config->>'engine','box'),'entryModel',g.entry_model,'pointCost',g.point_cost,
	'freeIndividualCount',g.free_individual_count,'freeIndividualPeriod',g.free_individual_period,
	'freeGroupCount',g.free_group_count,'freeGroupPeriod',g.free_group_period,
	'prizeTicketMin',g.prize_ticket_min,'prizeTicketMax',g.prize_ticket_max,'rewardXp',g.reward_xp,
	'targetLotteryId',g.target_lottery_id,
	'durationSec',COALESCE((g.config->>'durationSec')::integer,(t.default_config->>'durationSec')::integer,10),
	'status',g.status,'createdAt',g.created_at,'updatedAt',g.updated_at,
	'catalogSuspension',CASE WHEN t.id IS NOT NULL AND (NOT t.is_active OR t.suspended_at IS NOT NULL)
		THEN jsonb_build_object('reason',t.suspend_reason,'at',t.suspended_at,'retired',NOT t.is_active AND t.suspended_at IS NULL) ELSE NULL END,
	'suspension',CASE WHEN g.status='suspended' THEN jsonb_build_object('reason',audit.params->>'reason','at',audit.created_at) ELSE NULL END,
	'stats',jsonb_build_object('plays',stats.plays,'freePlays',stats.free_plays,'paidPlays',stats.paid_plays,'ticketsIssued',stats.tickets))
	FROM jazebe_game g JOIN jazebe_planet p ON p.id=g.planet_id
	LEFT JOIN jazebe_game_template t ON t.id=g.template_id
	LEFT JOIN LATERAL (SELECT count(*) plays,count(*) FILTER(WHERE entry_type<>'paid') free_plays,
		count(*) FILTER(WHERE entry_type='paid') paid_plays,COALESCE(sum(awarded_tickets),0) tickets
		FROM jazebe_game_session WHERE game_id=g.id) stats ON true
	LEFT JOIN LATERAL (SELECT params,created_at FROM jazebe_audit_log WHERE target_type='game' AND target_id=g.id
		AND action='game.suspend' ORDER BY created_at DESC,id DESC LIMIT 1) audit ON true`

func brandGameRow(ctx context.Context, db *sql.DB, planetID, gameID string) (map[string]any, error) {
	var raw []byte
	err := db.QueryRowContext(ctx, brandGameSelect+` WHERE g.planet_id=$1 AND g.id=$2`, planetID, gameID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "game not found")
	}
	if err != nil {
		return nil, err
	}
	var row map[string]any
	if err = json.Unmarshal(raw, &row); err != nil {
		return nil, err
	}
	return row, nil
}

func gameBrandGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	gameID, err := requiredUUID(in.GameID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	row, err := brandGameRow(ctx, db, planetID, gameID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func gameBrandList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Page > 100000 || in.Limit < 1 || in.Limit > 100 || len([]rune(in.Search)) > 128 {
		return "", status.Error(codes.InvalidArgument, "invalid game filter")
	}
	if in.Status != "" && in.Status != "active" && in.Status != "paused" && in.Status != "suspended" {
		return "", status.Error(codes.InvalidArgument, "invalid game status")
	}
	search := strings.TrimSpace(in.Search)
	search = strings.ReplaceAll(search, `\`, `\\`)
	search = strings.ReplaceAll(search, `%`, `\%`)
	search = strings.ReplaceAll(search, `_`, `\_`)
	var total int64
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_game g WHERE g.planet_id=$1
		AND ($2::text IS NULL OR g.status=$2) AND ($3::text IS NULL OR g.name ILIKE $3 ESCAPE E'\\')`,
		planetID, optionalText(in.Status), optionalSearch(search)).Scan(&total)
	if err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, brandGameSelect+` WHERE g.planet_id=$1
		AND ($2::text IS NULL OR g.status=$2) AND ($3::text IS NULL OR g.name ILIKE $3 ESCAPE E'\\')
		ORDER BY g.created_at,g.id LIMIT $4 OFFSET $5`, planetID, optionalText(in.Status), optionalSearch(search), in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			return "", err
		}
		var row map[string]any
		if err = json.Unmarshal(raw, &row); err != nil {
			return "", err
		}
		items = append(items, row)
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

// Games are brand content. Both technical and marketing administrators may
// manage them; an unrelated planet's administrator never can.
func gameBrandAccess(ctx context.Context, db *sql.DB, planetID, user string) error {
	var role sql.NullString
	err := db.QueryRowContext(ctx, `SELECT a.role FROM jazebe_planet p JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id
		AND a.user_id=$2 AND a.is_active WHERE p.id=$1`, planetID, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.PermissionDenied, "brand content role required")
	}
	if err != nil {
		return err
	}
	if role.String != "main" && role.String != "technical" && role.String != "marketing" {
		return status.Error(codes.PermissionDenied, "brand content role required")
	}
	return nil
}

func gameBrandEnable(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	templateID, err := requiredUUID(in.TemplateID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	var name string
	var description, icon sql.NullString
	var config []byte
	var active bool
	err = db.QueryRowContext(ctx, `SELECT name,description,icon,default_config,is_active AND suspended_at IS NULL
		FROM jazebe_game_template WHERE id=$1`, templateID).Scan(&name, &description, &icon, &config, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "game template not found")
	}
	if err != nil {
		return "", err
	}
	if !active {
		return "", status.Error(codes.FailedPrecondition, "game template suspended")
	}
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
	}
	if len([]rune(name)) < 1 || len([]rune(name)) > 128 {
		return "", status.Error(codes.InvalidArgument, "invalid game name")
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = db.ExecContext(ctx, `INSERT INTO jazebe_game(id,planet_id,template_id,name,description,icon,entry_model,
		free_individual_count,free_individual_period,prize_ticket_min,prize_ticket_max,config,status)
		VALUES($1,$2,$3,$4,$5,$6,'free',3,'daily',1,1,$7::jsonb,'paused')`, id, planetID, templateID, name, description, icon, string(config))
	if err != nil {
		return "", err
	}
	row, err := brandGameRow(ctx, db, planetID, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func gameConfigValid(g playableGame) error {
	if g.EntryModel != "free" && g.EntryModel != "cost" && g.EntryModel != "hybrid" {
		return status.Error(codes.InvalidArgument, "invalid entry model")
	}
	if g.PointCost < 0 || g.PointCost > 1000000 || g.EntryModel != "free" && g.PointCost == 0 {
		return status.Error(codes.InvalidArgument, "invalid point cost")
	}
	if g.FreeIndividualCount < 0 || g.FreeIndividualCount > 1000 || g.FreeGroupCount < 0 || g.FreeGroupCount > 10000 {
		return status.Error(codes.InvalidArgument, "invalid free quota")
	}
	if g.EntryModel != "cost" && g.FreeIndividualCount == 0 && g.FreeGroupCount == 0 {
		return status.Error(codes.InvalidArgument, "free quota required")
	}
	if g.PrizeMin < 0 || g.PrizeMax < g.PrizeMin || g.PrizeMax > 10000 {
		return status.Error(codes.InvalidArgument, "invalid ticket prize range")
	}
	if g.RewardXp < 0 || g.RewardXp > 1000000 {
		return status.Error(codes.InvalidArgument, "invalid game XP")
	}
	if g.PrizeMax > 0 && !g.LotteryID.Valid {
		return status.Error(codes.InvalidArgument, "lottery required for tickets")
	}
	if g.DurationSec < 3 || g.DurationSec > 600 {
		return status.Error(codes.InvalidArgument, "invalid game duration")
	}
	return nil
}

func gameBrandUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	var fields map[string]json.RawMessage
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	if err = json.Unmarshal([]byte(payload), &fields); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid game update")
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	gameID, err := requiredUUID(in.GameID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	g, err := loadPlayableGame(ctx, tx, planetID, gameID)
	if err != nil {
		return "", err
	}
	if g.Status == "suspended" {
		return "", status.Error(codes.FailedPrecondition, "game suspended")
	}
	var description, icon sql.NullString
	var accent string
	if err = tx.QueryRowContext(ctx, `SELECT description,icon,accent FROM jazebe_game WHERE id=$1 AND planet_id=$2 FOR UPDATE`, gameID, planetID).Scan(&description, &icon, &accent); err != nil {
		return "", err
	}
	if in.EntryModel != nil {
		g.EntryModel = *in.EntryModel
	}
	if in.PointCost != nil {
		g.PointCost = *in.PointCost
	}
	if in.FreeIndividualCount != nil {
		g.FreeIndividualCount = *in.FreeIndividualCount
	} else if string(fields["freeIndividualCount"]) == "null" {
		g.FreeIndividualCount = 0
	}
	if in.FreeGroupCount != nil {
		g.FreeGroupCount = *in.FreeGroupCount
	} else if string(fields["freeGroupCount"]) == "null" {
		g.FreeGroupCount = 0
	}
	if in.PrizeTicketMin != nil {
		g.PrizeMin = *in.PrizeTicketMin
	}
	if in.PrizeTicketMax != nil {
		g.PrizeMax = *in.PrizeTicketMax
	}
	if in.RewardXp != nil {
		g.RewardXp = *in.RewardXp
	}
	if in.TargetLotteryID != nil {
		if *in.TargetLotteryID == "" {
			g.LotteryID = sql.NullString{}
		} else {
			lottery, uuidErr := requiredUUID(*in.TargetLotteryID)
			if uuidErr != nil {
				return "", uuidErr
			}
			var valid bool
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery WHERE id=$1 AND planet_id=$2 AND status IN ('draft','active'))`, lottery, planetID).Scan(&valid); err != nil {
				return "", err
			}
			if !valid {
				return "", status.Error(codes.InvalidArgument, "invalid target lottery")
			}
			g.LotteryID = sql.NullString{String: lottery, Valid: true}
		}
	} else if string(fields["targetLotteryId"]) == "null" {
		g.LotteryID = sql.NullString{}
	}
	if in.DurationSec != nil {
		g.DurationSec = *in.DurationSec
	}
	if string(fields["freeIndividualPeriod"]) == "null" {
		g.FreeIndividualPeriod = sql.NullString{}
	}
	if string(fields["freeGroupPeriod"]) == "null" {
		g.FreeGroupPeriod = sql.NullString{}
	}
	if err = gameConfigValid(g); err != nil {
		return "", err
	}
	name := g.Name
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
	}
	if len([]rune(name)) < 1 || len([]rune(name)) > 128 {
		return "", status.Error(codes.InvalidArgument, "invalid game name")
	}
	if in.Description != nil {
		value := strings.TrimSpace(*in.Description)
		description = sql.NullString{String: value, Valid: value != ""}
	} else if string(fields["description"]) == "null" {
		description = sql.NullString{}
	}
	if in.Icon != nil {
		value := strings.TrimSpace(*in.Icon)
		icon = sql.NullString{String: value, Valid: value != ""}
	} else if string(fields["icon"]) == "null" {
		icon = sql.NullString{}
	}
	if len([]rune(description.String)) > 4000 || len([]rune(icon.String)) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid game description or icon")
	}
	if in.Accent != nil {
		accent = *in.Accent
	}
	if accent != "gold" && accent != "brand" && accent != "success" {
		return "", status.Error(codes.InvalidArgument, "invalid game accent")
	}
	period := func(in *string, old sql.NullString) (any, error) {
		if in == nil {
			if old.Valid {
				return old.String, nil
			}
			return nil, nil
		}
		if *in == "" {
			return nil, nil
		}
		if *in != "daily" && *in != "weekly" && *in != "monthly" {
			return nil, status.Error(codes.InvalidArgument, "invalid quota period")
		}
		return *in, nil
	}
	individualPeriod, err := period(in.FreeIndividualPeriod, g.FreeIndividualPeriod)
	if err != nil {
		return "", err
	}
	groupPeriod, err := period(in.FreeGroupPeriod, g.FreeGroupPeriod)
	if err != nil {
		return "", err
	}
	if g.FreeIndividualCount > 0 && individualPeriod == nil {
		individualPeriod = "daily"
	}
	if g.FreeGroupCount > 0 && groupPeriod == nil {
		groupPeriod = "daily"
	}
	var target any
	if g.LotteryID.Valid {
		target = g.LotteryID.String
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_game SET name=$3,entry_model=$4,point_cost=$5,
		free_individual_count=NULLIF($6::integer,0),free_individual_period=$7,
		free_group_count=NULLIF($8::integer,0),free_group_period=$9,prize_ticket_min=$10,prize_ticket_max=$11,
		reward_xp=$12,target_lottery_id=$13,config=jsonb_set(config,'{durationSec}',$14::jsonb,true),updated_at=$15,
		description=$16,icon=$17,accent=$18 WHERE id=$1 AND planet_id=$2`, gameID, planetID, name, g.EntryModel, g.PointCost, g.FreeIndividualCount, individualPeriod, g.FreeGroupCount, groupPeriod, g.PrizeMin, g.PrizeMax, g.RewardXp, target, stringInt(g.DurationSec), time.Now().UTC(), description, icon, accent)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	row, err := brandGameRow(ctx, db, planetID, gameID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func stringInt(n int) string { return strconv.Itoa(n) }

func gameBrandStatus(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	gameID, err := requiredUUID(in.GameID)
	if err != nil {
		return "", err
	}
	if in.Status != "active" && in.Status != "paused" {
		return "", status.Error(codes.InvalidArgument, "invalid game status")
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	g, err := loadPlayableGame(ctx, tx, planetID, gameID)
	if err != nil {
		return "", err
	}
	if g.Status == "suspended" {
		return "", status.Error(codes.FailedPrecondition, "game suspended")
	}
	if in.Status == "active" {
		if !g.TemplateActive {
			return "", status.Error(codes.FailedPrecondition, "game template suspended")
		}
		if err = gameConfigValid(g); err != nil {
			return "", err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_game SET status=$3,updated_at=$4 WHERE id=$1 AND planet_id=$2`, gameID, planetID, in.Status, time.Now().UTC())
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	row, err := brandGameRow(ctx, db, planetID, gameID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func gameClientList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID string `json:"planetId"`
	}
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var groupID string
	err = tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active' LIMIT 1`, planetID, user).Scan(&groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	rows, err := tx.QueryContext(ctx, `SELECT g.id,g.name,g.description,g.icon,g.accent,g.entry_model,g.point_cost,
		g.free_individual_count,g.free_individual_period,g.free_group_count,g.free_group_period,
		g.prize_ticket_min,g.prize_ticket_max,g.reward_xp,g.target_lottery_id,
		COALESCE((g.config->>'durationSec')::integer,(t.default_config->>'durationSec')::integer,10),
		COALESCE(t.engine,g.config->>'engine','box'),COALESCE(t.is_active AND t.suspended_at IS NULL,true),
		COALESCE(t.icon,'gmico-box'),COALESCE(p.brand_color,'#8E73F0'),
		COALESCE((SELECT l.status='active' AND l.starts_at<=$2 AND l.draw_at>$2 FROM jazebe_lottery l WHERE l.id=g.target_lottery_id),false)
		FROM jazebe_game g JOIN jazebe_planet p ON p.id=g.planet_id
		LEFT JOIN jazebe_game_template t ON t.id=g.template_id
		WHERE g.planet_id=$1 AND g.status='active' AND p.status='active' ORDER BY g.created_at,g.id`, planetID, now)
	if err != nil {
		return "", err
	}
	type listing struct {
		g                                            playableGame
		desc, icon, accent, templateIcon, brandColor sql.NullString
		onSale                                       bool
	}
	games := []listing{}
	for rows.Next() {
		var item listing
		var individual, shared sql.NullInt64
		if err = rows.Scan(&item.g.ID, &item.g.Name, &item.desc, &item.icon, &item.accent, &item.g.EntryModel, &item.g.PointCost,
			&individual, &item.g.FreeIndividualPeriod, &shared, &item.g.FreeGroupPeriod,
			&item.g.PrizeMin, &item.g.PrizeMax, &item.g.RewardXp, &item.g.LotteryID, &item.g.DurationSec,
			&item.g.Engine, &item.g.TemplateActive, &item.templateIcon, &item.brandColor, &item.onSale); err != nil {
			rows.Close()
			return "", err
		}
		if individual.Valid {
			item.g.FreeIndividualCount = int(individual.Int64)
		}
		if shared.Valid {
			item.g.FreeGroupCount = int(shared.Int64)
		}
		games = append(games, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", err
	}
	result := []map[string]any{}
	for _, item := range games {
		g := item.g
		if !g.TemplateActive || g.Engine != "tap" && g.Engine != "box" {
			continue
		}
		own, shared, _, _, e := gameQuotaUsed(ctx, tx, g, user, groupID, now)
		if e != nil {
			return "", e
		}
		freeOwn := g.FreeIndividualCount - own
		if freeOwn < 0 {
			freeOwn = 0
		}
		freeShared := g.FreeGroupCount - shared
		if groupID == "" {
			freeShared = 0
		}
		if freeShared < 0 {
			freeShared = 0
		}
		color := "#8E73F0"
		switch item.accent.String {
		case "gold":
			color = "#F5B94A"
		case "success":
			color = "#63D7A6"
		default:
			if item.brandColor.Valid {
				color = item.brandColor.String
			}
		}
		icon := item.templateIcon.String
		if item.icon.Valid {
			icon = item.icon.String
		}
		desc := ""
		if item.desc.Valid {
			desc = item.desc.String
		}
		view := map[string]any{"id": g.ID, "engine": g.Engine, "name": g.Name, "icon": icon, "color": color, "desc": desc,
			"entry": g.EntryModel, "prize": []int{g.PrizeMin, g.PrizeMax}}
		if g.FreeIndividualCount > 0 {
			view["freeInd"] = freeOwn
			view["freeIndMax"] = g.FreeIndividualCount
		} else if g.FreeGroupCount > 0 && groupID != "" {
			view["freeInd"] = freeShared
			view["freeIndMax"] = g.FreeGroupCount
		}
		if g.FreeGroupCount > 0 && groupID != "" {
			view["freeGrp"] = freeShared
			view["freeGrpMax"] = g.FreeGroupCount
		}
		if g.EntryModel != "free" {
			view["cost"] = g.PointCost
		}
		if g.Engine == "tap" {
			view["duration"] = g.DurationSec
		}
		if g.PrizeMax > 0 && !item.onSale {
			view["lotteryClosed"] = true
		}
		result = append(result, view)
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return encode(result)
}

func gameBrandReport(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	gameID, err := requiredUUID(in.GameID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	var exists bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_game WHERE id=$1 AND planet_id=$2)`, gameID, planetID).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return "", status.Error(codes.NotFound, "game not found")
	}
	var plays, free, paid, points, tickets, players, abandoned int64
	err = db.QueryRowContext(ctx, `SELECT count(*),count(*) FILTER(WHERE entry_type<>'paid'),count(*) FILTER(WHERE entry_type='paid'),
		COALESCE(sum(point_cost),0),COALESCE(sum(awarded_tickets) FILTER(WHERE status='finished'),0),
		count(DISTINCT user_id),count(*) FILTER(WHERE status='abandoned')
		FROM jazebe_game_session WHERE game_id=$1 AND planet_id=$2`, gameID, planetID).
		Scan(&plays, &free, &paid, &points, &tickets, &players, &abandoned)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"plays": plays, "freePlays": free, "paidPlays": paid, "pointsSpent": points,
		"ticketsIssued": tickets, "players": players, "abandoned": abandoned})
}

func gameBrandCatalog(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT t.id,t.key,t.name,t.engine,t.description,t.icon,
		(t.default_config->>'durationSec')::integer,
		EXISTS(SELECT 1 FROM jazebe_game g WHERE g.planet_id=$1 AND g.template_id=t.id),
		(NOT t.is_active OR t.suspended_at IS NOT NULL)
		FROM jazebe_game_template t ORDER BY t.name,t.id`, planetID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	result := []map[string]any{}
	for rows.Next() {
		var id, key, name, engine string
		var desc, icon sql.NullString
		var duration sql.NullInt64
		var enabled, suspended bool
		if err = rows.Scan(&id, &key, &name, &engine, &desc, &icon, &duration, &enabled, &suspended); err != nil {
			return "", err
		}
		var d, i, seconds any
		if desc.Valid {
			d = desc.String
		}
		if icon.Valid {
			i = icon.String
		}
		if duration.Valid {
			seconds = duration.Int64
		}
		result = append(result, map[string]any{"id": id, "key": key, "name": name, "engine": engine,
			"description": d, "icon": i, "durationSec": seconds, "enabled": enabled, "suspended": suspended})
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(result)
}

func gameBrandMeta(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gameManageInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = gameBrandAccess(ctx, db, planetID, user); err != nil {
		return "", err
	}
	var pointName sql.NullString
	if err = db.QueryRowContext(ctx, `SELECT point_name FROM jazebe_planet WHERE id=$1`, planetID).Scan(&pointName); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT id,title,status,draw_at FROM jazebe_lottery
		WHERE planet_id=$1 AND status IN ('draft','active') ORDER BY draw_at`, planetID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	lotteries := []map[string]any{}
	for rows.Next() {
		var id, title, state string
		var draw time.Time
		if err = rows.Scan(&id, &title, &state, &draw); err != nil {
			return "", err
		}
		lotteries = append(lotteries, map[string]any{"id": id, "title": title, "status": state, "drawAt": draw})
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(map[string]any{"pointName": pointName.String, "lotteries": lotteries})
}
