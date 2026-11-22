package samplecleanup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/secretbox"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	Sample            = "stargame"
	StorePackage      = "ir.nakama.stargame"
	Leaderboard       = "stargame_tictactoe_wins"
	RunCollection     = "base_maintenance"
	RunKey            = "stargame_cleanup"
	systemUserID      = "00000000-0000-0000-0000-000000000000"
	reasonEdited      = "edited"
	reasonInUse       = "in_use"
	reasonInUseOff    = "in_use_switched_off"
	reasonNoSecret    = "no_secret"
	reasonNotSample   = "not_sample"
	reasonPackageKept = "sample_package_edited"
	actionReset       = "reset"
)

type Options struct {
	Apply bool
	By    string
	Env   map[string]string
	Now   time.Time
}

type Kept struct {
	Key    string `json:"key"`
	Reason string `json:"reason"`
}

type Table struct {
	Table   string   `json:"table"`
	Kind    string   `json:"kind"`
	Deleted []string `json:"deleted"`
	Kept    []Kept   `json:"kept"`
	Related int64    `json:"related"`
}

type Change struct {
	Kind   string `json:"kind"`
	Key    string `json:"key"`
	Action string `json:"action"`
	Reason string `json:"reason,omitempty"`
}

type Result struct {
	Applied      bool     `json:"applied"`
	By           string   `json:"by,omitempty"`
	At           string   `json:"at"`
	Tables       []Table  `json:"tables"`
	Changes      []Change `json:"changes"`
	Leaderboards []string `json:"leaderboards"`
	Deleted      int      `json:"deleted"`
	Kept         int      `json:"kept"`
	Related      int64    `json:"related"`
	Changed      int      `json:"changed"`
}

func (r *Result) Empty() bool {
	return r.Deleted == 0 && r.Changed == 0 && len(r.Leaderboards) == 0
}

type step struct {
	table   string
	kind    string
	match   string
	keys    []string
	guard   string
	related []string
	before  []string
	off     bool
}

