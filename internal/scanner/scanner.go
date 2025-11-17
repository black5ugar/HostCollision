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

// Scanner coordinates the host collision scan across IP and host targets.
type Scanner struct {
	cfg config.Config

	httpClient *http.Client

	muSuccess    sync.Mutex
	successPerIP map[model.IP]int

	muBaseline   sync.Mutex
	baselineByIP map[model.IP][]byte
}

// task represents one scanning unit consisting of an IP and a host.
type task struct {
	IP   model.IP
	Host model.Host
}

// NewScanner creates a new Scanner with the given configuration.
func NewScanner(cfg config.Config) *Scanner {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Scanner{
		cfg:          cfg,
		httpClient:   client,
		successPerIP: make(map[model.IP]int),
		baselineByIP: make(map[model.IP][]byte),
	}
}

// Scan is the package-level helper that creates a Scanner and runs the scan.
func Scan(ctx context.Context, ips []model.IP, hosts []model.Host, cfg config.Config) ([]model.Result, error) {
	s := NewScanner(cfg)
	return s.Scan(ctx, ips, hosts)
}

// Scan executes the host collision process for all IP and host combinations.
func (s *Scanner) Scan(ctx context.Context, ips []model.IP, hosts []model.Host) ([]model.Result, error) {
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

	return out, nil
}

// reachedLimit reports whether the maximum number of successful hosts
// has been reached for the given IP.
func (s *Scanner) reachedLimit(ip model.IP) bool {
	s.muSuccess.Lock()
	defer s.muSuccess.Unlock()

	return s.successPerIP[ip] >= s.cfg.MaxHostsPerIP
}

// incrementSuccess increments the successful host count for the given IP.
func (s *Scanner) incrementSuccess(ip model.IP) {
	s.muSuccess.Lock()
	s.successPerIP[ip]++
	s.muSuccess.Unlock()
}

// similarityForIP returns the similarity score between the baseline response
// for the given IP and the provided body. The first successful response for
// an IP is used as the baseline.
func (s *Scanner) similarityForIP(ip model.IP, body []byte) int {
	s.muBaseline.Lock()
	defer s.muBaseline.Unlock()

	base, ok := s.baselineByIP[ip]
	if !ok {
		copied := make([]byte, len(body))
		copy(copied, body)
		s.baselineByIP[ip] = copied
		return 100
	}

	return similarity.Score(base, body)
}
