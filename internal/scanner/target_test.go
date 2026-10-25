package scanner

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"hostCollision/internal/config"
	"hostCollision/internal/model"
)

func TestExpandTargets(t *testing.T) {
	for _, tt := range []struct {
		name string
		cfg  config.Config
		ips  []model.IP
		want []endpoint
	}{
		{"default", config.Config{}, []model.IP{"192.0.2.1"}, []endpoint{{"192.0.2.1", 80, "http"}}},
		{"ports and duplicates", config.Config{Ports: []int{80, 443, 80}}, []model.IP{"192.0.2.1", "192.0.2.1"}, []endpoint{{"192.0.2.1", 80, "http"}, {"192.0.2.1", 443, "https"}}},
		{"explicit port wins", config.Config{Ports: []int{80, 443}}, []model.IP{"192.0.2.1:8080", "[2001:db8::1]:8443"}, []endpoint{{"192.0.2.1", 8080, "http"}, {"2001:db8::1", 8443, "https"}}},
		{"IPv6 and custom TLS", config.Config{Ports: []int{9443}, HTTPSPorts: []int{9443}}, []model.IP{"2001:db8::1", "[2001:db8::2]"}, []endpoint{{"2001:db8::1", 9443, "https"}, {"2001:db8::2", 9443, "https"}}},
		{"HTTP only", config.Config{Ports: []int{443}, HTTPSPorts: []int{}}, []model.IP{"192.0.2.1"}, []endpoint{{"192.0.2.1", 443, "http"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScanner(tt.cfg).expandTargets(tt.ips)
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, %v; want %#v", got, err, tt.want)
			}
		})
	}
	for _, value := range []model.IP{"example.com", "192.0.2.1:0", "192.0.2.1:65536", "[2001:db8::1]:bad", "http://192.0.2.1/"} {
		if _, err := NewScanner(config.Config{}).expandTargets([]model.IP{value}); err == nil {
			t.Errorf("accepted invalid target %q", value)
		}
	}
	if _, err := NewScanner(config.Config{Ports: []int{0}}).expandTargets([]model.IP{"192.0.2.1"}); err == nil {
		t.Error("accepted invalid port")
	}
}

func TestScanSeparatesPortBaselinesAndLimits(t *testing.T) {
	s := NewScanner(config.Config{Ports: []int{8080, 8081}, Goroutines: 8, Similarity: 85, MaxHostsPerIP: 1})
	s.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := strings.Repeat("a", 256)
		if req.URL.Port() == "8081" {
			body = strings.Repeat("z", 256)
		}
		if req.Host != baselineHost && req.Host != "default.example" {
			if req.URL.Port() == "8080" {
				body = strings.Repeat("z", 256)
			} else {
				body = strings.Repeat("a", 256)
			}
		}
		return response(http.StatusOK, body), nil
	})
	results, err := s.Scan(context.Background(), []model.IP{"192.0.2.1"}, []model.Host{"default.example", "one.example", "two.example", "three.example"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("want one hit per port, got %#v", results)
	}
	seen := map[int]bool{}
	for _, r := range results {
		if r.Host == "default.example" || r.Similar != 0 || seen[r.Port] || r.IP != "192.0.2.1" || r.Scheme != "http" {
			t.Fatalf("incorrect per-port result: %#v", r)
		}
		seen[r.Port] = true
	}
}