func steps() []step {
	return []step{
		{table: "publication_banner", kind: "banner", match: "key=$1", keys: []string{"star_bn_welcome", "star_bn_vip", "star_bn_galaxy", "star_bn_missions", "star_bn_fan", "star_bn_nova"},
			related: []string{`SELECT count(*) FROM publication_banner_day WHERE banner_id=$1`, `SELECT count(*) FROM publication_banner_seen WHERE banner_id=$1`}, off: true},
		{table: "publication_banner_list", kind: "banner_list", match: "key=$1", keys: []string{"menu_carousel"},
			guard: `name->>'en'='Home menu carousel' AND layout='carousel'`, off: true},
		{table: "publication_banner_list", kind: "banner_list", match: "key=$1", keys: []string{"luck_sponsor"},
			guard: `name->>'en'='Luck screen sponsor' AND (sponsor_id IS NULL OR sponsor_id IN (SELECT id FROM publication_sponsor WHERE key='star_sp_nova'))`, off: true},
		{table: "publication_sponsor", kind: "sponsor", match: "key=$1", keys: []string{"star_sp_nova"},
			related: []string{`SELECT count(*) FROM publication_sponsor_day WHERE sponsor_id=$1`, `SELECT count(*) FROM publication_sponsor_seen WHERE sponsor_id=$1`}, off: true},
		{table: "publication_auto_rule", kind: "auto_rule", match: "key=$1", keys: []string{"star_auto_badge", "star_auto_wheel", "star_auto_streak", "star_auto_comeback"},
			related: []string{`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1`}, off: true},
		{table: "base_config_key", kind: "config_key", match: "key=$1", keys: []string{"stargame.banner", "stargame.golden_rush"},
			before: []string{`DELETE FROM base_config_history WHERE key_id=$1`}, off: true},
		{table: "economy_gift", kind: "gift", match: "code=$1", keys: []string{"STAR100", "HALF50", "STARFAN"},
			before: []string{`DELETE FROM economy_gift_use WHERE gift_id=$1 AND payment_id IS NULL`}, off: true},
		{table: "economy_game", kind: "game", match: "key=$1", keys: []string{"stargame", "stargame_tictactoe"}, off: true},
		{table: "base_category", kind: "game_category", match: "key_type='game' AND key=$1", keys: []string{"star_arcade", "star_board"}},
		{table: "economy_chance_game", kind: "chance_game", match: "key=$1", keys: []string{"star_wheel_daily", "star_lucky_box"},
			related: []string{`SELECT count(*) FROM economy_chance_play WHERE game_id=$1`, `SELECT count(*) FROM economy_chance_state WHERE game_id=$1`}, off: true},
		{table: "economy_prize_group", kind: "prize_group", match: "key=$1", keys: []string{"star_box_coins", "star_box_big"}},
		{table: "economy_achievement", kind: "achievement", match: "key=$1", keys: []string{"star_ach_first", "star_ach_rounds", "star_ach_hunter", "star_ach_supporter", "star_ach_legend"},
			related: []string{`SELECT count(*) FROM economy_user_achievement WHERE achievement_id=$1`}, off: true},
		{table: "base_category", kind: "achievement_category", match: "key_type='achievement' AND key=$1", keys: []string{"star_ach_play", "star_ach_special"}},
		{table: "economy_task", kind: "task", match: "key=$1", keys: []string{"star_mission_starter_round", "star_mission_starter_golden", "star_mission_starter_big", "star_chain_one", "star_chain_two", "star_chain_three", "star_mission_two_rounds", "star_mission_golden_five", "star_mission_week_points", "star_mission_veteran", "star_mission_starter"},
			related: []string{`SELECT count(*) FROM economy_user_task WHERE task_id=$1`}, off: true},
		{table: "economy_task_group", kind: "task_group", match: "key=$1", keys: []string{"star_hunter_chain"},
			guard:   `NOT EXISTS (SELECT 1 FROM economy_task t WHERE t.group_id=economy_task_group.id)`,
			related: []string{`SELECT count(*) FROM economy_user_task_group WHERE group_id=$1`}, off: true},
		{table: "base_category", kind: "task_category", match: "key_type='task' AND key=$1", keys: []string{"star_missions_daily", "star_missions_story"}},
		{table: "economy_product", kind: "product", match: "key=$1", keys: []string{"star_pack_small", "star_pack_medium", "star_pack_large", "star_pack_vip", "star_skin_gold", "star_skin_neon", "star_skin_galaxy", "star_item_ruby", "star_starter_pack"}, off: true},
		{table: "base_category", kind: "product_category", match: "key_type='product' AND key=$1", keys: []string{"star_coin_packs", "star_skins"}},
		{table: "economy_item", kind: "item", match: "key=$1", keys: []string{"star_basket_classic", "star_basket_ruby", "star_basket_ice", "star_frame_gold"},
			before: []string{`DELETE FROM economy_user_item WHERE item_id=$1`}, off: true},
		{table: "economy_event_rule", kind: "event_rule", match: "key=$1", keys: []string{"star_daily_rounds", "star_golden_daily", "star_first_100", "star_every_1000"},
			related: []string{`SELECT count(*) FROM economy_event_rule_progress WHERE rule_id=$1`}, off: true},
		{table: "economy_meta_rule", kind: "meta_rule", match: "key=$1", keys: []string{"star_xp_rounds", "star_xp_golden"}, off: true},
		{table: "economy_level_table", kind: "level_table", match: "key=$1", keys: []string{"star_levels"},
			related: []string{`SELECT count(*) FROM economy_user_level WHERE table_id=$1`}, off: true},
		{table: "economy_streak", kind: "streak", match: "key=$1", keys: []string{"daily_login"},
			guard: `name->>'en'='Daily sign-in reward' AND cycle_length=7 AND description->>'en' IN ('` + streakDescEn + `','` + streakDescOldEn + `')
				AND NOT EXISTS (SELECT 1 FROM economy_streak_reward r JOIN economy_prize p ON p.id=r.prize_id WHERE r.streak_id=economy_streak.id AND p.key NOT LIKE 'star\_coins\_%')
				AND NOT EXISTS (SELECT 1 FROM economy_streak_milestone m JOIN economy_prize p ON p.id=m.prize_id WHERE m.streak_id=economy_streak.id AND p.key NOT LIKE 'star\_coins\_%')`,
			related: []string{`SELECT count(*) FROM economy_user_streak WHERE streak_id=$1`}, off: true},
		{table: "economy_prize", kind: "prize", match: "key=$1", keys: samplePrizes(),
			before: prizeBefore},
		{table: "economy_prize", kind: "prize", match: "key=$1", keys: []string{"referral_inviter"},
			guard: `message->>'en'='Your friend joined with your code and played; you got 100 coins' AND items @> '[{"type":"currency","key":"coin","value":100}]' AND ` + notReferral, before: prizeBefore},
		{table: "economy_prize", kind: "prize", match: "key=$1", keys: []string{"referral_invited"},
			guard: `message->>'en'='You got 50 coins for using a referral code' AND items @> '[{"type":"currency","key":"coin","value":50}]' AND ` + notReferral, before: prizeBefore},
		{table: "economy_event_type", kind: "event_type", match: "key=$1", keys: []string{"round_finished"},
			guard:   `name->>'en'='Round finished' AND NOT is_system AND NOT EXISTS (SELECT 1 FROM economy_referral_settings WHERE goal_event_type_id=economy_event_type.id)`,
			related: []string{`SELECT count(*) FROM economy_player_event WHERE type_id=$1`}, off: true},
		{table: "economy_event_type", kind: "event_type", match: "key=$1", keys: []string{"golden_star"},
			guard:   `name->>'en'='Golden star caught' AND NOT is_system`,
			related: []string{`SELECT count(*) FROM economy_player_event WHERE type_id=$1`}, off: true},
		{table: "base_tag", kind: "player_tag", match: "key_type='user' AND key=$1", keys: []string{"star-vip"},
			related: []string{`SELECT count(*) FROM base_user_tag WHERE tag_id=$1`}},
		{table: "base_app_version", kind: "app_version", match: "platform=$1", keys: []string{"web"},
			guard: `store_links::text LIKE '%/stargame/%' AND update_message->>'en'='` + appUpdateEn + `'`, off: true},
		{table: "base_app_version", kind: "app_version", match: "platform=$1", keys: []string{"android"},
			guard: `store_links::text LIKE '%` + StorePackage + `%' AND update_message->>'en'='` + appUpdateEn + `'`, off: true},
	}
}

