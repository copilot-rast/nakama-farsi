package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	maxBotBulk     = 5000
	maxBotImport   = 2 * 1024 * 1024
	maxBotTemplate = 64
)

type botRoom struct {
	Game string `json:"game"`
	Room string `json:"room"`
}

func (s *service) botConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/bots", http.MethodGet, s.consoleBotList},
		{consolePrefix + "/bots/status", http.MethodPost, s.consoleBotStatus},
		{consolePrefix + "/bots/bulk", http.MethodPost, s.consoleBotBulk},
		{consolePrefix + "/bots/import", http.MethodPost, s.consoleBotImport},
		{consolePrefix + "/bots/topup", http.MethodPost, s.consoleBotTopup},
		{consolePrefix + "/bots/{id}", http.MethodGet, s.consoleBotGet},
		{consolePrefix + "/bots/{id}", http.MethodPut, s.consoleBotPut},
	}
}

const botJSON = `jsonb_build_object('userId',b.user_id,'username',u.username,'displayName',COALESCE(u.display_name,''),'avatarUrl',COALESCE(u.avatar_url,''),'enabled',b.enabled,
	'banned',u.disable_time > '1970-01-01 00:00:01+00','level',b.level,'template',b.template,'rooms',b.rooms,'look',b.look,'note',b.note,'source',b.source,'games',b.games,
	'lastUsedAt',b.last_used_at,'wallet',u.wallet,'createdAt',b.created_at,'updatedAt',b.updated_at,'updatedBy',b.updated_by)::text`

func (s *service) consoleBotList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	where := &sqlWhere{}
	switch q.Get("enabled") {
	case "true":
		where.add("b.enabled = ?::boolean", "true")
	case "false":
		where.add("b.enabled = ?::boolean", "false")
	}
	if v := strings.TrimSpace(q.Get("template")); v != "" {
		where.add("b.template = ?", v)
	}
	if v := strings.TrimSpace(q.Get("game")); v != "" {
		room := strings.TrimSpace(q.Get("room"))
		n := where.next()
		where.args = append(where.args, v)
		m := where.next()
		where.args = append(where.args, room)
		where.parts = append(where.parts, "(b.rooms = '[]'::jsonb OR b.rooms @> jsonb_build_array(jsonb_build_object('game',"+n+"::text,'room',"+m+"::text)) OR b.rooms @> jsonb_build_array(jsonb_build_object('game',"+n+"::text,'room','')))")
	}
	if v, err := strconv.Atoi(q.Get("minLevel")); err == nil && v > 0 {
		where.add("b.level >= ?::int", strconv.Itoa(v))
	}
	if v, err := strconv.Atoi(q.Get("maxLevel")); err == nil && v > 0 {
		where.add("b.level <= ?::int", strconv.Itoa(v))
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		n := where.next()
		where.args = append(where.args, likePattern(v))
		where.parts = append(where.parts, "(u.username ILIKE "+n+" OR u.display_name ILIKE "+n+" OR b.user_id::text ILIKE "+n+" OR b.note ILIKE "+n+")")
	}
	from := ` FROM economy_bot b JOIN users u ON u.id=b.user_id`
	order := ` ORDER BY u.username, b.user_id`
	switch q.Get("sort") {
	case "level":
		order = ` ORDER BY b.level DESC, u.username`
	case "games":
		order = ` ORDER BY b.games DESC, u.username`
	case "recent":
		order = ` ORDER BY b.last_used_at DESC NULLS LAST, u.username`
	}
	s.listRaw(w, r, `SELECT count(*)`+from, `SELECT `+botJSON+from, where, order, readPage(r))
}

