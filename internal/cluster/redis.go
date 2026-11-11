package cluster

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOptions struct {
	Addresses  string
	Username   string
	Password   string
	DB         int
	MasterName string
	TLS        bool
	PoolSize   int
}

func NewRedis(o RedisOptions) redis.UniversalClient {
	addrs := make([]string, 0, 1)
	for _, a := range strings.Split(o.Addresses, ",") {
		if a = strings.TrimSpace(a); a != "" {
			addrs = append(addrs, a)
		}
	}
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:6379"}
	}
	opts := &redis.UniversalOptions{
		Addrs:           addrs,
		Username:        o.Username,
		Password:        o.Password,
		DB:              o.DB,
		MasterName:      o.MasterName,
		PoolSize:        o.PoolSize,
		DialTimeout:     2 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		MaxRetries:      2,
		DisableIdentity: true,
	}
	if o.TLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return redis.NewUniversalClient(opts)
}

type Keys struct {
	Prefix string
}

func (k Keys) Nodes() string                { return k.Prefix + ":{reg}:nodes" }
func (k Keys) Heartbeat(name string) string { return k.Prefix + ":{reg}:hb:" + name }
func (k Keys) Epoch() string                { return k.Prefix + ":{reg}:epoch" }
func (k Keys) Reap(field string) string     { return k.Prefix + ":{reg}:reap:" + field }
func (k Keys) Dead(field string) string     { return k.Prefix + ":{reg}:dead:" + field }
func (k Keys) NodePresences(field string) string {
	return k.Prefix + ":np:" + field
}
func (k Keys) NodeFollows(field string) string {
	return k.Prefix + ":nf:" + field
}
func (k Keys) Invalidations(console bool) string {
	if console {
		return k.Prefix + ":invc"
	}
	return k.Prefix + ":inv"
}
func (k Keys) Signal(topic string) string { return k.Prefix + ":sig:" + topic }
func (k Keys) SignalPattern() string      { return k.Prefix + ":sig:*" }