const (
	streakDescEn    = "Sign in every day and collect that day''s prize. Missing one day is fine; two days away start you again from day one, but once you reach day 7 you are safe up to day 7 and win 100 coins too."
	streakDescOldEn = "Sign in every day and collect that day''s prize. Missing one day is fine, but two days away start you again from day one."
	appUpdateEn     = "A new Star Catcher is out. Update to catch more stars!"
)

const notReferral = `NOT EXISTS (SELECT 1 FROM economy_referral_settings WHERE referrer_prize_id=economy_prize.id OR referee_prize_id=economy_prize.id)`

var prizeBefore = []string{
	`DELETE FROM economy_user_prize WHERE prize_id=$1`,
	`DELETE FROM economy_reward_claim WHERE prize_id=$1`,
}

func samplePrizes() []string {
	out := []string{}
	for _, n := range []int{10, 15, 20, 25, 30, 40, 50, 100, 200, 500, 1200} {
		out = append(out, fmt.Sprintf("star_coins_%d", n))
	}
	return append(out, "star_skin_gold", "star_skin_neon", "star_skin_galaxy", "star_item_ruby", "star_starter_pack", "star_fan_badge")
}

type sampleMethod struct {
	method string
	config map[string]any
}

var sampleMethods = []sampleMethod{
	{"zarinpal", map[string]any{"sandbox": true}},
	{"zibal", map[string]any{"sandbox": true}},
	{"bazaar", map[string]any{"packageName": StorePackage, "sandbox": true}},
	{"myket", map[string]any{"packageName": StorePackage, "sandbox": true}},
}

