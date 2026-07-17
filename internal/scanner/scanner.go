package scanner

import (
	"context"
	"hostCollision/internal/config"
	"hostCollision/internal/model"
	"hostCollision/internal/similarity"
	"net/http"
	"sync"
	"time"
)

const baselineHost = "hostcollision-baseline.invalid"

// Scanner coordinates the host collision scan across IP and host targets.
type Scanner struct {
	cfg config.Config

	httpClient *http.Client
	scanMu     sync.Mutex

	muSuccess    sync.Mutex
	successPerIP map[model.IP]int

	muBaseline   sync.Mutex
	baselineByIP map[model.IP]probeResponse
}

// task represents one scanning unit consisting of an IP and a host.
type task struct {
	IP   model.IP
	Host model.Host
}

// NewScanner creates a new Scanner with the given configuration.
func NewScanner(cfg config.Config) *Scanner {
	if cfg.Goroutines < 1 {
		cfg.Goroutines = 1
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		// A redirect may point at an unrelated host. Keep every probe pinned to
		// the target IP and evaluate the redirect response itself.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Scanner{
		cfg:          cfg,
		httpClient:   client,
		successPerIP: make(map[model.IP]int),
		baselineByIP: make(map[model.IP]probeResponse),
	}
}

// Scan is the package-level helper that creates a Scanner and runs the scan.
func Scan(ctx context.Context, ips []model.IP, hosts []model.Host, cfg config.Config) ([]model.Result, error) {
	s := NewScanner(cfg)
	return s.Scan(ctx, ips, hosts)
}

// Scan executes the host collision process for all IP and host combinations.
func (s *Scanner) Scan(ctx context.Context, ips []model.IP, hosts []model.Host) ([]model.Result, error) {
	// A Scanner may be reused, but two scans on the same instance must not
	// share counters or baselines. Serialize scans so resetting that state is
	// safe and deterministic.
	s.scanMu.Lock()
	defer s.scanMu.Unlock()
	s.resetScanState()

	if err := s.loadBaselines(ctx, ips); err != nil {
		return nil, err
	}

	tasks := make(chan task)
	results := make(chan model.Result)

	var wg sync.WaitGroup

	for i := 0; i < s.cfg.Goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx, tasks, results)
		}()
	}

	go func() {
		defer close(tasks)
		for _, ip := range ips {
			for _, host := range hosts {
				select {
				case <-ctx.Done():
					return
				case tasks <- task{IP: ip, Host: host}:
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []model.Result
	for r := range results {
		if r.Error == nil {
			out = append(out, r)
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Scanner) resetScanState() {
	s.muSuccess.Lock()
	s.successPerIP = make(map[model.IP]int)
	s.muSuccess.Unlock()

	s.muBaseline.Lock()
	s.baselineByIP = make(map[model.IP]probeResponse)
	s.muBaseline.Unlock()
}

// loadBaselines probes each IP with a deliberately invalid hostname. This
// prevents a real dictionary entry from being consumed as the baseline.
func (s *Scanner) loadBaselines(ctx context.Context, ips []model.IP) error {
	jobs := make(chan model.IP)
	var wg sync.WaitGroup
	workers := s.cfg.Goroutines
	if workers > len(ips) {
		workers = len(ips)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				response, err := s.fetchResponse(ctx, task{IP: ip, Host: model.Host(baselineHost)})
				if err != nil {
					// A missing baseline should not make the target unscannable.
					// Its candidates will receive similarity zero.
					continue
				}
				response.body = append([]byte(nil), response.body...)
				s.muBaseline.Lock()
				s.baselineByIP[ip] = response
				s.muBaseline.Unlock()
			}
		}()
	}

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return ctx.Err()
		case jobs <- ip:
		}
	}
	close(jobs)
	wg.Wait()
	return ctx.Err()
}

// reachedLimit reports whether the maximum number of successful hosts
// has been reached for the given IP.
func (s *Scanner) reachedLimit(ip model.IP) bool {
	s.muSuccess.Lock()
	defer s.muSuccess.Unlock()

	return s.successPerIP[ip] >= s.cfg.MaxHostsPerIP
}

// claimSuccess reserves a result slot without allowing concurrent workers to
// exceed the configured per-IP limit.
func (s *Scanner) claimSuccess(ip model.IP) bool {
	s.muSuccess.Lock()
	defer s.muSuccess.Unlock()
	if s.successPerIP[ip] >= s.cfg.MaxHostsPerIP {
		return false
	}
	s.successPerIP[ip]++
	return true
}

// similarityForIP returns the similarity score between the dedicated baseline
// response for the given IP and the candidate response. Redirect metadata is
// significant because redirects commonly have an empty or generic body.
func (s *Scanner) similarityForIP(ip model.IP, candidate probeResponse) int {
	s.muBaseline.Lock()
	base, ok := s.baselineByIP[ip]
	s.muBaseline.Unlock()
	if !ok {
		return 0
	}

	if isRedirect(base.status) || isRedirect(candidate.status) {
		if base.status != candidate.status || base.location != candidate.location {
			return 0
		}
	}

	return similarity.Score(base.body, candidate.body)
}

func isRedirect(status int) bool {
	return status >= http.StatusMultipleChoices && status < http.StatusBadRequest
}
