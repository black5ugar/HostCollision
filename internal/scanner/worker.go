package scanner

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hostCollision/internal/model"
	"io"
	"net/http"
	"time"
)

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
				continue
			}

			s.incrementSuccess(t.IP)

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
	url := fmt.Sprintf("http://%s/", string(t.IP))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.Result{
			IP:    t.IP,
			Host:  t.Host,
			Error: fmt.Errorf("create request: %w", err),
		}
	}

	req.Host = string(t.Host)
	req.Header.Set("Host", string(t.Host))
	req.Header.Set("User-Agent", "HostCollision/2.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return model.Result{
			IP:    t.IP,
			Host:  t.Host,
			Error: fmt.Errorf("execute request: %w", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Result{
			IP:    t.IP,
			Host:  t.Host,
			Error: fmt.Errorf("read response body: %w", err),
		}
	}

	result := model.Result{
		IP:       t.IP,
		Host:     t.Host,
		Status:   resp.StatusCode,
		Length:   len(body),
		BodyHash: bodyHash(body),
	}

	// Status code filter: treat 2xx and 3xx as candidates.
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return result
	}

	// Similarity filter: compare against per-IP baseline response.
	score := s.similarityForIP(t.IP, body)
	result.Similar = score

	if score >= s.cfg.Similarity {
		result.Error = fmt.Errorf("response too similar to baseline: %d >= %d", score, s.cfg.Similarity)
		return result
	}

	return result
}

// bodyHash computes a SHA-1 hash of the given response body.
func bodyHash(body []byte) string {
	sum := sha1.Sum(body)
	return hex.EncodeToString(sum[:])
}
