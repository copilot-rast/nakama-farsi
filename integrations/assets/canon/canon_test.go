package canon

import "testing"

func TestCanonical(t *testing.T) {
	cases := []struct{ in, out string }{
		{`{ "b": 1.50, "a": [1, { "y": null, "x": "q" }] }`, `{"a":[1,{"x":"q","y":null}],"b":1.50}`},
		{`{"a":1,"a":2}`, `{"a":2}`},
		{`{"s":"tab\there é \u0001 \/ \"q\""}`, "{\"s\":\"tab\\there é \\u0001 / \\\"q\\\"\"}"},
		{`{"B":true,"a":false,"_":1e3}`, `{"B":true,"_":1e3,"a":false}`},
		{"{\"｡\":1,\"\U0001F600\":2}", "{\"\U0001F600\":2,\"｡\":1}"},
		{`{"emoji":"😀"}`, "{\"emoji\":\"\U0001F600\"}"},
	}
	for _, c := range cases {
		d, err := Parse([]byte(c.in))
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		if string(d.Canonical) != c.out {
			t.Fatalf("%s\n got %s\nwant %s", c.in, d.Canonical, c.out)
		}
	}
	d, _ := Parse([]byte(`{ "b": 1, "a": 2 }`))
	if string(d.Compact) != `{"b":1,"a":2}` {
		t.Fatalf("compact keeps order: %s", d.Compact)
	}
	if len(d.Hash) != 64 {
		t.Fatalf("hash %s", d.Hash)
	}
	h1, _ := Hash([]byte(`{"a":2,"b":1}`))
	if h1 != d.Hash {
		t.Fatal("same document, different key order, must hash the same")
	}
}

func TestRefused(t *testing.T) {
	for in, want := range map[string]error{
		`[1]`:             ErrNotObject,
		`{"a":`:           ErrInvalid,
		`{"a":1} {}`:      ErrInvalid,
		`{"a":"\u0000"}`:  ErrNul,
		`{"a":"\ud800"}`:  ErrSurrogate,
		`{"a":"\udc00x"}`: ErrSurrogate,
		`{"a":"\ud800A"}`: ErrSurrogate,
	} {
		if _, err := Parse([]byte(in)); err != want {
			t.Fatalf("%s: got %v want %v", in, err, want)
		}
	}
}
