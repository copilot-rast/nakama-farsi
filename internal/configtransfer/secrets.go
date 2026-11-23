package configtransfer

import "strings"

var secretWords = []string{
	"secret", "password", "passwd", "passphrase", "token", "apikey", "privatekey", "signingkey", "serverkey", "httpkey",
	"accesskey", "authkey", "sessionkey", "encryptionkey", "hmackey", "licensekey", "masterkey", "merchant",
	"credential", "serviceaccount", "p8", "otpcode", "smskey", "pincode",
}

func SecretName(name string) bool {
	n := strings.ToLower(name)
	n = strings.NewReplacer("_", "", "-", "", ".", "", " ", "").Replace(n)
	if n == "" {
		return false
	}
	for _, w := range secretWords {
		if strings.Contains(n, w) {
			return true
		}
	}
	return false
}

func secretValue(v any) bool {
	_, isBool := v.(bool)
	return !isBool
}

func Scrub(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			if SecretName(k) && secretValue(e) {
				continue
			}
			out[k] = Scrub(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = Scrub(e)
		}
		return out
	default:
		return v
	}
}

func RestoreSecrets(next, prev any) any {
	switch n := next.(type) {
	case map[string]any:
		p, ok := prev.(map[string]any)
		if !ok {
			return next
		}
		for k, pv := range p {
			if SecretName(k) && secretValue(pv) {
				if _, given := n[k]; !given {
					n[k] = pv
				}
				continue
			}
			if nv, given := n[k]; given {
				n[k] = RestoreSecrets(nv, pv)
			}
		}
		return n
	case []any:
		p, ok := prev.([]any)
		if !ok || len(p) != len(n) {
			return next
		}
		for i := range n {
			n[i] = RestoreSecrets(n[i], p[i])
		}
		return n
	default:
		return next
	}
}
