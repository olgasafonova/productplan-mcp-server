package api

import (
	"net/http"
	"testing"
)

func TestClientUsesTunedTransport(t *testing.T) {
	c, err := NewSimple("tok")
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", c.httpClient.Transport)
	}
	if tr == http.DefaultTransport {
		t.Fatal("client must not share (and mutate) http.DefaultTransport")
	}
	if tr.MaxIdleConnsPerHost != transportMaxIdleConnsPerHost {
		t.Errorf("MaxIdleConnsPerHost = %d, want %d", tr.MaxIdleConnsPerHost, transportMaxIdleConnsPerHost)
	}
	if tr.MaxIdleConns != transportMaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", tr.MaxIdleConns, transportMaxIdleConns)
	}
	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 should be true")
	}
	if tr.IdleConnTimeout != transportIdleConnTimeout || tr.TLSHandshakeTimeout != transportTLSHandshakeTimeout {
		t.Errorf("timeouts = %v/%v", tr.IdleConnTimeout, tr.TLSHandshakeTimeout)
	}
	if tr.Proxy == nil {
		t.Error("clone should keep DefaultTransport's proxy func")
	}
	if c.httpClient.CheckRedirect == nil {
		t.Error("CheckRedirect must stay set (Article IX)")
	}
}
