package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeRaw(w http.ResponseWriter, status int, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func (s *service) writeError(w http.ResponseWriter, err error) {
	if app, ok := asApp(err); ok {
		body := map[string]string{"error": app.Code, "message": app.Message}
		if app.Message == "" {
			body["message"] = app.Code
		}
		if app.Field != "" {
			body["field"] = app.Field
		}
		writeJSON(w, app.Status, body)
		return
	}
	s.logger.Error("Publication console request failed: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR", "message": "internal error"})
}

func readBody(r *http.Request, out any) error {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return badRequest("INVALID_JSON", "", "cannot read body")
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return badRequest("INVALID_JSON", "", "request body is not valid JSON: "+err.Error())
	}
	return nil
}

func pathVar(r *http.Request, name string) string {
	if v := r.PathValue(name); v != "" {
		return v
	}
	return mux.Vars(r)[name]
}

type pageQuery struct {
	Page   int
	Limit  int
	Offset int
	Q      string
}

func readPage(r *http.Request) pageQuery {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return pageQuery{Page: page, Limit: limit, Offset: (page - 1) * limit, Q: strings.TrimSpace(q.Get("q"))}
}

func likePattern(q string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(q) + "%"
}

type sqlWhere struct {
	parts []string
	args  []any
}

func (w *sqlWhere) add(expr string, args ...any) {
	for _, a := range args {
		w.args = append(w.args, a)
		expr = strings.Replace(expr, "?", "$"+strconv.Itoa(len(w.args)), 1)
	}
	w.parts = append(w.parts, expr)
}

func (w *sqlWhere) sql() string {
	if len(w.parts) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(w.parts, " AND ")
}

func (w *sqlWhere) next() string {
	return "$" + strconv.Itoa(len(w.args)+1)
}
