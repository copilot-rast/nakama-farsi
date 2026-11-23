package configtransfer

import (
	"fmt"
	"strings"
)

type seg struct {
	field string
	each  bool
}

func parsePath(p string) ([]seg, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return nil, nil
	}
	var out []seg
	for _, part := range strings.Split(p, ".") {
		name := part
		n := 0
		for strings.HasSuffix(name, "[]") {
			name = strings.TrimSuffix(name, "[]")
			n++
		}
		if name == "" && n == 0 {
			return nil, fmt.Errorf("bad path %q", p)
		}
		if name != "" {
			if strings.ContainsAny(name, "[] ") {
				return nil, fmt.Errorf("bad path %q", p)
			}
			out = append(out, seg{field: name})
		}
		for i := 0; i < n; i++ {
			out = append(out, seg{each: true})
		}
	}
	return out, nil
}

type leafFn func(v any, inArray bool) (any, bool)

func mapPath(v any, path []seg, fn leafFn) any {
	if len(path) == 0 {
		return mapLeaf(v, fn)
	}
	s := path[0]
	if s.each {
		arr, ok := v.([]any)
		if !ok {
			return v
		}
		if len(path) == 1 {
			return mapLeaf(arr, fn)
		}
		for i := range arr {
			arr[i] = mapPath(arr[i], path[1:], fn)
		}
		return arr
	}
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	if e, has := m[s.field]; has {
		m[s.field] = mapPath(e, path[1:], fn)
	}
	return m
}

func mapLeaf(v any, fn leafFn) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		if x == "" {
			return x
		}
		out, _ := fn(x, false)
		return out
	case []any:
		out := make([]any, 0, len(x))
		for _, e := range x {
			s, ok := e.(string)
			if !ok || s == "" {
				out = append(out, e)
				continue
			}
			if nv, keep := fn(s, true); keep {
				out = append(out, nv)
			}
		}
		return out
	default:
		return v
	}
}
