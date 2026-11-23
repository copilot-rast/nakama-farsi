package configtransfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Source struct {
	Name       string `json:"name"`
	Node       string `json:"node,omitempty"`
	ExportedAt string `json:"exportedAt"`
	By         string `json:"by,omitempty"`
}

type File struct {
	Format  string     `json:"format"`
	Version int        `json:"version"`
	Source  Source     `json:"source"`
	Counts  Counted    `json:"counts"`
	Areas   []FileArea `json:"areas"`
}

type Counted map[string]int

type FileArea struct {
	Key    string      `json:"key"`
	Owner  string      `json:"owner"`
	Count  int         `json:"count"`
	Tables []FileTable `json:"tables"`
}

type FileTable struct {
	Table string `json:"table"`
	Count int    `json:"count"`
	Rows  []Row  `json:"rows"`
}

type Row struct {
	Key    string         `json:"key"`
	Values map[string]any `json:"values"`
}

func Decode(raw []byte) (*File, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var f File
	if err := dec.Decode(&f); err != nil {
		return nil, ErrFormat
	}
	if f.Format != Format {
		return nil, ErrFormat
	}
	if f.Version < 1 || f.Version > FormatVersion {
		return nil, ErrVersion
	}
	for i := range f.Areas {
		a := &f.Areas[i]
		if !keyPattern.MatchString(a.Key) {
			return nil, fmt.Errorf("%w: bad area key", ErrFormat)
		}
		for j := range a.Tables {
			t := &a.Tables[j]
			if !identPattern.MatchString(t.Table) {
				return nil, fmt.Errorf("%w: bad table name", ErrFormat)
			}
			for k := range t.Rows {
				if t.Rows[k].Values == nil {
					t.Rows[k].Values = map[string]any{}
				}
			}
		}
	}
	return &f, nil
}

func (f *File) Area(key string) *FileArea {
	for i := range f.Areas {
		if f.Areas[i].Key == key {
			return &f.Areas[i]
		}
	}
	return nil
}

func (a *FileArea) Table(name string) *FileTable {
	for i := range a.Tables {
		if a.Tables[i].Table == name {
			return &a.Tables[i]
		}
	}
	return nil
}

func canon(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

func decodeJSON(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func keyLabel(parts []any) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, partLabel(p))
	}
	return strings.Join(out, "/")
}

func partLabel(p any) string {
	switch v := p.(type) {
	case nil:
		return "-"
	case string:
		return v
	case json.Number:
		return v.String()
	case []any:
		return keyLabel(v)
	default:
		return canon(v)
	}
}

func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = deepCopy(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = deepCopy(e)
		}
		return out
	default:
		return v
	}
}