func Keys() map[string][]string {
	out := map[string][]string{}
	for _, s := range steps() {
		out[s.kind] = append(out[s.kind], s.keys...)
	}
	for _, m := range sampleMethods {
		out["payment_method"] = append(out["payment_method"], m.method)
	}
	return out
}

func Run(ctx context.Context, db *sql.DB, opt Options) (*Result, error) {
	now := opt.Now
	if now.IsZero() {
		now = time.Now()
	}
	res := &Result{Applied: opt.Apply, By: opt.By, At: now.UTC().Format(time.RFC3339), Tables: []Table{}, Changes: []Change{}, Leaderboards: []string{}}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('samplecleanup:stargame'))`); err != nil {
		return nil, err
	}
	if err := resetReferral(ctx, tx, res); err != nil {
		return nil, err
	}
	byTable := map[string]int{}
	for _, s := range steps() {
		ok, err := tableExists(ctx, tx, s.table)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		i, found := byTable[s.table+"/"+s.kind]
		if !found {
			res.Tables = append(res.Tables, Table{Table: s.table, Kind: s.kind, Deleted: []string{}, Kept: []Kept{}})
			i = len(res.Tables) - 1
			byTable[s.table+"/"+s.kind] = i
		}
		if err := runStep(ctx, tx, s, &res.Tables[i]); err != nil {
			return nil, fmt.Errorf("%s: %w", s.table, err)
		}
	}
	if err := resetMethods(ctx, tx, opt.Env, res); err != nil {
		return nil, err
	}
	var board int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM leaderboard WHERE id=$1 AND metadata->>'game'='stargame_tictactoe'`, Leaderboard).Scan(&board); err != nil {
		return nil, err
	}
	if board > 0 {
		res.Leaderboards = append(res.Leaderboards, Leaderboard)
	}
	for _, t := range res.Tables {
		res.Deleted += len(t.Deleted)
		res.Kept += len(t.Kept)
		res.Related += t.Related
	}
	for _, m := range res.Changes {
		if m.Action == actionReset {
			res.Changed++
		}
	}
	if !opt.Apply {
		return res, nil
	}
	if err := saveRun(ctx, tx, res); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

func LastRun(ctx context.Context, db *sql.DB) (*Result, error) {
	var raw string
	err := db.QueryRowContext(ctx, `SELECT value::text FROM storage WHERE collection=$1 AND key=$2 AND user_id=$3`, RunCollection, RunKey, systemUserID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out Result
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func saveRun(ctx context.Context, tx *sql.Tx, res *Result) error {
	raw, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO storage (collection, key, user_id, value, version, read, write, create_time, update_time)
		VALUES ($1, $2, $3, $4::jsonb, md5($4::jsonb::text), 0, 0, now(), now())
		ON CONFLICT (collection, key, user_id) DO UPDATE SET value=EXCLUDED.value, version=EXCLUDED.version, update_time=now()`, RunCollection, RunKey, systemUserID, string(raw))
	return err
}

func tableExists(ctx context.Context, tx *sql.Tx, table string) (bool, error) {
	var ok bool
	err := tx.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+table).Scan(&ok)
	return ok, err
}

func runStep(ctx context.Context, tx *sql.Tx, s step, out *Table) error {
	guard := "true"
	if s.guard != "" {
		guard = s.guard
	}
	for _, key := range s.keys {
		var id string
		var sample bool
		err := tx.QueryRowContext(ctx, `SELECT id::text, (`+guard+`) FROM `+s.table+` WHERE `+s.match, key).Scan(&id, &sample)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if !sample {
			out.Kept = append(out.Kept, Kept{Key: key, Reason: reasonEdited})
			continue
		}
		if _, err := tx.ExecContext(ctx, `SAVEPOINT samplecleanup_row`); err != nil {
			return err
		}
		var related int64
		for _, q := range s.related {
			var n int64
			if err := tx.QueryRowContext(ctx, q, id).Scan(&n); err != nil {
				return err
			}
			related += n
		}
		deleted, err := deleteRow(ctx, tx, s, id, &related)
		if err != nil {
			return err
		}
		if deleted {
			if _, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT samplecleanup_row`); err != nil {
				return err
			}
			out.Deleted = append(out.Deleted, key)
			out.Related += related
			continue
		}
		if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT samplecleanup_row`); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT samplecleanup_row`); err != nil {
			return err
		}
		reason := reasonInUse
		if s.off {
			if _, err := tx.ExecContext(ctx, `UPDATE `+s.table+` SET is_active=false, updated_at=now() WHERE id::text=$1 AND is_active`, id); err != nil {
				return err
			}
			reason = reasonInUseOff
		}
		out.Kept = append(out.Kept, Kept{Key: key, Reason: reason})
	}
	return nil
}