func (s *service) botSummary(ctx context.Context) map[string]any {
	out := map[string]any{"total": 0, "enabled": 0, "templates": map[string]int{}}
	var total, enabled int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE enabled) FROM economy_bot`).Scan(&total, &enabled)
	out["total"], out["enabled"] = total, enabled
	rows, err := s.db.QueryContext(ctx, `SELECT template, count(*) FROM economy_bot GROUP BY template ORDER BY template`)
	if err == nil {
		t := map[string]int{}
		for rows.Next() {
			var k string
			var n int
			if rows.Scan(&k, &n) == nil {
				t[k] = n
			}
		}
		rows.Close()
		out["templates"] = t
	}
	return out
}

func (s *service) consoleBotGet(w http.ResponseWriter, r *http.Request) {
	var raw string
	err := s.db.QueryRowContext(r.Context(), `SELECT `+botJSON+` FROM economy_bot b JOIN users u ON u.id=b.user_id WHERE b.user_id::text=$1`, pathVar(r, "id")).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("BOT_NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	out["summary"] = s.botSummary(r.Context())
	writeJSON(w, http.StatusOK, out)
}

func cleanBotUsername(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" || utf8.RuneCountInString(v) > 128 {
		return "", badRequest("INVALID_USERNAME", "username", "a username is 1 to 128 characters")
	}
	for _, r := range v {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return "", badRequest("INVALID_USERNAME", "username", "a username has no spaces")
		}
	}
	return v, nil
}

func cleanBotRooms(ctx context.Context, q queryer, raw []botRoom) ([]botRoom, error) {
	out := []botRoom{}
	seen := map[string]bool{}
	for _, br := range raw {
		br.Game, br.Room = strings.TrimSpace(br.Game), strings.TrimSpace(br.Room)
		if br.Game == "" {
			continue
		}
		var roomsRaw string
		err := q.QueryRowContext(ctx, `SELECT rooms::text FROM economy_game WHERE key=$1`, br.Game).Scan(&roomsRaw)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, badRequest("GAME_NOT_FOUND", "rooms", "no game has the key "+br.Game)
		}
		if err != nil {
			return nil, err
		}
		if br.Room != "" {
			var rooms []gamecatalog.Room
			_ = json.Unmarshal([]byte(roomsRaw), &rooms)
			found := false
			for _, x := range rooms {
				found = found || x.Key == br.Room
			}
			if !found {
				return nil, badRequest("ROOM_NOT_FOUND", "rooms", "the game "+br.Game+" has no room "+br.Room)
			}
		}
		k := br.Game + "/" + br.Room
		if !seen[k] {
			seen[k] = true
			out = append(out, br)
		}
	}
	return out, nil
}

func cleanBotLook(raw map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for k, v := range raw {
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		if len(k) > 64 || len(v) > 128 {
			return nil, badRequest("INVALID_LOOK", "look", "slot and item keys are at most 64 and 128 characters")
		}
		out[k] = v
	}
	return out, nil
}

type botPut struct {
	Enabled     *bool             `json:"enabled"`
	Level       *int              `json:"level"`
	Template    *string           `json:"template"`
	Rooms       *[]botRoom        `json:"rooms"`
	Look        map[string]string `json:"look"`
	Note        *string           `json:"note"`
	Username    *string           `json:"username"`
	DisplayName *string           `json:"displayName"`
	AvatarURL   *string           `json:"avatarUrl"`
	Wallet      map[string]int64  `json:"wallet"`
}

func (s *service) consoleBotPut(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in botPut
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.writeError(w, err)
		return
	}
	staff := pluginkit.ConsoleUser(r)
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var exists bool
		if err := tx.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM economy_bot WHERE user_id::text=$1)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return notFound("BOT_NOT_FOUND")
		}
		if in.Enabled != nil {
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET enabled=$2 WHERE user_id::text=$1`, id, *in.Enabled); err != nil {
				return err
			}
		}
		if in.Level != nil {
			if *in.Level < 1 || *in.Level > 1000 {
				return badRequest("INVALID_LEVEL", "level", "a level from 1 to 1000")
			}
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET level=$2 WHERE user_id::text=$1`, id, *in.Level); err != nil {
				return err
			}
		}
		if in.Template != nil {
			t := strings.TrimSpace(*in.Template)
			if len(t) > maxBotTemplate {
				return badRequest("INVALID_BOT_TEMPLATE", "template", "at most 64 characters")
			}
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET template=$2 WHERE user_id::text=$1`, id, t); err != nil {
				return err
			}
		}
		if in.Rooms != nil {
			rooms, err := cleanBotRooms(r.Context(), tx, *in.Rooms)
			if err != nil {
				return err
			}
			raw, _ := json.Marshal(rooms)
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET rooms=$2::jsonb WHERE user_id::text=$1`, id, string(raw)); err != nil {
				return err
			}
		}
		if in.Look != nil {
			look, err := cleanBotLook(in.Look)
			if err != nil {
				return err
			}
			raw, _ := json.Marshal(look)
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET look=$2::jsonb WHERE user_id::text=$1`, id, string(raw)); err != nil {
				return err
			}
		}
		if in.Note != nil {
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET note=$2 WHERE user_id::text=$1`, id, truncate(strings.TrimSpace(*in.Note), 500)); err != nil {
				return err
			}
		}
		if in.Username != nil {
			name, err := cleanBotUsername(*in.Username)
			if err != nil {
				return err
			}
			var taken bool
			if err := tx.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE username=$1 AND id::text<>$2)`, name, id).Scan(&taken); err != nil {
				return err
			}
			if taken {
				return conflict("USERNAME_TAKEN", "another account has this username")
			}
			if _, err := tx.ExecContext(r.Context(), `UPDATE users SET username=$2, update_time=now() WHERE id::text=$1`, id, name); err != nil {
				return err
			}
		}
		if in.DisplayName != nil {
			if _, err := tx.ExecContext(r.Context(), `UPDATE users SET display_name=NULLIF($2,''), update_time=now() WHERE id::text=$1`, id, truncate(strings.TrimSpace(*in.DisplayName), 255)); err != nil {
				return err
			}
		}
		if in.AvatarURL != nil {
			if _, err := tx.ExecContext(r.Context(), `UPDATE users SET avatar_url=NULLIF($2,''), update_time=now() WHERE id::text=$1`, id, truncate(strings.TrimSpace(*in.AvatarURL), 512)); err != nil {
				return err
			}
		}
		if in.Wallet != nil {
			if err := setBotWallet(r.Context(), tx, id, in.Wallet, "console:"+staff); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(r.Context(), `UPDATE economy_bot SET updated_at=now(), updated_by=$2 WHERE user_id::text=$1`, id, staff)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleBotGet(w, r)
}

