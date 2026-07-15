package scanner

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hostCollision/internal/model"
	"io"
	"log"
	"net/http"
	"time"
)

const maxResponseBody = 10 << 20 // 10 MiB

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
	body, status, err := s.fetchResponse(ctx, t)
	if err != nil {
		return model.Result{IP: t.IP, Host: t.Host, Error: err, Duration: time.Since(start).Milliseconds()}
	}

	result := model.Result{
		IP: t.IP, Host: t.Host, Status: status, Length: len(body),
		BodyHash: bodyHash(body), Duration: time.Since(start).Milliseconds(),
	}

	if status < 200 || status >= 400 {
		result.Error = fmt.Errorf("unexpected status code: %d", status)
		return result
	}

	score := s.similarityForIP(t.IP, body)
	result.Similar = score
	if score >= s.cfg.Similarity {
		result.Error = fmt.Errorf("response too similar to baseline: %d >= %d", score, s.cfg.Similarity)
	}
	return result
}

func (s *Scanner) fetchBody(ctx context.Context, t task) ([]byte, error) {
	body, _, err := s.fetchResponse(ctx, t)
	return body, err
}

func (s *Scanner) fetchResponse(ctx context.Context, t task) ([]byte, int, error) {
	url := fmt.Sprintf("http://%s/", string(t.IP))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Host = string(t.Host)
	req.Header.Set("Host", string(t.Host))
	req.Header.Set("User-Agent", "HostCollision/2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody+1))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}
	if len(body) > maxResponseBody {
		return nil, resp.StatusCode, fmt.Errorf("response body exceeds %d bytes", maxResponseBody)
	}
	return body, resp.StatusCode, nil
}

// bodyHash computes a SHA-1 hash of the given response body.
func bodyHash(body []byte) string {
	sum := sha1.Sum(body)
	return hex.EncodeToString(sum[:])
}
