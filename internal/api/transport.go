package api

import (
	"net/http"
	"time"
)

// Transport tuning. get_roadmap_complete fans out to four parallel GETs
// against a single host, and pagination can add more; the stdlib default of
// 2 idle connections per host forces most of those to redo the TCP and TLS
// handshake on every call.
const (
	transportMaxIdleConns        = 32
	transportMaxIdleConnsPerHost = 16
	transportIdleConnTimeout     = 90 * time.Second
	transportTLSHandshakeTimeout = 10 * time.Second
)

// newTransport clones http.DefaultTransport (keeping its proxy, dialer and
// HTTP/2 defaults) and raises the per-host idle pool so concurrent requests
// to the one ProductPlan host reuse warm connections. One transport is built
// per Client and shared by every request that client makes.
func newTransport() *http.Transport {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		// DefaultTransport was replaced process-wide with a foreign type;
		// build an equivalent rather than inherit something unknown.
		base = &http.Transport{Proxy: http.ProxyFromEnvironment}
	}
	t := base.Clone()
	t.ForceAttemptHTTP2 = true
	t.MaxIdleConns = transportMaxIdleConns
	t.MaxIdleConnsPerHost = transportMaxIdleConnsPerHost
	t.IdleConnTimeout = transportIdleConnTimeout
	t.TLSHandshakeTimeout = transportTLSHandshakeTimeout
	return t
}
