package scanner

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"hostCollision/internal/config"
	"hostCollision/internal/model"
)

func serverPort(t *testing.T, address string) int {
	t.Helper()
	u, err := url.Parse(address)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func TestScanHTTPAndHTTPSWithCandidateSNI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil && r.TLS.ServerName != r.Host {
			http.Error(w, "SNI does not match Host", http.StatusBadRequest)
			return
		}
		if r.Host == baselineHost {
			fmt.Fprint(w, strings.Repeat("default", 100))
			return
		}
		fmt.Fprintf(w, "<title>%s</title>%s", r.Host, strings.Repeat(r.Host, 100))
	})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()
	tlsServer := httptest.NewTLSServer(handler)
	defer tlsServer.Close()
	httpPort, tlsPort := serverPort(t, httpServer.URL), serverPort(t, tlsServer.URL)
	cfg := config.Config{Ports: []int{httpPort, tlsPort}, HTTPSPorts: []int{tlsPort}, Insecure: true, Goroutines: 1, Similarity: 85, MaxHostsPerIP: 10}
	results, err := Scan(context.Background(), []model.IP{"127.0.0.1"}, []model.Host{"first.example", "second.example"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 4 {
		t.Fatalf("want both hosts on both ports, got %#v", results)
	}
	seen := map[string]bool{}
	for _, r := range results {
		wantScheme := "http"
		if r.Port == tlsPort {
			wantScheme = "https"
		}
		key := fmt.Sprintf("%s:%d", r.Host, r.Port)
		if r.Title != string(r.Host) || r.Scheme != wantScheme || seen[key] {
			t.Fatalf("unexpected result: %#v", r)
		}
		seen[key] = true
	}
}

func TestHTTPSCertificateVerificationAndCancellation(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "TLS response") }))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	port := serverPort(t, server.URL)
	s := NewScanner(config.Config{HTTPSPorts: []int{port}})
	target := task{IP: "127.0.0.1", Port: port, Host: "example.com"}
	if _, err := s.fetchResponse(context.Background(), target); err == nil {
		t.Fatal("untrusted candidate certificate accepted")
	}
	// Trust the test certificate and verify the candidate name, not the IP.
	transport := s.httpClient.Transport.(*probeTransport)
	transport.http.TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	if _, err := s.fetchResponse(context.Background(), target); err != nil {
		t.Fatalf("trusted candidate failed: %v", err)
	}
	target.Host = "wrong.example"
	if _, err := s.fetchResponse(context.Background(), target); err == nil {
		t.Fatal("mismatched candidate certificate accepted")
	} else {
		var hostnameError x509.HostnameError
		if !errors.As(err, &hostnameError) {
			t.Fatalf("expected hostname verification error, got %v", err)
		}
	}
	// A synthetic baseline is allowed despite its deliberately invalid name.
	target.Host = model.Host(baselineHost)
	if _, err := s.fetchResponse(context.Background(), target); err != nil {
		t.Fatalf("baseline failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.fetchResponse(ctx, target); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