func setBotWallet(ctx context.Context, tx *sql.Tx, userID string, target map[string]int64, grant string) error {
	var raw []byte
	if err := tx.QueryRowContext(ctx, `SELECT wallet FROM users WHERE id::text=$1 FOR UPDATE`, userID).Scan(&raw); err != nil {
		return err
	}
	wallet := map[string]int64{}
	_ = json.Unmarshal(raw, &wallet)
	changes := map[string]int64{}
	for k, v := range target {
		k = strings.TrimSpace(k)
		if k == "" || v < 0 {
			return badRequest("INVALID_WALLET", "wallet", "a currency key and an amount of 0 or more")
		}
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, k).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("CURRENCY_NOT_FOUND", "wallet", "no currency has the key "+k)
		}
		if d := v - wallet[k]; d != 0 {
			changes[k] = d
		}
	}
	if len(changes) == 0 {
		return nil
	}
	id, _ := uuid.NewV4()
	return pluginkit.WalletTx(ctx, tx, userID, grant+":"+id.String(), changes, map[string]any{"module": "economy", "reason": "bot_wallet", "sourceType": "bot"})
}

func (s *service) consoleBotStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs     []string `json:"ids"`
		Enabled bool     `json:"enabled"`
	}
	if err := pluginkit.ReadBody(r, &in, 1024*1024); err != nil {
		s.writeError(w, err)
		return
	}
	ids := []string{}
	for _, id := range in.IDs {
		if _, err := uuid.FromString(id); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		s.writeError(w, badRequest("IDS_REQUIRED", "ids", "pick one or more bots"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE economy_bot SET enabled=$2, updated_at=now(), updated_by=$3 WHERE user_id::text = ANY($1::text[])`, ids, in.Enabled, pluginkit.ConsoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, map[string]any{"updated": n, "summary": s.botSummary(r.Context())})
}

type botBand struct {
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Template string `json:"template"`
}

type botSpec struct {
	Username, DisplayName, Avatar, Template, Note string
	Level, Line                                   int
	Rooms                                         []botRoom
	Look                                          map[string]string
	Wallet                                        map[string]int64
}

func randInt(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo+1)))
	return lo + int(n.Int64())
}

func bandTemplate(bands []botBand, level int, fallback string) string {
	for _, b := range bands {
		if level >= b.Min && (b.Max == 0 || level <= b.Max) {
			return strings.TrimSpace(b.Template)
		}
	}
	return fallback
}

func (s *service) createBots(ctx context.Context, specs []botSpec, source, staff string) (int, []map[string]any, error) {
	created := 0
	problems := []map[string]any{}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		for i, sp := range specs {
			if sp.Line == 0 {
				sp.Line = i + 1
			}
			name, err := cleanBotUsername(sp.Username)
			if err != nil {
				problems = append(problems, map[string]any{"line": sp.Line, "username": sp.Username, "code": "INVALID_USERNAME"})
				continue
			}
			id, _ := uuid.NewV4()
			wallet := map[string]int64{}
			for k, v := range sp.Wallet {
				if v > 0 {
					wallet[k] = v
				}
			}
			walletJSON, _ := json.Marshal(wallet)
			res, err := tx.ExecContext(ctx, `INSERT INTO users (id, username, display_name, avatar_url, metadata, wallet, create_time, update_time) VALUES ($1,$2,NULLIF($3,''),NULLIF($4,''),'{"bot":true}',$5::jsonb,now(),now()) ON CONFLICT (username) DO NOTHING`,
				id.String(), name, truncate(sp.DisplayName, 255), truncate(sp.Avatar, 512), string(walletJSON))
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				problems = append(problems, map[string]any{"line": sp.Line, "username": name, "code": "USERNAME_TAKEN"})
				continue
			}
			level := sp.Level
			if level < 1 {
				level = 1
			}
			if level > 1000 {
				level = 1000
			}
			roomsJSON, _ := json.Marshal(sp.Rooms)
			if sp.Rooms == nil {
				roomsJSON = []byte("[]")
			}
			look := sp.Look
			if look == nil {
				look = map[string]string{}
			}
			lookJSON, _ := json.Marshal(look)
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_bot (user_id, enabled, level, template, rooms, look, note, source, updated_by) VALUES ($1,true,$2,$3,$4::jsonb,$5::jsonb,$6,$7,$8)`,
				id.String(), level, truncate(sp.Template, maxBotTemplate), string(roomsJSON), string(lookJSON), truncate(sp.Note, 500), source, staff); err != nil {
				return err
			}
			if len(wallet) > 0 {
				changeJSON, _ := json.Marshal(wallet)
				meta, _ := json.Marshal(map[string]any{"module": "economy", "reason": "bot_wallet", "sourceType": "bot", "grantId": "bot-create:" + id.String()})
				if _, err := tx.ExecContext(ctx, `INSERT INTO wallet_ledger (id, user_id, changeset, metadata) VALUES (gen_random_uuid(), $1, $2, $3)`, id.String(), changeJSON, meta); err != nil {
					return err
				}
			}
			created++
		}
		return nil
	})
	return created, problems, err
}

