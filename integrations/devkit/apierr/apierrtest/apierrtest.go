package apierrtest

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
)

func DB(errorFields map[string]any) *sql.DB {
	all := settings.Defaults()
	raw, _ := json.Marshal(all.Errors)
	cur := map[string]any{}
	_ = json.Unmarshal(raw, &cur)
	for k, v := range errorFields {
		cur[k] = v
	}
	value, _ := json.Marshal(map[string]any{"errors": cur})
	settings.Invalidate()
	apierr.Invalidate()
	return sql.OpenDB(connector{value: value})
}

type connector struct{ value []byte }

func (c connector) Connect(context.Context) (driver.Conn, error) { return conn(c), nil }
func (c connector) Driver() driver.Driver                        { return drv{} }

type drv struct{}

func (drv) Open(string) (driver.Conn, error) { return nil, errors.New("apierrtest: use DB") }

type conn struct{ value []byte }

func (c conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("apierrtest: no prepare") }
func (c conn) Close() error                        { return nil }
func (c conn) Begin() (driver.Tx, error)           { return nil, errors.New("apierrtest: no tx") }

func (c conn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "FROM devkit_settings"):
		return &rows{cols: []string{"value", "updated_by", "updated_at"}, data: [][]driver.Value{{c.value, "", time.Unix(0, 0)}}}, nil
	case strings.Contains(query, "FROM devkit_error_message"):
		return &rows{cols: []string{"code", "messages", "status", "retry_after_seconds", "is_custom", "updated_by", "updated_at"}}, nil
	}
	return nil, errors.New("apierrtest: table not stubbed")
}

type rows struct {
	cols []string
	data [][]driver.Value
}

func (r *rows) Columns() []string { return r.cols }
func (r *rows) Close() error      { return nil }
func (r *rows) Next(dest []driver.Value) error {
	if len(r.data) == 0 {
		return io.EOF
	}
	copy(dest, r.data[0])
	r.data = r.data[1:]
	return nil
}
