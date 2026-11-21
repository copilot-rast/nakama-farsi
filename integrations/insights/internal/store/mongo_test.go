package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	mw "github.com/heroiclabs/nakama/v3/integrations/insights/internal/mongowire"
)

type memMongo struct {
	mu   sync.Mutex
	docs map[string]mw.D
}

func lookup(doc mw.D, path string) (any, bool) {
	var cur any = doc
	for _, p := range strings.Split(path, ".") {
		d, ok := cur.(mw.D)
		if !ok {
			return nil, false
		}
		found := false
		for _, e := range d {
			if e.Key == p {
				cur, found = e.Val, true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return cur, true
}

func cmpAny(a, b any) int {
	if fa, ok := mw.Num(a); ok {
		if fb, ok := mw.Num(b); ok {
			switch {
			case fa < fb:
				return -1
			case fa > fb:
				return 1
			}
			return 0
		}
	}
	if ta, ok := a.(time.Time); ok {
		if tb, ok := b.(time.Time); ok {
			return ta.Compare(tb)
		}
	}
	return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
}

func matches(doc mw.D, q mw.D) bool {
	for _, c := range q {
		if c.Key == "$or" {
			any := false
			for _, alt := range c.Val.(mw.A) {
				if matches(doc, alt.(mw.D)) {
					any = true
				}
			}
			if !any {
				return false
			}
			continue
		}
		v, has := lookup(doc, c.Key)
		ops, isOps := c.Val.(mw.D)
		if !isOps || len(ops) == 0 || !strings.HasPrefix(ops[0].Key, "$") {
			if !has || cmpAny(v, c.Val) != 0 {
				return false
			}
			continue
		}
		for _, op := range ops {
			switch op.Key {
			case "$in":
				in := false
				for _, x := range op.Val.(mw.A) {
					if has && cmpAny(v, x) == 0 {
						in = true
					}
				}
				if !in {
					return false
				}
			case "$ne":
				if has && cmpAny(v, op.Val) == 0 {
					return false
				}
			case "$exists":
				if has != op.Val.(bool) {
					return false
				}
			case "$gte":
				if !has || cmpAny(v, op.Val) < 0 {
					return false
				}
			case "$lte":
				if !has || cmpAny(v, op.Val) > 0 {
					return false
				}
			case "$lt":
				if !has || cmpAny(v, op.Val) >= 0 {
					return false
				}
			case "$gt":
				if !has || cmpAny(v, op.Val) <= 0 {
					return false
				}
			default:
				panic("op " + op.Key)
			}
		}
	}
	return true
}

func eval(doc mw.D, expr any) any {
	switch x := expr.(type) {
	case string:
		if strings.HasPrefix(x, "$") {
			v, _ := lookup(doc, x[1:])
			return v
		}
		return x
	case mw.D:
		if len(x) == 1 && strings.HasPrefix(x[0].Key, "$") {
			switch x[0].Key {
			case "$ifNull":
				args := x[0].Val.(mw.A)
				if v := eval(doc, args[0]); v != nil {
					return v
				}
				return eval(doc, args[1])
			case "$size":
				return int32(len(eval(doc, x[0].Val).(mw.A)))
			case "$objectToArray":
				d, _ := eval(doc, x[0].Val).(mw.D)
				out := mw.A{}
				for _, e := range d {
					out = append(out, mw.D{{Key: "k", Val: e.Key}, {Key: "v", Val: e.Val}})
				}
				return out
			case "$dateToString":
				a := x[0].Val.(mw.D)
				t, _ := eval(doc, a.Get("date")).(time.Time)
				loc, _ := time.LoadLocation(a.Get("timezone").(string))
				return t.In(loc).Format("2006-01-02T15:00")
			}
			panic("expr " + x[0].Key)
		}
		out := mw.D{}
		for _, e := range x {
			out = append(out, mw.E{Key: e.Key, Val: eval(doc, e.Val)})
		}
		return out
	}
	return expr
}

func keyOf(v any) string { return fmt.Sprintf("%#v", v) }

func (m *memMongo) pipeline(stages mw.A) []mw.D {
	m.mu.Lock()
	docs := make([]mw.D, 0, len(m.docs))
	for _, d := range m.docs {
		docs = append(docs, d)
	}
	m.mu.Unlock()
	for _, st := range stages {
		s := st.(mw.D)[0]
		switch s.Key {
		case "$match":
			out := []mw.D{}
			for _, d := range docs {
				if matches(d, s.Val.(mw.D)) {
					out = append(out, d)
				}
			}
			docs = out
		case "$group":
			spec := s.Val.(mw.D)
			groups := map[string]mw.D{}
			order := []string{}
			sets := map[string]map[string]map[string]bool{}
			counts := map[string]map[string]int{}
			for _, d := range docs {
				id := eval(d, spec.Get("_id"))
				k := keyOf(id)
				g, ok := groups[k]
				if !ok {
					g = mw.D{{Key: "_id", Val: id}}
					order = append(order, k)
					sets[k], counts[k] = map[string]map[string]bool{}, map[string]int{}
				}
				for _, acc := range spec[1:] {
					op := acc.Val.(mw.D)[0]
					v := eval(d, op.Val)
					cur := g.Get(acc.Key)
					var next any = cur
					switch op.Key {
					case "$sum":
						f, ok := mw.Num(v)
						if !ok {
							f = 0
						}
						c, _ := mw.Num(cur)
						next = c + f
					case "$avg":
						f, ok := mw.Num(v)
						if ok {
							c, _ := mw.Num(cur)
							n := counts[k][acc.Key]
							next = (c*float64(n) + f) / float64(n+1)
							counts[k][acc.Key] = n + 1
						} else if cur == nil {
							next = nil
						}
					case "$min", "$max":
						if v != nil && (cur == nil || (op.Key == "$min") == (cmpAny(v, cur) < 0)) {
							next = v
						}
					case "$addToSet":
						if sets[k][acc.Key] == nil {
							sets[k][acc.Key] = map[string]bool{}
						}
						arr, _ := cur.(mw.A)
						if !sets[k][acc.Key][keyOf(v)] {
							sets[k][acc.Key][keyOf(v)] = true
							arr = append(arr, v)
						}
						next = arr
					}
					found := false
					for i := range g {
						if g[i].Key == acc.Key {
							g[i].Val, found = next, true
						}
					}
					if !found {
						g = append(g, mw.E{Key: acc.Key, Val: next})
					}
				}
				groups[k] = g
			}
			out := []mw.D{}
			for _, k := range order {
				out = append(out, groups[k])
			}
			docs = out
		case "$project":
			out := []mw.D{}
			for _, d := range docs {
				p := mw.D{{Key: "_id", Val: d.Get("_id")}}
				for _, e := range s.Val.(mw.D) {
					if n, ok := e.Val.(int32); ok && n == 1 {
						v, _ := lookup(d, e.Key)
						p = append(p, mw.E{Key: e.Key, Val: v})
						continue
					}
					p = append(p, mw.E{Key: e.Key, Val: eval(d, e.Val)})
				}
				out = append(out, p)
			}
			docs = out
		case "$unwind":
			out := []mw.D{}
			path := strings.TrimPrefix(s.Val.(string), "$")
			for _, d := range docs {
				arr, _ := lookup(d, path)
				for _, item := range arr.(mw.A) {
					nd := mw.D{}
					for _, e := range d {
						if e.Key == path {
							nd = append(nd, mw.E{Key: path, Val: item})
						} else {
							nd = append(nd, e)
						}
					}
					out = append(out, nd)
				}
			}
			docs = out
		case "$sort":
			spec := s.Val.(mw.D)
			sort.SliceStable(docs, func(i, j int) bool {
				for _, e := range spec {
					a, _ := lookup(docs[i], e.Key)
					b, _ := lookup(docs[j], e.Key)
					c := cmpAny(a, b)
					if e.Val.(int32) < 0 {
						c = -c
					}
					if c != 0 {
						return c < 0
					}
				}
				return false
			})
		case "$limit":
			if n := int(s.Val.(int32)); len(docs) > n {
				docs = docs[:n]
			}
		default:
			panic("stage " + s.Key)
		}
	}
	return docs
}

func (m *memMongo) handle(cmd mw.D) mw.D {
	switch cmd[0].Key {
	case "createIndexes", "ping":
		return mw.D{{Key: "ok", Val: 1.0}}
	case "insert":
		m.mu.Lock()
		defer m.mu.Unlock()
		errs := mw.A{}
		for i, d := range cmd.Get("documents").(mw.A) {
			doc := d.(mw.D)
			id := doc.Get("_id").(string)
			if _, dup := m.docs[id]; dup {
				errs = append(errs, mw.D{{Key: "index", Val: int32(i)}, {Key: "code", Val: int32(11000)}, {Key: "errmsg", Val: "E11000"}})
				continue
			}
			m.docs[id] = doc
		}
		out := mw.D{{Key: "ok", Val: 1.0}}
		if len(errs) > 0 {
			out = append(out, mw.E{Key: "writeErrors", Val: errs})
		}
		return out
	case "aggregate":
		rows := m.pipeline(cmd.Get("pipeline").(mw.A))
		batch := mw.A{}
		for _, r := range rows {
			batch = append(batch, r)
		}
		return mw.D{{Key: "cursor", Val: mw.D{{Key: "id", Val: int64(0)}, {Key: "firstBatch", Val: batch}}}, {Key: "ok", Val: 1.0}}
	case "delete":
		m.mu.Lock()
		defer m.mu.Unlock()
		q := cmd.Get("deletes").(mw.A)[0].(mw.D).Get("q").(mw.D)
		n := 0
		for id, d := range m.docs {
			if matches(d, q) {
				delete(m.docs, id)
				n++
			}
		}
		return mw.D{{Key: "n", Val: int32(n)}, {Key: "ok", Val: 1.0}}
	}
	return mw.D{{Key: "ok", Val: 0.0}, {Key: "code", Val: int32(59)}, {Key: "errmsg", Val: "no such command " + cmd[0].Key}}
}

func TestMongoStoreContractAgainstAnInMemoryServer(t *testing.T) {
	mem := &memMongo{docs: map[string]mw.D{}}
	srv, err := mw.NewFakeServer("ins", "pw", mem.handle)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cl, _ := mw.New(mw.Options{Hosts: []string{srv.Addr()}, User: "ins", Password: "pw", Timeout: 5 * time.Second})
	st, err := NewMongo(cl, "analytics", "")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := st.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	storeContract(t, st)
	if _, err := NewMongo(cl, "bad-db", ""); err != ErrBadQuery {
		t.Fatal("bad db accepted")
	}
}
