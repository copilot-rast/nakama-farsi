package handoffkit

import (
	"html"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	KeyPlaceholder   = "{{SERVER_KEY}}"
	HostPlaceholder  = "{{SERVER_HOST}}"
	PortPlaceholder  = "{{SERVER_PORT}}"
	SSLPlaceholder   = "{{SERVER_SSL}}"
	URLPlaceholder   = "{{SERVER_URL}}"
	DefaultServerKey = "defaultkey"

	EnvPublicHost = "HANDOFF_PUBLIC_HOST"
	EnvPublicPort = "HANDOFF_PUBLIC_PORT"
	EnvPublicSSL  = "HANDOFF_PUBLIC_SSL"
	EnvPublicBase = "ECONOMY_PUBLIC_BASE_URL"
)

var Placeholders = []string{KeyPlaceholder, HostPlaceholder, PortPlaceholder, SSLPlaceholder, URLPlaceholder}

type Connection struct {
	ServerKey  string `json:"serverKey,omitempty"`
	KeyShown   bool   `json:"keyShown"`
	DefaultKey bool   `json:"defaultKey"`
	Host       string `json:"host,omitempty"`
	Port       int    `json:"port,omitempty"`
	SSL        bool   `json:"ssl"`
	HostSet    bool   `json:"hostSet"`
	URL        string `json:"url,omitempty"`
}

type Server struct {
	ServerKey string
	Env       map[string]string
}

var hostRe = regexp.MustCompile(`^[A-Za-z0-9.\-]+$|^\[[0-9A-Fa-f:.]+\]$`)

func (s Server) env(name string) string {
	if v := strings.TrimSpace(s.Env[name]); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(name))
}

func Resolve(s Server, set Settings) Connection {
	c := Connection{DefaultKey: s.ServerKey == DefaultServerKey}
	if set.ShowServerKey && s.ServerKey != "" {
		c.ServerKey, c.KeyShown = s.ServerKey, true
	}
	host, port, ssl, sslSet := "", 0, false, false
	if raw := s.env(EnvPublicHost); raw != "" {
		host, port, ssl, sslSet = parseHost(raw)
	} else if raw := s.env(EnvPublicBase); raw != "" && strings.Contains(raw, "://") {
		host, port, ssl, sslSet = parseHost(raw)
		if strings.HasSuffix(host, "example.com") {
			host = ""
		}
	}
	if host == "" {
		return c
	}
	if v, err := strconv.ParseBool(s.env(EnvPublicSSL)); err == nil {
		ssl, sslSet = v, true
	}
	if !sslSet {
		ssl = true
	}
	if p, err := strconv.Atoi(s.env(EnvPublicPort)); err == nil && p > 0 && p < 65536 {
		port = p
	}
	if port == 0 {
		port = 80
		if ssl {
			port = 443
		}
	}
	c.Host, c.Port, c.SSL, c.HostSet = host, port, ssl, true
	scheme := "http"
	if ssl {
		scheme = "https"
	}
	c.URL = scheme + "://" + host
	if (ssl && port != 443) || (!ssl && port != 80) {
		c.URL += ":" + strconv.Itoa(port)
	}
	return c
}

func parseHost(raw string) (string, int, bool, bool) {
	ssl, sslSet := false, false
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", 0, false, false
		}
		switch u.Scheme {
		case "https", "wss":
			ssl, sslSet = true, true
		case "http", "ws":
			ssl, sslSet = false, true
		}
		raw = u.Host
	}
	raw = strings.TrimRight(raw, "/")
	host, port := raw, 0
	if h, p, err := net.SplitHostPort(raw); err == nil {
		host = h
		if strings.Contains(h, ":") {
			host = "[" + h + "]"
		}
		port, _ = strconv.Atoi(p)
	}
	if !hostRe.MatchString(host) {
		return "", 0, false, false
	}
	return host, port, ssl, sslSet
}

const (
	codeAskKey  = "ASK_THE_SERVER_ADMIN"
	codeAskHost = "API_HOST_FROM_ADMIN"
	codeAskPort = "API_PORT_FROM_ADMIN"
	codeAskSSL  = "true"
	codeAskURL  = "https://API_HOST_FROM_ADMIN"
	textAskKey  = "از مدیر سرور بگیرید"
	textAskHost = "مدیر سرور تعیین می‌کند"
)

