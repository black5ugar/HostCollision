package config

import (
	"flag"
	"fmt"
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
	MaxHostsPerIP int           // Maximum number of successful hosts per IP.
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
	flag.IntVar(&cfg.MaxHostsPerIP, "m", 50, "max successful hosts per IP")

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

	return cfg, nil
}