func (s *service) currencyKeys(ctx context.Context) map[string]bool {
	out := map[string]bool{}
	rows, err := s.db.QueryContext(ctx, `SELECT key FROM economy_currency`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		if rows.Scan(&k) == nil {
			out[k] = true
		}
	}
	return out
}

func (s *service) consoleBotBulk(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Count    int               `json:"count"`
		Prefix   string            `json:"prefix"`
		Names    []string          `json:"names"`
		LevelMin int               `json:"levelMin"`
		LevelMax int               `json:"levelMax"`
		Template string            `json:"template"`
		Bands    []botBand         `json:"bands"`
		Rooms    []botRoom         `json:"rooms"`
		Wallet   map[string]int64  `json:"wallet"`
		Look     map[string]string `json:"look"`
		Note     string            `json:"note"`
	}
	if err := pluginkit.ReadBody(r, &in, maxBotImport); err != nil {
		s.writeError(w, err)
		return
	}
	names := []string{}
	for _, n := range in.Names {
		if n = strings.TrimSpace(n); n != "" {
			names = append(names, n)
		}
	}
	count := in.Count
	if len(names) > 0 {
		count = len(names)
	}
	if count < 1 || count > maxBotBulk {
		s.writeError(w, badRequest("INVALID_COUNT", "count", "from 1 to 5000 bots at a time"))
		return
	}
	prefix := strings.TrimSpace(in.Prefix)
	if len(names) == 0 && prefix == "" {
		s.writeError(w, badRequest("PREFIX_REQUIRED", "prefix", "give a name prefix or a list of names"))
		return
	}
	lo, hi := in.LevelMin, in.LevelMax
	if lo < 1 {
		lo = 1
	}
	if hi < lo {
		hi = lo
	}
	if hi > 1000 {
		s.writeError(w, badRequest("INVALID_LEVEL", "levelMax", "a level from 1 to 1000"))
		return
	}
	rooms, err := cleanBotRooms(r.Context(), s.db, in.Rooms)
	if err != nil {
		s.writeError(w, err)
		return
	}
	look, err := cleanBotLook(in.Look)
	if err != nil {
		s.writeError(w, err)
		return
	}
	cur := s.currencyKeys(r.Context())
	for k, v := range in.Wallet {
		if !cur[k] || v < 0 {
			s.writeError(w, badRequest("CURRENCY_NOT_FOUND", "wallet", "no currency has the key "+k))
			return
		}
	}
	specs := make([]botSpec, 0, count)
	for i := 0; i < count; i++ {
		name := ""
		if len(names) > 0 {
			name = names[i]
		} else {
			name = prefix + strconv.Itoa(randInt(1000, 99999999))
		}
		level := randInt(lo, hi)
		specs = append(specs, botSpec{Username: name, Level: level, Template: bandTemplate(in.Bands, level, strings.TrimSpace(in.Template)), Rooms: rooms, Look: look, Wallet: in.Wallet, Note: in.Note})
	}
	created, problems, err := s.createBots(r.Context(), specs, "bulk", pluginkit.ConsoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"created": created, "problems": problems, "summary": s.botSummary(r.Context())})
}

