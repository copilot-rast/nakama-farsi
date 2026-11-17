package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/chatfilter"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const (
	artEmblem = "emblem"
	artCover  = "cover"

	maxUploadPayload = 3 << 20
	filesPath        = "/v2/clan/files/"
)

var (
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
	filePattern    = regexp.MustCompile(`^[0-9a-f]{32}\.(png|jpg|webp)$`)
	imageTypes     = map[string]string{"png": "image/png", "jpg": "image/jpeg", "webp": "image/webp"}
)

func openStore(env map[string]string, args []string) (objectstore.Store, string, error) {
	dir := strings.TrimSpace(env["CLAN_UPLOAD_DIR"])
	if dir == "" {
		if base := strings.TrimSpace(env["ECONOMY_UPLOAD_DIR"]); base != "" {
			dir = filepath.Join(base, "clan")
		} else {
			dir = filepath.Join(dataDir(args), "clan-uploads")
		}
	}
	cfg, err := objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: dir, Prefix: "clan/"})
	if err != nil {
		return nil, "", err
	}
	store, err := objectstore.New(cfg, nil)
	if err != nil {
		return nil, "", err
	}
	return store, strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/"), nil
}

func dataDir(args []string) string {
	for i, a := range args {
		if (a == "--data_dir" || a == "-data_dir") && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, "--data_dir=") {
			return strings.TrimPrefix(a, "--data_dir=")
		}
	}
	return "./data"
}

type art struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"`
	Key      string            `json:"key"`
	Name     map[string]string `json:"name"`
	ImageURL string            `json:"imageUrl"`
	Sort     int               `json:"sort"`
	Active   bool              `json:"isActive"`
	MinLevel int               `json:"minLevel"`
	Currency string            `json:"priceCurrency"`
	Amount   int64             `json:"priceAmount"`
	Default  bool              `json:"isDefault"`
}

const artCols = `id::text, kind, key, name::text, image_url, sort, is_active, min_level, price_currency, price_amount, is_default`

func scanArt(row interface{ Scan(...any) error }) (*art, error) {
	a := &art{}
	var name string
	if err := row.Scan(&a.ID, &a.Kind, &a.Key, &name, &a.ImageURL, &a.Sort, &a.Active, &a.MinLevel, &a.Currency, &a.Amount, &a.Default); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(name), &a.Name)
	if a.Name == nil {
		a.Name = map[string]string{}
	}
	return a, nil
}

