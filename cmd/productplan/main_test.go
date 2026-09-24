package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

func TestHealthCheckReportsCacheStats(t *testing.T) {
	client, err := api.New(api.Config{Token: "t", BaseURL: "http://127.0.0.1:0", CacheTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	report := newHealthChecker(client, "test").Check(context.Background(), false)
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Cache struct {
			Enabled    bool    `json:"enabled"`
			TTLSeconds float64 `json:"ttl_seconds"`
			Hits       *uint64 `json:"hits"`
			Misses     *uint64 `json:"misses"`
		} `json:"cache"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if !parsed.Cache.Enabled || parsed.Cache.TTLSeconds != 60 || parsed.Cache.Hits == nil || parsed.Cache.Misses == nil {
		t.Errorf("health report cache block = %s", raw)
	}
}
