package scanner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"hostCollision/internal/config"
	"hostCollision/internal/model"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestScanUsesSeparateBaselineAndEnforcesLimit(t *testing.T) {
	cfg := config.Config{Goroutines: 8, Similarity: 85, MaxHostsPerIP: 2}
	s := NewScanner(cfg)
	s.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Host == baselineHost || req.Host == "default.example" {
			return response(http.StatusOK, "default response"), nil
		}
		return response(http.StatusOK, fmt.Sprintf("unique response for %s", req.Host)), nil
	})

	hosts := []model.Host{"one.example", "two.example", "three.example", "default.example"}
	results, err := s.Scan(context.Background(), []model.IP{"192.0.2.1"}, hosts)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(results) != cfg.MaxHostsPerIP {
		t.Fatalf("expected exactly %d results, got %d", cfg.MaxHostsPerIP, len(results))
	}
	for _, result := range results {
		if result.Host == "default.example" {
			t.Fatal("default response should have been filtered by the dedicated baseline")
		}
	}
}

func TestClientDoesNotFollowRedirects(t *testing.T) {
	var requests atomic.Int32
	s := NewScanner(config.Config{})
	s.httpClient.Transport = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		requests.Add(1)
		resp := response(http.StatusFound, "redirect")
		resp.Header.Set("Location", "http://unrelated.example/")
		return resp, nil
	})

	_, status, err := s.fetchResponse(context.Background(), task{IP: "192.0.2.1", Host: "example.com"})
	if err != nil {
		t.Fatalf("fetchResponse returned error: %v", err)
	}
	if status != http.StatusFound {
		t.Fatalf("expected redirect status, got %d", status)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("expected one request, got %d", got)
	}
}
