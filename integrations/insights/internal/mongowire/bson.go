package mongowire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type E struct {
	Key string
	Val any
}

type D []E

type A []any

type Binary struct {
	Subtype byte
	Data    []byte
}

type ObjectID [12]byte

func (d D) Get(key string) any {
	for _, e := range d {
		if e.Key == key {
			return e.Val
		}
	}
	return nil
}

func (d D) Map() map[string]any {
	m := make(map[string]any, len(d))
	for _, e := range d {
		m[e.Key] = e.Val
	}
	return m
}

var ErrBSON = errors.New("mongowire: bad bson")

func Marshal(doc D) ([]byte, error) {
	var b bytes.Buffer
	if err := writeDoc(&b, doc); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func writeDoc(b *bytes.Buffer, doc D) error {
	start := b.Len()
	b.Write([]byte{0, 0, 0, 0})
	for _, e := range doc {
		if err := writeElem(b, e.Key, e.Val); err != nil {
			return err
		}
	}
	b.WriteByte(0)
	binary.LittleEndian.PutUint32(b.Bytes()[start:], uint32(b.Len()-start))
	return nil
}

func cstring(b *bytes.Buffer, s string) error {
	if bytes.IndexByte([]byte(s), 0) >= 0 {
		return ErrBSON
	}
	b.WriteString(s)
	b.WriteByte(0)
	return nil
}

func writeElem(b *bytes.Buffer, key string, v any) error {
	head := func(t byte) error {
		b.WriteByte(t)
		return cstring(b, key)
	}
	var tmp [8]byte
	switch x := v.(type) {
	case nil:
		return head(0x0A)
	case float64:
		if err := head(0x01); err != nil {
			return err
		}
		binary.LittleEndian.PutUint64(tmp[:], math.Float64bits(x))
		b.Write(tmp[:])
	case *float64:
		if x == nil {
			return head(0x0A)
		}
		return writeElem(b, key, *x)
	case string:
		if err := head(0x02); err != nil {
			return err
		}
		binary.LittleEndian.PutUint32(tmp[:4], uint32(len(x)+1))
		b.Write(tmp[:4])
		b.WriteString(x)
		b.WriteByte(0)
	case D:
		if err := head(0x03); err != nil {
			return err
		}
		return writeDoc(b, x)
	case map[string]string:
		d := make(D, 0, len(x))
		for _, k := range sortedMapKeys(x) {
			d = append(d, E{k, x[k]})
		}
		return writeElem(b, key, d)
	case map[string]any:
		d := make(D, 0, len(x))
		for _, k := range sortedAnyKeys(x) {
			d = append(d, E{k, x[k]})
		}
		return writeElem(b, key, d)
	case A:
		if err := head(0x04); err != nil {
			return err
		}
		d := make(D, len(x))
		for i, item := range x {
			d[i] = E{strconv.Itoa(i), item}
		}
		return writeDoc(b, d)
	case []any:
		return writeElem(b, key, A(x))
	case []string:
		a := make(A, len(x))
		for i, s := range x {
			a[i] = s
		}
		return writeElem(b, key, a)
	case []D:
		a := make(A, len(x))
		for i, s := range x {
			a[i] = s
		}
		return writeElem(b, key, a)
	case Binary:
		if err := head(0x05); err != nil {
			return err
		}
		binary.LittleEndian.PutUint32(tmp[:4], uint32(len(x.Data)))
		b.Write(tmp[:4])
		b.WriteByte(x.Subtype)
		b.Write(x.Data)
	case ObjectID:
		if err := head(0x07); err != nil {
			return err
		}
		b.Write(x[:])
	case bool:
		if err := head(0x08); err != nil {
			return err
		}
		if x {
			b.WriteByte(1)
		} else {
			b.WriteByte(0)
		}
	case time.Time:
		if err := head(0x09); err != nil {
			return err
		}
		binary.LittleEndian.PutUint64(tmp[:], uint64(x.UnixMilli()))
		b.Write(tmp[:])
	case int32:
		if err := head(0x10); err != nil {
			return err
		}
		binary.LittleEndian.PutUint32(tmp[:4], uint32(x))
		b.Write(tmp[:4])
	case int:
		if x >= math.MinInt32 && x <= math.MaxInt32 {
			return writeElem(b, key, int32(x))
		}
		return writeElem(b, key, int64(x))
	case int64:
		if err := head(0x12); err != nil {
			return err
		}
		binary.LittleEndian.PutUint64(tmp[:], uint64(x))
		b.Write(tmp[:])
	default:
		return fmt.Errorf("mongowire: cannot encode %T", v)
	}
	return nil
}

func sortedMapKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sortStrings(out)
	return out
}

func sortedAnyKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sortStrings(out)
	return out
}

func sortStrings(v []string) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j-1] > v[j]; j-- {
			v[j-1], v[j] = v[j], v[j-1]
		}
	}
}

func Unmarshal(raw []byte) (D, error) {
	d, n, err := readDoc(raw)
	if err != nil {
		return nil, err
	}
	if n != len(raw) {
		return nil, ErrBSON
	}
	return d, nil
}

func readDoc(raw []byte) (D, int, error) {
	if len(raw) < 5 {
		return nil, 0, ErrBSON
	}
	size := int(binary.LittleEndian.Uint32(raw))
	if size < 5 || size > len(raw) || raw[size-1] != 0 {
		return nil, 0, ErrBSON
	}
	body := raw[4 : size-1]
	out := D{}
	for len(body) > 0 {
		t := body[0]
		end := bytes.IndexByte(body[1:], 0)
		if end < 0 {
			return nil, 0, ErrBSON
		}
		key := string(body[1 : 1+end])
		body = body[2+end:]
		v, n, err := readValue(t, body)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, E{key, v})
		body = body[n:]
	}
	return out, size, nil
}

func readValue(t byte, b []byte) (any, int, error) {
	need := func(n int) error {
		if len(b) < n {
			return ErrBSON
		}
		return nil
	}
	switch t {
	case 0x01:
		if err := need(8); err != nil {
			return nil, 0, err
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(b)), 8, nil
	case 0x02, 0x0D, 0x0E:
		if err := need(4); err != nil {
			return nil, 0, err
		}
		n := int(binary.LittleEndian.Uint32(b))
		if n < 1 || len(b) < 4+n {
			return nil, 0, ErrBSON
		}
		return string(b[4 : 4+n-1]), 4 + n, nil
	case 0x03:
		d, n, err := readDoc(b)
		return d, n, err
	case 0x04:
		d, n, err := readDoc(b)
		if err != nil {
			return nil, 0, err
		}
		a := make(A, len(d))
		for i, e := range d {
			a[i] = e.Val
		}
		return a, n, nil
	case 0x05:
		if err := need(5); err != nil {
			return nil, 0, err
		}
		n := int(binary.LittleEndian.Uint32(b))
		if len(b) < 5+n {
			return nil, 0, ErrBSON
		}
		return Binary{Subtype: b[4], Data: append([]byte{}, b[5:5+n]...)}, 5 + n, nil
	case 0x06, 0x0A, 0xFF, 0x7F:
		return nil, 0, nil
	case 0x07:
		if err := need(12); err != nil {
			return nil, 0, err
		}
		var id ObjectID
		copy(id[:], b[:12])
		return id, 12, nil
	case 0x08:
		if err := need(1); err != nil {
			return nil, 0, err
		}
		return b[0] == 1, 1, nil
	case 0x09:
		if err := need(8); err != nil {
			return nil, 0, err
		}
		return time.UnixMilli(int64(binary.LittleEndian.Uint64(b))).UTC(), 8, nil
	case 0x0B:
		i := bytes.IndexByte(b, 0)
		if i < 0 {
			return nil, 0, ErrBSON
		}
		j := bytes.IndexByte(b[i+1:], 0)
		if j < 0 {
			return nil, 0, ErrBSON
		}
		return string(b[:i]), i + j + 2, nil
	case 0x10:
		if err := need(4); err != nil {
			return nil, 0, err
		}
		return int32(binary.LittleEndian.Uint32(b)), 4, nil
	case 0x11, 0x12:
		if err := need(8); err != nil {
			return nil, 0, err
		}
		return int64(binary.LittleEndian.Uint64(b)), 8, nil
	case 0x13:
		if err := need(16); err != nil {
			return nil, 0, err
		}
		return Binary{Subtype: 0x13, Data: append([]byte{}, b[:16]...)}, 16, nil
	}
	return nil, 0, fmt.Errorf("mongowire: bson type 0x%x not supported", t)
}

func Num(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	}
	return 0, false
}
