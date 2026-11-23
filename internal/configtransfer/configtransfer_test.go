package configtransfer

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSecretName(t *testing.T) {
	for _, n := range []string{"apiKey", "api_key", "merchantId", "webhookSecret", "password", "accessToken", "privateKeyP8", "serverKey", "http_key", "serviceAccountJson", "SMS_KEY"} {
		if !SecretName(n) {
			t.Errorf("%s must be a secret", n)
		}
	}
	for _, n := range []string{"key", "currencyKey", "prizeKey", "packageName", "rsaPublicKey", "keyId", "name", "metadata", "sandbox"} {
		if SecretName(n) {
			t.Errorf("%s must not be a secret", n)
		}
	}
}

func TestScrubAndRestore(t *testing.T) {
	v, _ := decodeJSON([]byte(`{"mode":"x","apiKey":"S1","showServerKey":true,"list":[{"token":"S2","n":1}],"deep":{"password":"S3"}}`))
	out := Scrub(deepCopy(v))
	raw, _ := json.Marshal(out)
	for _, s := range []string{"S1", "S2", "S3"} {
		if strings.Contains(string(raw), s) {
			t.Fatalf("%s not scrubbed: %s", s, raw)
		}
	}
	if !strings.Contains(string(raw), `"showServerKey":true`) {
		t.Fatal("booleans are not secrets")
	}
	next, _ := decodeJSON([]byte(`{"mode":"y","list":[{"n":2}],"deep":{}}`))
	back, _ := json.Marshal(RestoreSecrets(next, v))
	for _, s := range []string{`"apiKey":"S1"`, `"token":"S2"`, `"password":"S3"`, `"mode":"y"`} {
		if !strings.Contains(string(back), s) {
			t.Fatalf("restore lost %s: %s", s, back)
		}
	}
}

func TestMapPath(t *testing.T) {
	v, _ := decodeJSON([]byte(`[{"requirements":{"itemIds":["a","b",""],"metas":[{"metaId":"m"}]}},{"requirements":null}]`))
	ids := map[string]string{"a": "A", "m": "M"}
	fn := func(x any, inArray bool) (any, bool) {
		s, _ := x.(string)
		if k, ok := ids[s]; ok {
			return k, true
		}
		if inArray {
			return nil, false
		}
		return missingPrefix + s, true
	}
	for _, p := range []string{"[].requirements.itemIds[]", "[].requirements.metas[].metaId"} {
		segs, err := parsePath(p)
		if err != nil {
			t.Fatal(err)
		}
		v = mapPath(v, segs, fn)
	}
	raw, _ := json.Marshal(v)
	if string(raw) != `[{"requirements":{"itemIds":["A",""],"metas":[{"metaId":"M"}]}},{"requirements":null}]` {
		t.Fatalf("got %s", raw)
	}
	if _, err := parsePath("a..b"); err == nil {
		t.Fatal("bad path accepted")
	}
}

func TestDecode(t *testing.T) {
	if _, err := Decode([]byte(`{"format":"other","version":1}`)); err != ErrFormat {
		t.Fatalf("got %v", err)
	}
	if _, err := Decode([]byte(`{"format":"nakama-config","version":99}`)); err != ErrVersion {
		t.Fatalf("got %v", err)
	}
	f, err := Decode([]byte(`{"format":"nakama-config","version":1,"areas":[{"key":"items","tables":[{"table":"economy_item","rows":[{"key":"a","values":{"n":1.50}}]}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if n := f.Area("items").Table("economy_item").Rows[0].Values["n"]; canon(n) != "1.50" {
		t.Fatalf("numbers must keep their text: %v", n)
	}
}

func TestRegisterValidates(t *testing.T) {
	if err := Register(Area{Key: "Bad Key", Tables: []Table{{Name: "x", Key: []string{"key"}}}}); err == nil {
		t.Fatal("bad key accepted")
	}
	if err := Register(Area{Key: "ok_area", Tables: []Table{{Name: "x;drop", Key: []string{"key"}}}}); err == nil {
		t.Fatal("bad table accepted")
	}
	all := append(append(BaseAreas(), EconomyAreas()...), LeaderboardAreas()...)
	if err := Register(all...); err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, a := range all {
			Unregister(a.Key)
		}
	}()
	list := List()
	if len(list) != len(all) || list[0].Key != "taxonomy" {
		t.Fatalf("bad order %+v", list[0])
	}
}
