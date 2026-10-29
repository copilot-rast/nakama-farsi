package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type resource struct {
	name     string
	table    string
	keyTypes []string
	usedSQL  string
}

const taxonomyJSON = `jsonb_build_object('id',t.id,'key',t.key,'keyType',t.key_type,'name',t.name,'description',t.description,'iconUrl',t.icon_url,'metadata',t.metadata,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`

const categoryUsedSQL = `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE category_id=$1) OR EXISTS (SELECT 1 FROM economy_product WHERE category_id=$1)
	OR EXISTS (SELECT 1 FROM publication_faq WHERE category_id=$1) OR EXISTS (SELECT 1 FROM publication_news WHERE category_id=$1)`

const tagUsedSQL = `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE $1::text = ANY(tag_ids)) OR EXISTS (SELECT 1 FROM economy_product WHERE $1::text = ANY(tag_ids))`

var categoriesResource = resource{name: "categories", table: "base_category", keyTypes: categoryKeyTypes, usedSQL: categoryUsedSQL}

var tagsResource = resource{name: "tags", table: "base_tag", keyTypes: tagKeyTypes, usedSQL: tagUsedSQL}

func translatedSearch(col, n string) string {
	return "EXISTS (SELECT 1 FROM jsonb_each_text(" + col + ") e WHERE e.value ILIKE " + n + ")"
}

func itoa(n int) string { return strconv.Itoa(n) }

func (s *service) listRaw(r *http.Request, from, json, order string, where *sqlWhere, p pageQuery) (map[string]any, error) {
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*)`+from+where.sql(), where.args...).Scan(&total); err != nil {
		return nil, err
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+json+`::text`+from+where.sql()+` ORDER BY `+order+` LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(append([]any{}, where.args...), p.Limit, p.Offset)...)
	if err != nil {
		return nil, err
	}
	items, err := rawRows(rows)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit}, nil
}

func rawRows(rows *sql.Rows) ([]json.RawMessage, error) {
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		items = append(items, json.RawMessage(raw))
	}
	return items, rows.Err()
}

func (s *service) taxonomyList(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := readPage(r)
		where := &sqlWhere{}
		if p.Q != "" {
			n := where.next()
			where.args = append(where.args, likePattern(p.Q))
			where.parts = append(where.parts, "(t.key ILIKE "+n+" OR "+translatedSearch("t.name", n)+" OR "+translatedSearch("t.description", n)+")")
		}
		if v := strings.TrimSpace(r.URL.Query().Get("keyType")); v != "" {
			if !validKeyType(res.keyTypes, v) {
				s.writeError(w, badRequest("INVALID_KEY_TYPE", "keyType", "unknown key type"))
				return
			}
			where.add("t.key_type = ?", v)
		}
		out, err := s.listRaw(r, ` FROM `+res.table+` t`, taxonomyJSON, "t.created_at DESC, t.id DESC", where, p)
		if err != nil {
			s.writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func fetchTaxonomy(ctx context.Context, q queryer, res resource, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+taxonomyJSON+`::text FROM `+res.table+` t WHERE t.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) taxonomyGet(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := fetchTaxonomy(r.Context(), s.db, res, pathVar(r, "id"))
		if err != nil {
			s.writeError(w, err)
			return
		}
		writeRaw(w, http.StatusOK, raw)
	}
}

type taxonomyInput struct {
	Key         string          `json:"key"`
	KeyType     string          `json:"keyType"`
	Name        localized.Text  `json:"name"`
	Description localized.Text  `json:"description"`
	IconURL     string          `json:"iconUrl"`
	Metadata    json.RawMessage `json:"metadata"`
	Sort        int             `json:"sort"`
}

func (s *service) saveTaxonomy(ctx context.Context, tx *sql.Tx, res resource, id string, create bool, raw []byte) error {
	var in taxonomyInput
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	in.KeyType = strings.TrimSpace(in.KeyType)
	if !validKeyType(res.keyTypes, in.KeyType) {
		return badRequest("INVALID_KEY_TYPE", "keyType", "key type must be one of "+strings.Join(res.keyTypes, ", "))
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	langs, err := s.languages(ctx, tx)
	if err != nil {
		return err
	}
	name, err := cleanText(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanText(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		return err
	}
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	if create {
		_, err := tx.ExecContext(ctx, `INSERT INTO `+res.table+` (id, key, key_type, name, description, icon_url, metadata, sort) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id, key, in.KeyType, name.JSON(), desc.JSON(), icon, meta, in.Sort)
		return err
	}
	var oldType string
	err = tx.QueryRowContext(ctx, `SELECT key_type FROM `+res.table+` WHERE id=$1 FOR UPDATE`, id).Scan(&oldType)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound("NOT_FOUND")
	}
	if err != nil {
		return err
	}
	if oldType != in.KeyType && in.KeyType != "general" {
		var used bool
		if err := tx.QueryRowContext(ctx, res.usedSQL, id).Scan(&used); err != nil {
			return err
		}
		if used {
			return badRequest("IN_USE", "keyType", "a record in use cannot change key type")
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE `+res.table+` SET key=$2, key_type=$3, name=$4, description=$5, icon_url=$6, metadata=$7, sort=$8, updated_at=now() WHERE id=$1`,
		id, key, in.KeyType, name.JSON(), desc.JSON(), icon, meta, in.Sort)
	return err
}

func (s *service) taxonomyWrite(res resource, create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body json.RawMessage
		if err := readBody(r, &body); err != nil {
			s.writeError(w, err)
			return
		}
		id := pathVar(r, "id")
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		} else if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			if err := s.saveTaxonomy(r.Context(), tx, res, id, create, body); err != nil {
				return err
			}
			var err error
			out, err = fetchTaxonomy(r.Context(), tx, res, id)
			return err
		})
		if err != nil {
			s.writeError(w, mapWriteError(err))
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) taxonomyDelete(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathVar(r, "id")
		if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			var found string
			if err := tx.QueryRowContext(r.Context(), `SELECT id::text FROM `+res.table+` WHERE id=$1 FOR UPDATE`, id).Scan(&found); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return notFound("NOT_FOUND")
				}
				return err
			}
			if res.name == "tags" {
				for _, table := range []string{"economy_prize", "economy_product"} {
					if _, err := tx.ExecContext(r.Context(), `UPDATE `+table+` SET tag_ids = array_remove(tag_ids, $1::text), updated_at=now() WHERE $1::text = ANY(tag_ids)`, id); err != nil {
						return err
					}
				}
			} else {
				var used bool
				if err := tx.QueryRowContext(r.Context(), res.usedSQL, id).Scan(&used); err != nil {
					return err
				}
				if used {
					return conflict("IN_USE", "this record is referenced and cannot be deleted")
				}
			}
			_, err := tx.ExecContext(r.Context(), `DELETE FROM `+res.table+` WHERE id=$1`, id)
			return err
		})
		if err != nil {
			if isForeignKeyViolation(err) {
				err = conflict("IN_USE", "this record is referenced and cannot be deleted")
			}
			s.writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{})
	}
}

func (s *service) keyTypes(keys []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": keyTypeList(keys)})
	}
}

func mapWriteError(err error) error {
	if _, ok := isUniqueViolation(err); ok {
		return fail(http.StatusConflict, "ALREADY_EXISTS", "key", "a record with this key already exists")
	}
	if isForeignKeyViolation(err) {
		return badRequest("REFERENCE_NOT_FOUND", "", "a referenced record does not exist")
	}
	return err
}

func (s *service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