func (s *service) consoleBotImport(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CSV      string    `json:"csv"`
		Template string    `json:"template"`
		Bands    []botBand `json:"bands"`
		LevelMin int       `json:"levelMin"`
		LevelMax int       `json:"levelMax"`
		DryRun   bool      `json:"dryRun"`
	}
	if err := pluginkit.ReadBody(r, &in, maxBotImport+4096); err != nil {
		s.writeError(w, err)
		return
	}
	rd := csv.NewReader(strings.NewReader(strings.TrimPrefix(in.CSV, "\uFEFF")))
	rd.FieldsPerRecord = -1
	records, err := rd.ReadAll()
	if err != nil {
		s.writeError(w, badRequest("INVALID_CSV", "csv", err.Error()))
		return
	}
	if len(records) < 2 {
		s.writeError(w, badRequest("INVALID_CSV", "csv", "a header row and at least one bot"))
		return
	}
	if len(records)-1 > maxBotBulk {
		s.writeError(w, badRequest("TOO_MANY_ROWS", "csv", "at most 5000 bots per file"))
		return
	}
	head := map[string]int{}
	for i, h := range records[0] {
		head[strings.ToLower(strings.TrimSpace(h))] = i
	}
	if _, ok := head["username"]; !ok {
		s.writeError(w, badRequest("USERNAME_COLUMN_REQUIRED", "csv", "the header needs a username column"))
		return
	}
	cur := s.currencyKeys(r.Context())
	lo, hi := in.LevelMin, in.LevelMax
	if lo < 1 {
		lo = 1
	}
	if hi < lo {
		hi = lo
	}
	get := func(row []string, k string) string {
		if i, ok := head[k]; ok && i < len(row) {
			return strings.TrimSpace(row[i])
		}
		return ""
	}
	specs := []botSpec{}
	problems := []map[string]any{}
	for n, row := range records[1:] {
		line := n + 2
		sp := botSpec{Line: line, Username: get(row, "username"), DisplayName: get(row, "displayname"), Avatar: get(row, "avatarurl"), Note: get(row, "note"), Wallet: map[string]int64{}, Look: map[string]string{}}
		if v := get(row, "level"); v != "" {
			lv, err := strconv.Atoi(v)
			if err != nil || lv < 1 || lv > 1000 {
				problems = append(problems, map[string]any{"line": line, "username": sp.Username, "code": "INVALID_LEVEL"})
				continue
			}
			sp.Level = lv
		} else {
			sp.Level = randInt(lo, hi)
		}
		sp.Template = get(row, "template")
		if sp.Template == "" {
			sp.Template = bandTemplate(in.Bands, sp.Level, strings.TrimSpace(in.Template))
		}
		bad := ""
		for k, i := range head {
			if i >= len(row) || strings.TrimSpace(row[i]) == "" {
				continue
			}
			v := strings.TrimSpace(row[i])
			switch {
			case cur[k]:
				amount, err := strconv.ParseInt(v, 10, 64)
				if err != nil || amount < 0 {
					bad = "INVALID_WALLET"
				}
				sp.Wallet[k] = amount
			case strings.HasPrefix(k, "look."):
				sp.Look[strings.TrimPrefix(k, "look.")] = v
			case k == "rooms":
				for _, part := range strings.Split(v, ";") {
					gr := strings.SplitN(strings.TrimSpace(part), ":", 2)
					if gr[0] == "" {
						continue
					}
					br := botRoom{Game: gr[0]}
					if len(gr) == 2 {
						br.Room = gr[1]
					}
					sp.Rooms = append(sp.Rooms, br)
				}
			}
		}
		if bad != "" {
			problems = append(problems, map[string]any{"line": line, "username": sp.Username, "code": bad})
			continue
		}
		if _, err := cleanBotUsername(sp.Username); err != nil {
			problems = append(problems, map[string]any{"line": line, "username": sp.Username, "code": "INVALID_USERNAME"})
			continue
		}
		if sp.Rooms != nil {
			rooms, err := cleanBotRooms(r.Context(), s.db, sp.Rooms)
			if err != nil {
				problems = append(problems, map[string]any{"line": line, "username": sp.Username, "code": codeOf(err)})
				continue
			}
			sp.Rooms = rooms
		}
		specs = append(specs, sp)
	}
	if in.DryRun {
		writeJSON(w, http.StatusOK, map[string]any{"valid": len(specs), "problems": problems, "dryRun": true})
		return
	}
	created, more, err := s.createBots(r.Context(), specs, "import", pluginkit.ConsoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"created": created, "problems": append(problems, more...), "summary": s.botSummary(r.Context())})
}

