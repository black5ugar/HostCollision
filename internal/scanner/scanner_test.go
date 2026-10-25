package scanner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"hostCollision/internal/config"
	"hostCollision/internal/model"
)

func TestScanIncludesTitleWithoutExtraRequests(t *testing.T) {
	var requests atomic.Int32
	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousOutput) })
	s := NewScanner(config.Config{Goroutines: 1, Similarity: 85, MaxHostsPerIP: 10})
	const body = "<html><title> 用户中心 &amp; Settings\n</title><body>Account page</body></html>"
	s.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests.Add(1)
		if req.Host == baselineHost {
			return response(http.StatusOK, "default response"), nil
		}
		return response(http.StatusOK, body), nil
	})

	results, err := s.Scan(context.Background(), []model.IP{"192.0.2.1"}, []model.Host{"account.example"})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %#v", results)
	}
	if got := results[0].Title; got != "用户中心 & Settings" {
		t.Fatalf("unexpected title: %q", got)
	}
	if results[0].Length != len(body) || results[0].BodyHash != bodyHash([]byte(body)) {
		t.Fatal("title extraction changed the original response fingerprint")
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("expected one baseline and one candidate request, got %d", got)
	}
	if !strings.Contains(logs.String(), `title="用户中心 & Settings"`) {
		t.Fatalf("success log is missing the title: %s", logs.String())
	}
}

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

	response, err := s.fetchResponse(context.Background(), task{IP: "192.0.2.1", Host: "example.com"})
	if err != nil {
		t.Fatalf("fetchResponse returned error: %v", err)
	}
	if response.status != http.StatusFound {
		t.Fatalf("expected redirect status, got %d", response.status)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("expected one request, got %d", got)
	}
}

func TestScanDistinguishesRedirectLocations(t *testing.T) {
	cfg := config.Config{Goroutines: 2, Similarity: 85, MaxHostsPerIP: 10}
	s := NewScanner(cfg)
	s.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		resp := response(http.StatusMovedPermanently, "")
		switch req.Host {
		case baselineHost, "default.example":
			resp.Header.Set("Location", "https://default.invalid/")
		default:
			resp.Header.Set("Location", "https://real.example/")
		}
		return resp, nil
	})

	results, err := s.Scan(context.Background(), []model.IP{"192.0.2.1"}, []model.Host{"default.example", "real.example"})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one distinct redirect, got %d results", len(results))
	}
	if results[0].Host != "real.example" || results[0].Similar != 0 {
		t.Fatalf("unexpected redirect result: %#v", results[0])
	}
}

func TestScannerCanBeReused(t *testing.T) {
	cfg := config.Config{Goroutines: 1, Similarity: 85, MaxHostsPerIP: 1}
	s := NewScanner(cfg)
	s.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Host == baselineHost {
			return response(http.StatusOK, "default response"), nil
		}
		return response(http.StatusOK, "unique response for "+req.Host), nil
	})

	for _, host := range []model.Host{"first.example", "second.example"} {
		results, err := s.Scan(context.Background(), []model.IP{"192.0.2.1"}, []model.Host{host})
		if err != nil {
			t.Fatalf("Scan for %s returned error: %v", host, err)
		}
		if len(results) != 1 || results[0].Host != host {
			t.Fatalf("expected a fresh result for %s, got %#v", host, results)
		}
	}
}

func TestFetchResponseFormatsBareIPv6Target(t *testing.T) {
	s := NewScanner(config.Config{})
	s.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "[2001:db8::1]:80" {
			t.Fatalf("unexpected URL host: %q", req.URL.Host)
		}
		return response(http.StatusOK, "ok"), nil
	})

	if _, err := s.fetchResponse(context.Background(), task{IP: "2001:db8::1", Host: "example.com"}); err != nil {
		t.Fatalf("fetchResponse returned error: %v", err)
	}
}