func (c Connection) values(inCode bool) map[string]string {
	key, host, port, ssl, u := c.ServerKey, c.Host, strconv.Itoa(c.Port), strconv.FormatBool(c.SSL), c.URL
	if !inCode {
		ssl = "خیر"
		if c.SSL {
			ssl = "بله"
		}
	}
	if !c.KeyShown {
		key = codeAskKey
		if !inCode {
			key = textAskKey
		}
	}
	if !c.HostSet {
		host, port, ssl, u = codeAskHost, codeAskPort, codeAskSSL, codeAskURL
		if !inCode {
			host, port, ssl, u = textAskHost, textAskHost, textAskHost, textAskHost
		}
	}
	return map[string]string{KeyPlaceholder: key, HostPlaceholder: host, PortPlaceholder: port, SSLPlaceholder: ssl, URLPlaceholder: u}
}

var codeSpanRe = regexp.MustCompile(`(?s)<code[^>]*>.*?</code>`)

func (c Connection) Apply(body string) string {
	if !strings.Contains(body, "{{SERVER_") {
		return body
	}
	code := c.values(true)
	text := c.values(false)
	swap := func(s string, vals map[string]string) string {
		for _, p := range Placeholders {
			s = strings.ReplaceAll(s, p, html.EscapeString(vals[p]))
		}
		return s
	}
	var b strings.Builder
	last := 0
	for _, m := range codeSpanRe.FindAllStringIndex(body, -1) {
		b.WriteString(swap(body[last:m[0]], text))
		b.WriteString(swap(body[m[0]:m[1]], code))
		last = m[1]
	}
	b.WriteString(swap(body[last:], text))
	return b.String()
}

func connCode(v string) string {
	return `<div class="codeblock"><div class="code-head"><span class="lang">Text</span><button class="copy" type="button">Copy</button></div><pre class="oneline"><code>` + html.EscapeString(v) + `</code></pre></div>`
}

func connectionBlock(c Connection) string {
	esc := html.EscapeString
	var b strings.Builder
	b.WriteString(`<section class="step" id="connection" data-connection data-key-shown="` + strconv.FormatBool(c.KeyShown) + `" data-host-set="` + strconv.FormatBool(c.HostSet) + `"><div class="step-head"><div class="num">◆</div><div><h2>اطلاعات اتصال</h2><div class="sub">همان مقدارهایی که اپ برای وصل شدن به سرور بازی لازم دارد</div></div></div>` + "\n")
	b.WriteString(`<div class="tablewrap"><table><thead><tr><th>چه چیزی</th><th>مقدار</th></tr></thead><tbody>` + "\n")
	if c.HostSet {
		b.WriteString(`<tr><td>آدرس سرور بازی</td><td>` + connCode(c.URL) + "</td></tr>\n")
		b.WriteString(`<tr><td>میزبان (host)</td><td><code>` + esc(c.Host) + "</code></td></tr>\n")
		b.WriteString(`<tr><td>پورت</td><td><code>` + strconv.Itoa(c.Port) + "</code></td></tr>\n")
		b.WriteString(`<tr><td>SSL</td><td><code>` + strconv.FormatBool(c.SSL) + "</code></td></tr>\n")
	} else {
		b.WriteString(`<tr><td>آدرس سرور بازی (میزبان، پورت، SSL)</td><td data-host-missing>مدیر سرور هنوز آدرس عمومی سرور را تنظیم نکرده است. آن را از مدیر سرور بگیرید.</td></tr>` + "\n")
	}
	if c.KeyShown {
		b.WriteString(`<tr><td>کلید سرور (server key)</td><td data-server-key>` + connCode(c.ServerKey) + "</td></tr>\n")
	} else {
		b.WriteString(`<tr><td>کلید سرور (server key)</td><td data-key-hidden>از مدیر سرور بگیرید.</td></tr>` + "\n")
	}
	b.WriteString("</tbody></table></div>\n")
	if c.KeyShown && c.DefaultKey {
		b.WriteString(`<div class="warn" data-default-key>این سرور هنوز کلید پیش‌فرض <code>defaultkey</code> را دارد. برای آزمایش مشکلی نیست، ولی سرور اصلی (production) باید کلید دیگری داشته باشد. بعد از عوض شدن کلید، این فایل را دوباره بگیرید و بیلد اپ را تازه کنید.</div>` + "\n")
	}
	if !c.HostSet {
		b.WriteString(`<div class="note">مدیر سرور: آدرس عمومی را با <code>HANDOFF_PUBLIC_HOST</code>، <code>HANDOFF_PUBLIC_PORT</code> و <code>HANDOFF_PUBLIC_SSL</code> تنظیم کنید تا اینجا خودکار پر شود.</div>` + "\n")
	}
	b.WriteString(`<p>کلید سرور داخل بیلد اپ هم هست، پس رمز به حساب نمی‌آید. کلید HTTP، رمز کنسول و کلید کنسول هیچ‌وقت در اپ نیستند.</p>` + "\n")
	b.WriteString("</section>\n")
	return b.String()
}