func codeOf(err error) string {
	if app, ok := asApp(err); ok {
		return app.Code
	}
	return "ERROR"
}

func (s *service) consoleBotTopup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs      []string `json:"ids"`
		All      bool     `json:"all"`
		Currency string   `json:"currency"`
		Below    int64    `json:"below"`
		Amount   int64    `json:"amount"`
	}
	if err := pluginkit.ReadBody(r, &in, 1024*1024); err != nil {
		s.writeError(w, err)
		return
	}
	in.Currency = strings.TrimSpace(in.Currency)
	if !s.currencyKeys(r.Context())[in.Currency] {
		s.writeError(w, badRequest("CURRENCY_NOT_FOUND", "currency", "pick a currency"))
		return
	}
	if in.Amount <= 0 || in.Amount > 1_000_000_000_000 {
		s.writeError(w, badRequest("INVALID_AMOUNT", "amount", "a whole amount above 0"))
		return
	}
	if in.Below <= 0 || in.Below > in.Amount {
		in.Below = in.Amount
	}
	ids := []string{}
	for _, id := range in.IDs {
		if _, err := uuid.FromString(id); err == nil {
			ids = append(ids, id)
		}
	}
	if !in.All && len(ids) == 0 {
		s.writeError(w, badRequest("IDS_REQUIRED", "ids", "pick bots or all"))
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT b.user_id::text FROM economy_bot b JOIN users u ON u.id=b.user_id WHERE ($1 OR b.user_id::text = ANY($2::text[])) AND COALESCE((u.wallet->>$3)::bigint,0) < $4`, in.All, ids, in.Currency, in.Below)
	if err != nil {
		s.writeError(w, err)
		return
	}
	targets := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			targets = append(targets, id)
		}
	}
	rows.Close()
	staff := pluginkit.ConsoleUser(r)
	done := 0
	for _, id := range targets {
		if err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			return setBotWallet(r.Context(), tx, id, map[string]int64{in.Currency: in.Amount}, "console-topup:"+staff)
		}); err != nil {
			s.writeError(w, err)
			return
		}
		done++
	}
	writeJSON(w, http.StatusOK, map[string]any{"toppedUp": done})
}
