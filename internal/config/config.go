package config

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration parameters for the application.
type Config struct {
	IPFile        string        // Path to the IP list file.
	HostFile      string        // Path to the host dictionary file.
	OutputFile    string        // Path to the output file.
	Goroutines    int           // Number of worker goroutines.
	Sleep         time.Duration // Delay between requests.
	Similarity    int           // Similarity threshold in percentage (0-100).
	MaxHostsPerIP int           // Maximum number of successful hosts per IP/port.
	Ports         []int         // Ports for bare IP entries; nil defaults to port 80.
	HTTPSPorts    []int         // Ports using TLS; nil defaults to 443 and 8443.
	Insecure      bool          // Allow untrusted HTTPS certificates.
}

// FromFlags parses command-line flags and returns a validated Config.
func FromFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.IPFile, "i", "", "path to IP list file (required)")
	flag.StringVar(&cfg.HostFile, "d", "", "path to host dictionary file (required)")
	flag.StringVar(&cfg.OutputFile, "o", "", "path to output file (required)")

	flag.IntVar(&cfg.Goroutines, "n", 20, "number of goroutines")
	sleepMs := flag.Int("s", 1000, "sleep between requests in milliseconds")
	flag.IntVar(&cfg.Similarity, "r", 85, "similarity threshold (0-100)")
	flag.IntVar(&cfg.MaxHostsPerIP, "m", 50, "max successful hosts per IP/port")
	ports := flag.String("p", "", "comma-separated ports for bare IPs (default 80)")
	httpsPorts := flag.String("https-ports", "443,8443", "comma-separated HTTPS ports (empty for HTTP only)")
	flag.BoolVar(&cfg.Insecure, "insecure", false, "allow untrusted or mismatched HTTPS certificates")

	flag.Parse()

	cfg.Sleep = time.Duration(*sleepMs) * time.Millisecond

	if cfg.IPFile == "" || cfg.HostFile == "" || cfg.OutputFile == "" {
		flag.Usage()
		return nil, fmt.Errorf("missing required arguments: -i, -d and -o are required")
	}

	if cfg.Goroutines <= 0 {
		return nil, fmt.Errorf("goroutines (-n) must be greater than 0")
	}

	if *sleepMs < 0 {
		return nil, fmt.Errorf("sleep (-s) must not be negative")
	}

	if cfg.Similarity < 0 || cfg.Similarity > 100 {
		return nil, fmt.Errorf("similarity (-r) must be between 0 and 100")
	}

	if cfg.MaxHostsPerIP <= 0 {
		return nil, fmt.Errorf("max hosts per IP (-m) must be greater than 0")
	}
	var err error
	if cfg.Ports, err = parsePorts(*ports); err != nil {
		return nil, fmt.Errorf("invalid -p: %w", err)
	}
	if cfg.HTTPSPorts, err = parsePorts(*httpsPorts); err != nil {
		return nil, fmt.Errorf("invalid -https-ports: %w", err)
	}

	return cfg, nil
}

// parsePorts validates and deduplicates a comma-separated list in input order.
func parsePorts(value string) ([]int, error) {
	ports := make([]int, 0)
	if strings.TrimSpace(value) == "" {
		return ports, nil
	}
	seen := make(map[int]bool)
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" || strings.IndexFunc(item, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
			return nil, fmt.Errorf("invalid port %q; use integers from 1 to 65535", item)
		}
		port, err := strconv.Atoi(item)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid port %q; use integers from 1 to 65535", item)
		}
		if !seen[port] {
			ports = append(ports, port)
			seen[port] = true
		}
	}
	return ports, nil
}
