// Package canon builds the canonical form of a JSON document and its SHA-256
// hash, so an editor tool and the server agree on whether two payloads are the
// same: object keys sorted by UTF-16 code units (a repeated key keeps its last
// value), no whitespace, numbers exactly as written ("1.50" stays "1.50"),
// strings escaped only where JSON needs it (", \, \b \f \n \r \t, other
// characters below U+0020 as \u00xx in lower case; everything else raw).
package canon

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const MaxDepth = 64

var (
	ErrNotObject = errors.New("the document must be a JSON object")
	ErrInvalid   = errors.New("the document is not valid JSON")
	ErrTooDeep   = errors.New("the document is nested too deeply")
	ErrNul       = errors.New("strings must not hold the NUL character")
	ErrSurrogate = errors.New("strings must not hold a lone surrogate")
)

type member struct {
	key   string
	value any
}

// Doc is a parsed document: Compact keeps the sent key order, Canonical is sorted.
type Doc struct {
	Compact   []byte
	Canonical []byte
	Hash      string
}

// Parse checks raw (a JSON object) and returns its compact form, canonical form and hash.
func Parse(raw []byte) (Doc, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' {
		return Doc{}, ErrNotObject
	}
	if !utf8.Valid(raw) {
		return Doc{}, ErrInvalid
	}
	if err := checkEscapes(raw); err != nil {
		return Doc{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := readValue(dec, 0)
	if err != nil {
		return Doc{}, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return Doc{}, ErrInvalid
	}
	var c bytes.Buffer
	if err := json.Compact(&c, raw); err != nil {
		return Doc{}, ErrInvalid
	}
	var b bytes.Buffer
	write(&b, v)
	sum := sha256.Sum256(b.Bytes())
	return Doc{Compact: c.Bytes(), Canonical: b.Bytes(), Hash: hex.EncodeToString(sum[:])}, nil
}

// Hash returns the canonical hash of a JSON object.
func Hash(raw []byte) (string, error) {
	d, err := Parse(raw)
	return d.Hash, err
}

func readValue(dec *json.Decoder, depth int) (any, error) {
	if depth > MaxDepth {
		return nil, ErrTooDeep
	}
	tok, err := dec.Token()
	if err != nil {
		return nil, ErrInvalid
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			members := []member{}
			index := map[string]int{}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, ErrInvalid
				}
				key, ok := kt.(string)
				if !ok {
					return nil, ErrInvalid
				}
				if strings.ContainsRune(key, 0) {
					return nil, ErrNul
				}
				v, err := readValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				if i, seen := index[key]; seen {
					members[i].value = v
				} else {
					index[key] = len(members)
					members = append(members, member{key, v})
				}
			}
			if _, err := dec.Token(); err != nil {
				return nil, ErrInvalid
			}
			sort.SliceStable(members, func(i, j int) bool { return lessUTF16(members[i].key, members[j].key) })
			return members, nil
		case '[':
			list := []any{}
			for dec.More() {
				v, err := readValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				list = append(list, v)
			}
			if _, err := dec.Token(); err != nil {
				return nil, ErrInvalid
			}
			return list, nil
		}
		return nil, ErrInvalid
	case string:
		if strings.ContainsRune(t, 0) {
			return nil, ErrNul
		}
		return t, nil
	case json.Number, bool, nil:
		return t, nil
	}
	return nil, ErrInvalid
}

func lessUTF16(a, b string) bool {
	ea, eb := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(ea) && i < len(eb); i++ {
		if ea[i] != eb[i] {
			return ea[i] < eb[i]
		}
	}
	return len(ea) < len(eb)
}

const hexDigits = "0123456789abcdef"

func writeString(b *bytes.Buffer, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(`\u00`)
				b.WriteByte(hexDigits[r>>4])
				b.WriteByte(hexDigits[r&0xf])
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

func write(b *bytes.Buffer, v any) {
	switch t := v.(type) {
	case []member:
		b.WriteByte('{')
		for i, m := range t {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, m.key)
			b.WriteByte(':')
			write(b, m.value)
		}
		b.WriteByte('}')
	case []any:
		b.WriteByte('[')
		for i, x := range t {
			if i > 0 {
				b.WriteByte(',')
			}
			write(b, x)
		}
		b.WriteByte(']')
	case string:
		writeString(b, t)
	case json.Number:
		b.WriteString(string(t))
	case bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case nil:
		b.WriteString("null")
	}
}

// checkEscapes refuses \u escapes that leave a surrogate unpaired (Go would turn them into U+FFFD).
func checkEscapes(raw []byte) error {
	inString := false
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if !inString {
			if c == '"' {
				inString = true
			}
			continue
		}
		switch c {
		case '"':
			inString = false
		case '\\':
			if i+1 >= len(raw) {
				return ErrInvalid
			}
			if raw[i+1] != 'u' {
				i++
				continue
			}
			u, ok := hex4(raw, i+2)
			if !ok {
				return ErrInvalid
			}
			switch {
			case u >= 0xD800 && u <= 0xDBFF:
				if i+11 < len(raw)+0 && raw[i+6] == '\\' && raw[i+7] == 'u' {
					if lo, ok := hex4(raw, i+8); ok && lo >= 0xDC00 && lo <= 0xDFFF {
						i += 11
						continue
					}
				}
				return ErrSurrogate
			case u >= 0xDC00 && u <= 0xDFFF:
				return ErrSurrogate
			}
			i += 5
		}
	}
	return nil
}

func hex4(raw []byte, at int) (rune, bool) {
	if at+4 > len(raw) {
		return 0, false
	}
	var r rune
	for _, c := range raw[at : at+4] {
		r <<= 4
		switch {
		case c >= '0' && c <= '9':
			r |= rune(c - '0')
		case c >= 'a' && c <= 'f':
			r |= rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			r |= rune(c-'A') + 10
		default:
			return 0, false
		}
	}
	return r, true
}
