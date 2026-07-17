package scanner

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hostCollision/internal/model"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"
)

const maxResponseBody = 10 << 20 // 10 MiB

type probeResponse struct {
	body     []byte
	status   int
	location string
}

// worker consumes tasks from the task channel, executes the scan for each
// IP and host pair, and sends successful results to the results channel.
func (s *Scanner) worker(ctx context.Context, tasks <-chan task, results chan<- model.Result) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-tasks:
			if !ok {
				return
			}

			if s.reachedLimit(t.IP) {
				log.Printf("[SKIP-LIMIT] ip=%s host=%s reason=max hosts per IP reached (%d)",
					t.IP, t.Host, s.cfg.MaxHostsPerIP)
				continue
			}

			if s.cfg.Sleep > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(s.cfg.Sleep):
				}
			}

			res := s.processTarget(ctx, t)

			if res.Error != nil {
				log.Printf("[SKIP] ip=%s host=%s status=%d duration=%dms similar=%d error=%v",
					res.IP, res.Host, res.Status, res.Duration, res.Similar, res.Error)
				continue
			}

			if !s.claimSuccess(t.IP) {
				log.Printf("[SKIP-LIMIT] ip=%s host=%s reason=max hosts per IP reached (%d)",
					t.IP, t.Host, s.cfg.MaxHostsPerIP)
				continue
			}

			log.Printf("[OK]   ip=%s host=%s status=%d duration=%dms length=%d similar=%d",
				res.IP, res.Host, res.Status, res.Duration, res.Length, res.Similar)

			select {
			case <-ctx.Done():
				return
			case results <- res:
			}
		}
	}
}

// processTarget executes the HTTP probe against the given IP and host pair
// and returns a populated Result. A successful result is defined as a
// response with a valid status code and a similarity score below the
// configured threshold.
func (s *Scanner) processTarget(ctx context.Context, t task) model.Result {
	start := time.Now()
	response, err := s.fetchResponse(ctx, t)
	if err != nil {
		return model.Result{IP: t.IP, Host: t.Host, Error: err, Duration: time.Since(start).Milliseconds()}
	}

	result := model.Result{
		IP: t.IP, Host: t.Host, Status: response.status, Length: len(response.body),
		BodyHash: bodyHash(response.body), Duration: time.Since(start).Milliseconds(),
	}

	if response.status < 200 || response.status >= 400 {
		result.Error = fmt.Errorf("unexpected status code: %d", response.status)
		return result
	}

	score := s.similarityForIP(t.IP, response)
	result.Similar = score
	if score >= s.cfg.Similarity {
		result.Error = fmt.Errorf("response too similar to baseline: %d >= %d", score, s.cfg.Similarity)
	}
	return result
}

func (s *Scanner) fetchResponse(ctx context.Context, t task) (probeResponse, error) {
	targetURL, err := buildTargetURL(string(t.IP))
	if err != nil {
		return probeResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return probeResponse{}, fmt.Errorf("create request: %w", err)
	}

	req.Host = string(t.Host)
	req.Header.Set("Host", string(t.Host))
	req.Header.Set("User-Agent", "HostCollision/2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return probeResponse{}, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody+1))
	if err != nil {
		return probeResponse{}, fmt.Errorf("read response body: %w", err)
	}
	if len(body) > maxResponseBody {
		return probeResponse{}, fmt.Errorf("response body exceeds %d bytes", maxResponseBody)
	}
	return probeResponse{
		body:     body,
		status:   resp.StatusCode,
		location: resp.Header.Get("Location"),
	}, nil
}

func buildTargetURL(target string) (string, error) {
	host := target
	if addr, err := netip.ParseAddr(target); err == nil && addr.Is6() {
		host = net.JoinHostPort(target, "80")
	}

	targetURL := url.URL{Scheme: "http", Host: host, Path: "/"}
	if targetURL.Hostname() == "" {
		return "", fmt.Errorf("invalid target IP %q", target)
	}
	return targetURL.String(), nil
}

// bodyHash computes a SHA-1 hash of the given response body.
func bodyHash(body []byte) string {
	sum := sha1.Sum(body)
	return hex.EncodeToString(sum[:])
}