func deleteRow(ctx context.Context, tx *sql.Tx, s step, id string, related *int64) (bool, error) {
	for _, q := range s.before {
		r, err := tx.ExecContext(ctx, q, id)
		if err != nil {
			if isForeignKey(err) {
				return false, nil
			}
			return false, err
		}
		n, _ := r.RowsAffected()
		*related += n
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+s.table+` WHERE id::text=$1`, id); err != nil {
		if isForeignKey(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isForeignKey(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23503"
}

func resetReferral(ctx context.Context, tx *sql.Tx, res *Result) error {
	ok, err := tableExists(ctx, tx, "economy_referral_settings")
	if err != nil || !ok {
		return err
	}
	r, err := tx.ExecContext(ctx, `UPDATE economy_referral_settings SET referrer_prize_id=NULL, referee_prize_id=NULL, goal_event_type_id=NULL, updated_at=now()
		WHERE referrer_prize_id=(SELECT id FROM economy_prize WHERE key='referral_inviter')
		AND referee_prize_id=(SELECT id FROM economy_prize WHERE key='referral_invited')
		AND goal_event_type_id=(SELECT id FROM economy_event_type WHERE key='round_finished')
		AND goal_measure='count' AND goal_value=1`)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n > 0 {
		res.Changes = append(res.Changes, Change{Kind: "referral_rule", Key: "default", Action: actionReset})
	}
	return nil
}

func resetMethods(ctx context.Context, tx *sql.Tx, env map[string]string, res *Result) error {
	ok, err := tableExists(ctx, tx, "economy_payment_method")
	if err != nil || !ok {
		return err
	}
	var key []byte
	if secret, err := secretbox.Secret(env); err == nil {
		key, err = secretbox.Economy.DeriveKey(secret, secretbox.EconomyConfigInfo)
		if err != nil {
			return err
		}
	}
	for _, m := range sampleMethods {
		var sealed string
		var active bool
		err := tx.QueryRowContext(ctx, `SELECT config, is_active FROM economy_payment_method WHERE method=$1 FOR UPDATE`, m.method).Scan(&sealed, &active)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && sealed == "") {
			continue
		}
		if err != nil {
			return err
		}
		if key == nil {
			res.Changes = append(res.Changes, Change{Kind: "payment_method", Key: m.method, Action: "kept", Reason: reasonNoSecret})
			continue
		}
		plain, err := secretbox.Economy.Open(key, sealed)
		if err != nil {
			res.Changes = append(res.Changes, Change{Kind: "payment_method", Key: m.method, Action: "kept", Reason: reasonNotSample})
			continue
		}
		var cfg map[string]any
		if json.Unmarshal(plain, &cfg) != nil {
			res.Changes = append(res.Changes, Change{Kind: "payment_method", Key: m.method, Action: "kept", Reason: reasonNotSample})
			continue
		}
		if sameConfig(cfg, m.config) {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_payment_method SET is_active=false, config='', updated_at=now() WHERE method=$1`, m.method); err != nil {
				return err
			}
			res.Changes = append(res.Changes, Change{Kind: "payment_method", Key: m.method, Action: actionReset})
			continue
		}
		if pkg, _ := cfg["packageName"].(string); strings.TrimSpace(pkg) == StorePackage {
			res.Changes = append(res.Changes, Change{Kind: "payment_method", Key: m.method, Action: "kept", Reason: reasonPackageKept})
		}
	}
	return nil
}

func sameConfig(got, want map[string]any) bool {
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(want)
	return string(a) == string(b)
}