func (s *service) arts(ctx context.Context, activeOnly bool) ([]*art, error) {
	q := `SELECT ` + artCols + ` FROM clan_art`
	if activeOnly {
		q += ` WHERE is_active`
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY kind, sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*art{}
	for rows.Next() {
		a, err := scanArt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *service) artByKey(ctx context.Context, kind, key string) (*art, error) {
	a, err := scanArt(s.db.QueryRowContext(ctx, `SELECT `+artCols+` FROM clan_art WHERE kind=$1 AND key=$2 AND is_active`, kind, key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("CLAN_ART_NOT_FOUND").WithField(kind + "Key")
	}
	return a, err
}

func (s *service) defaultArt(ctx context.Context, kind string) *art {
	a, err := scanArt(s.db.QueryRowContext(ctx, `SELECT `+artCols+` FROM clan_art WHERE kind=$1 AND is_active AND is_default ORDER BY sort LIMIT 1`, kind))
	if err != nil {
		return nil
	}
	return a
}

func (s *service) wordFilter(ctx context.Context, cfg settings.All) *chatfilter.Filter {
	rules := []chatfilter.Rule{}
	for i, w := range cfg.Profile.BlockedWords {
		rules = append(rules, chatfilter.Rule{ID: "s" + strconv.Itoa(i), Word: w, Action: chatfilter.ActionBlock, Match: chatfilter.MatchAny})
	}
	if cfg.Profile.UseChatWords {
		rows, err := s.db.QueryContext(ctx, `SELECT id::text, word, match FROM base_chat_word WHERE is_active LIMIT 5000`)
		if err == nil {
			for rows.Next() {
				var id, word, match string
				if rows.Scan(&id, &word, &match) == nil {
					rules = append(rules, chatfilter.Rule{ID: id, Word: word, Action: chatfilter.ActionBlock, Match: match})
				}
			}
			rows.Close()
		}
	}
	return chatfilter.Compile(rules)
}

func blocked(f *chatfilter.Filter, text string) bool {
	if f == nil || f.Len() == 0 || strings.TrimSpace(text) == "" {
		return false
	}
	r := f.Apply(text)
	return r.Blocked || r.Changed || len(r.Hits) > 0
}

type lookInput struct {
	Motto     *string   `json:"motto"`
	Color     *string   `json:"color"`
	Country   *string   `json:"country"`
	Tags      *[]string `json:"tags"`
	EmblemKey *string   `json:"emblemKey"`
	CoverKey  *string   `json:"coverKey"`
}

type editor struct {
	userID string
	admin  bool
}

type lookPlan struct {
	name, lang, description string
	avatar, cover           *string
	emblemKey, coverKey     *string
	motto, color, country   string
	tags                    []string
	changed                 bool
	buy                     []*art
}

func (s *service) planProfile(ctx context.Context, cfg settings.All, c *clanRow, in profileInput, who editor) (*lookPlan, error) {
	if !who.admin && cfg.Profile.ChangeCooldownMinutes > 0 && c.LookChangedAt != nil {
		if left := c.LookChangedAt.Add(time.Duration(cfg.Profile.ChangeCooldownMinutes) * time.Minute).Sub(s.now()); left > 0 {
			return nil, apierr.Conflict("CLAN_CHANGE_COOLDOWN").With("retryAfter", int(math.Ceil(left.Seconds())))
		}
	}
	if !who.admin && in.AvatarURL != nil {
		return nil, apierr.BadRequest("CLAN_AVATAR_USE_GALLERY").WithField("avatarUrl")
	}
	p := &lookPlan{motto: c.Motto, color: c.Color, country: c.Country, tags: c.Tags}
	f := s.wordFilter(ctx, cfg)
	if in.Name != nil {
		n, err := cleanName(cfg, *in.Name)
		if err != nil {
			return nil, err
		}
		if blocked(f, n) {
			return nil, apierr.BadRequest("CLAN_NAME_BLOCKED").WithField("name")
		}
		if !strings.EqualFold(n, c.Name) || n != c.Name {
			var taken bool
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM groups WHERE lower(name)=lower($1) AND id::text<>$2)`, n, c.ID).Scan(&taken)
			if taken {
				return nil, apierr.Conflict("CLAN_NAME_TAKEN").WithField("name")
			}
			if n != c.Name {
				p.name, p.changed = n, true
			}
		}
	}
	if in.Description != nil {
		d := strings.TrimSpace(*in.Description)
		if blocked(f, d) {
			return nil, apierr.BadRequest("CLAN_TEXT_BLOCKED").WithField("description")
		}
		if d != c.Description {
			p.description, p.changed = d, true
		}
	}
	if in.Lang != nil && strings.TrimSpace(*in.Lang) != "" {
		l := strings.TrimSpace(*in.Lang)
		var ok bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_language WHERE code=$1 AND is_active)`, l).Scan(&ok)
		if !ok {
			return nil, apierr.BadRequest("CLAN_LANG_INVALID").WithField("lang")
		}
		if l != c.Lang {
			p.lang, p.changed = l, true
		}
	}
	if in.Motto != nil {
		m := strings.Join(strings.Fields(*in.Motto), " ")
		if utf8.RuneCountInString(m) > cfg.Profile.MottoMax {
			return nil, apierr.BadRequest("CLAN_MOTTO_TOO_LONG").WithField("motto").With("max", cfg.Profile.MottoMax)
		}
		if blocked(f, m) {
			return nil, apierr.BadRequest("CLAN_TEXT_BLOCKED").WithField("motto")
		}
		p.changed = p.changed || m != c.Motto
		p.motto = m
	}
	if in.Color != nil {
		col := strings.ToUpper(strings.TrimSpace(*in.Color))
		if col != "" && (!settings.ColorPattern.MatchString(col) || (len(cfg.Profile.Colors) > 0 && !containsFold(cfg.Profile.Colors, col))) {
			return nil, apierr.BadRequest("CLAN_COLOR_INVALID").WithField("color")
		}
		p.changed = p.changed || col != c.Color
		p.color = col
	}
	if in.Country != nil {
		co := strings.ToUpper(strings.TrimSpace(*in.Country))
		if co != "" {
			var ok bool
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_country WHERE code=$1 AND is_active)`, co).Scan(&ok)
			if !countryPattern.MatchString(co) || !ok {
				return nil, apierr.BadRequest("CLAN_COUNTRY_INVALID").WithField("country")
			}
		}
		p.changed = p.changed || co != c.Country
		p.country = co
	}
	if in.Tags != nil {
		allowedTags := map[string]bool{}
		for _, t := range cfg.Profile.Tags {
			allowedTags[t.Key] = true
		}
		out := []string{}
		for _, t := range *in.Tags {
			t = strings.TrimSpace(t)
			if !allowedTags[t] {
				return nil, apierr.BadRequest("CLAN_TAG_INVALID").WithField("tags").With("tag", t)
			}
			if !allowed(out, t) {
				out = append(out, t)
			}
		}
		if len(out) > cfg.Profile.MaxTags {
			return nil, apierr.BadRequest("CLAN_TOO_MANY_TAGS").WithField("tags").With("max", cfg.Profile.MaxTags)
		}
		p.changed = p.changed || strings.Join(out, ",") != strings.Join(c.Tags, ",")
		p.tags = out
	}
	for _, pick := range []struct {
		kind string
		key  *string
		cur  string
	}{{artEmblem, in.EmblemKey, c.EmblemKey}, {artCover, in.CoverKey, c.CoverKey}} {
		if pick.key == nil || strings.TrimSpace(*pick.key) == pick.cur {
			continue
		}
		key := strings.TrimSpace(*pick.key)
		url := ""
		if key != "" {
			a, err := s.artByKey(ctx, pick.kind, key)
			if err != nil {
				return nil, err
			}
			if !who.admin && c.Level < a.MinLevel {
				return nil, apierr.Forbidden("CLAN_ART_LOCKED").WithField(pick.kind + "Key").With("level", a.MinLevel)
			}
			if !who.admin && a.Amount > 0 && a.Currency != "" && !s.owns(ctx, c.ID, a.ID) {
				p.buy = append(p.buy, a)
			}
			url = a.ImageURL
		}
		k, u := key, url
		if pick.kind == artEmblem {
			p.emblemKey, p.avatar = &k, &u
		} else {
			p.coverKey, p.cover = &k, &u
		}
		p.changed = true
	}
	if !who.admin && in.AvatarURL == nil && cfg.Profile.RequireEmblem && c.EmblemKey == "" && c.AvatarURL == "" && (p.emblemKey == nil || *p.emblemKey == "") {
		return nil, apierr.BadRequest("CLAN_EMBLEM_REQUIRED").WithField("emblemKey")
	}
	if who.admin && in.AvatarURL != nil {
		u := strings.TrimSpace(*in.AvatarURL)
		e := ""
		p.avatar, p.emblemKey, p.changed = &u, &e, true
	}
	return p, nil
}

func containsFold(list []string, v string) bool {
	for _, x := range list {
		if strings.EqualFold(x, v) {
			return true
		}
	}
	return false
}

func (s *service) owns(ctx context.Context, clanID, artID string) bool {
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_art_owned WHERE group_id::text=$1 AND art_id::text=$2)`, clanID, artID).Scan(&ok)
	return ok
}

func (s *service) applyPlan(ctx context.Context, c *clanRow, joinType string, p *lookPlan, who editor) error {
	for _, a := range p.buy {
		if err := s.debit(ctx, who.userID, "clan-art:"+c.ID+":"+a.ID, map[string]int64{a.Currency: a.Amount}); err != nil {
			var ins *pluginkit.ErrInsufficient
			if errors.As(err, &ins) {
				return apierr.Conflict("CLAN_NOT_ENOUGH_CURRENCY").With("currency", ins.Path).With("amount", a.Amount)
			}
			return err
		}
		var by any
		if who.userID != "" {
			by = who.userID
		}
		if _, err := s.db.ExecContext(ctx, `INSERT INTO clan_art_owned (group_id, art_id, by_user) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, c.ID, a.ID, by); err != nil {
			return err
		}
	}
	cover := c.CoverURL
	if p.cover != nil {
		cover = *p.cover
	}
	avatar := ""
	if p.avatar != nil && *p.avatar != "" {
		avatar = *p.avatar
	}
	meta := map[string]any{"clan": true, "cover": cover, "color": p.color, "country": p.country, "tags": p.tags, "motto": p.motto}
	if err := s.groups.Update(ctx, c.ID, p.name, p.lang, p.description, avatar, joinType == settings.JoinOpen, 0, meta); err != nil {
		if errors.Is(err, runtime.ErrGroupNameInUse) {
			return apierr.Conflict("CLAN_NAME_TAKEN").WithField("name")
		}
		if !errors.Is(err, runtime.ErrGroupNoUpdateOps) && !errors.Is(err, runtime.ErrGroupNotUpdated) {
			return err
		}
	}
	if p.avatar != nil && *p.avatar == "" {
		_, _ = s.db.ExecContext(ctx, `UPDATE groups SET avatar_url='' WHERE id::text=$1`, c.ID)
	}
	emblemKey, coverKey := c.EmblemKey, c.CoverKey
	if p.emblemKey != nil {
		emblemKey = *p.emblemKey
	}
	if p.coverKey != nil {
		coverKey = *p.coverKey
	}
	_, err := s.db.ExecContext(ctx, `UPDATE clan SET motto=$2, color=$3, country=$4, tags=$5, emblem_key=$6, cover_key=$7, cover_url=$8,
		look_changed_at=CASE WHEN $9 THEN now() ELSE look_changed_at END, updated_at=now() WHERE group_id::text=$1`,
		c.ID, p.motto, p.color, p.country, p.tags, emblemKey, coverKey, cover, p.changed && !who.admin)
	return err
}

func (s *service) fileURL(name string) string {
	if u := s.store.PublicURL(name); u != "" {
		return u
	}
	return s.publicBase + filesPath + name
}

func sniff(data []byte) string {
	switch http.DetectContentType(data) {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	}
	return ""
}

func (s *service) rpcImageUpload(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Kind string `json:"kind"`
		Data string `json:"data"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if !cfg.Uploads.Enabled || s.store == nil {
		return nil, apierr.Forbidden("CLAN_UPLOADS_DISABLED")
	}
	if in.Kind != artEmblem && in.Kind != artCover {
		return nil, apierr.BadRequest("CLAN_UPLOAD_KIND_INVALID").WithField("kind")
	}
	m, err := s.requireRole(ctx, userID, cfg.Roles.Edit)
	if err != nil {
		return nil, err
	}
	c, err := s.clanByID(ctx, s.db, m.ClanID)
	if err != nil {
		return nil, err
	}
	if cfg.Profile.ChangeCooldownMinutes > 0 && c.LookChangedAt != nil {
		if left := c.LookChangedAt.Add(time.Duration(cfg.Profile.ChangeCooldownMinutes) * time.Minute).Sub(s.now()); left > 0 {
			return nil, apierr.Conflict("CLAN_CHANGE_COOLDOWN").With("retryAfter", int(math.Ceil(left.Seconds())))
		}
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in.Data))
	if err != nil || len(data) == 0 {
		return nil, apierr.BadRequest("CLAN_UPLOAD_INVALID").WithField("data")
	}
	if len(data) > cfg.Uploads.MaxBytes {
		return nil, apierr.BadRequest("CLAN_UPLOAD_TOO_LARGE").WithField("data").With("max", cfg.Uploads.MaxBytes)
	}
	ext := sniff(data)
	if ext == "" || !allowed(cfg.Uploads.Types, ext) {
		return nil, apierr.BadRequest("CLAN_UPLOAD_TYPE").WithField("data").With("types", cfg.Uploads.Types)
	}
	raw := make([]byte, 16)
	_, _ = rand.Read(raw)
	name := hex.EncodeToString(raw) + "." + ext
	if err := s.store.Put(ctx, name, data, imageTypes[ext]); err != nil {
		return nil, err
	}
	status := "approved"
	if cfg.Uploads.NeedsApproval {
		status = "pending"
	}
	id := uuid.Must(uuid.NewV4()).String()
	s.replacePending(ctx, c.ID, in.Kind)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO clan_upload (id, group_id, kind, file_name, content_type, size, status, uploaded_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, c.ID, in.Kind, name, imageTypes[ext], len(data), status, userID); err != nil {
		return nil, err
	}
	var url any
	if status == "approved" {
		if err := s.useUpload(ctx, c.ID, in.Kind, name, false); err != nil {
			return nil, err
		}
		url = s.fileURL(name)
	}
	return map[string]any{"uploadId": id, "kind": in.Kind, "status": status, "url": url}, nil
}

func (s *service) replacePending(ctx context.Context, clanID, kind string) {
	rows, err := s.db.QueryContext(ctx, `UPDATE clan_upload SET status='replaced', reviewed_at=now() WHERE group_id::text=$1 AND kind=$2 AND status='pending' RETURNING file_name`, clanID, kind)
	if err != nil {
		return
	}
	names := []string{}
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	rows.Close()
	for _, n := range names {
		_ = s.store.Delete(ctx, n)
	}
}

func (s *service) useUpload(ctx context.Context, clanID, kind, name string, admin bool) error {
	c, err := s.clanByID(ctx, s.db, clanID)
	if err != nil {
		return err
	}
	url := s.fileURL(name)
	p := &lookPlan{motto: c.Motto, color: c.Color, country: c.Country, tags: c.Tags, changed: true}
	empty := ""
	if kind == artEmblem {
		p.avatar, p.emblemKey = &url, &empty
	} else {
		p.cover, p.coverKey = &url, &empty
	}
	return s.applyPlan(ctx, c, c.JoinType, p, editor{admin: admin})
}

func (s *service) pendingUploads(ctx context.Context, clanID string) map[string]any {
	out := map[string]any{"emblem": false, "cover": false}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT kind FROM clan_upload WHERE group_id::text=$1 AND status='pending'`, clanID)
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

func (s *service) filesHTTP(w http.ResponseWriter, r *http.Request) {
	name := pluginkit.PathVar(r, "name")
	if !filePattern.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	var ok bool
	_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM clan_upload WHERE file_name=$1 AND status='approved')`, name).Scan(&ok)
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.serveFile(w, r, name, "public, max-age=86400")
}

func (s *service) serveFile(w http.ResponseWriter, r *http.Request, name, cache string) {
	obj, err := s.store.Get(r.Context(), name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer obj.Body.Close()
	w.Header().Set("Content-Type", imageTypes[strings.TrimPrefix(filepath.Ext(name), ".")])
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", cache)
	_, _ = io.Copy(w, io.LimitReader(obj.Body, 4<<20))
}

func (s *service) rpcOptions(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	tx := s.texts(ctx, userID, in.Lang)
	level, clanID := 0, ""
	var changed *time.Time
	if m, _ := s.memberOf(ctx, s.db, userID); m != nil {
		clanID = m.ClanID
		_ = s.db.QueryRowContext(ctx, `SELECT level, look_changed_at FROM clan WHERE group_id::text=$1`, m.ClanID).Scan(&level, &changed)
	}
	owned := map[string]bool{}
	if clanID != "" {
		rows, err := s.db.QueryContext(ctx, `SELECT art_id::text FROM clan_art_owned WHERE group_id::text=$1`, clanID)
		if err == nil {
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					owned[id] = true
				}
			}
			rows.Close()
		}
	}
	list, err := s.arts(ctx, true)
	if err != nil {
		return nil, err
	}
	emblems, covers := []map[string]any{}, []map[string]any{}
	for _, a := range list {
		var price any
		if a.Amount > 0 && a.Currency != "" {
			price = map[string]any{"currencyKey": a.Currency, "amount": a.Amount}
		}
		row := map[string]any{"key": a.Key, "name": tx.of(a.Name, a.Key), "imageUrl": a.ImageURL, "minLevel": a.MinLevel, "price": price, "owned": owned[a.ID] || price == nil,
			"locked": clanID != "" && level < a.MinLevel, "isDefault": a.Default}
		if a.Kind == artEmblem {
			emblems = append(emblems, row)
		} else {
			covers = append(covers, row)
		}
	}
	tags := make([]map[string]any, 0, len(cfg.Profile.Tags))
	for _, t := range cfg.Profile.Tags {
		tags = append(tags, map[string]any{"key": t.Key, "name": tx.of(t.Name, t.Key)})
	}
	countries := []map[string]any{}
	if rows, err := s.db.QueryContext(ctx, `SELECT code, name::text FROM base_country WHERE is_active ORDER BY sort, code`); err == nil {
		for rows.Next() {
			var code, raw string
			if rows.Scan(&code, &raw) == nil {
				var n map[string]string
				_ = json.Unmarshal([]byte(raw), &n)
				countries = append(countries, map[string]any{"code": code, "name": tx.of(n, code)})
			}
		}
		rows.Close()
	}
	languages := []map[string]any{}
	if rows, err := s.db.QueryContext(ctx, `SELECT code, COALESCE(native_name, code) FROM base_language WHERE is_active ORDER BY is_default DESC, code`); err == nil {
		for rows.Next() {
			var code, name string
			if rows.Scan(&code, &name) == nil {
				languages = append(languages, map[string]any{"code": code, "name": name})
			}
		}
		rows.Close()
	}
	var next any
	if changed != nil && cfg.Profile.ChangeCooldownMinutes > 0 {
		if t := changed.Add(time.Duration(cfg.Profile.ChangeCooldownMinutes) * time.Minute); t.After(s.now()) {
			next = t.UTC().Format(time.RFC3339)
		}
	}
	return map[string]any{"emblems": emblems, "covers": covers, "tags": tags, "colors": cfg.Profile.Colors, "countries": countries, "languages": languages,
		"uploads": map[string]any{"enabled": cfg.Uploads.Enabled, "needsApproval": cfg.Uploads.NeedsApproval, "maxBytes": cfg.Uploads.MaxBytes, "types": cfg.Uploads.Types},
		"nameMin": cfg.Create.NameMin, "nameMax": cfg.Create.NameMax, "descriptionMax": cfg.Create.DescriptionMax, "mottoMax": cfg.Profile.MottoMax, "maxTags": cfg.Profile.MaxTags,
		"maxMinTrophies": cfg.Create.MaxMinTrophies, "createCost": cfg.Create.Cost, "changeCooldownMinutes": cfg.Profile.ChangeCooldownMinutes, "nextChangeAt": next,
		"editRoles": cfg.Roles.Edit}, nil
}
