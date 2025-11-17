package app

import (
	"context"
	"fmt"
	"hostCollision/internal/config"
	"hostCollision/internal/iohelper"
	"hostCollision/internal/scanner"
)

// Run is the main application entry point.
// It loads input data, executes the scan, prints a summary to stdout
// and writes the results to the output file.
func Run(ctx context.Context, cfg *config.Config) error {
	ips, err := iohelper.ReadIPs(cfg.IPFile)
	if err != nil {
		return fmt.Errorf("read IP file: %w", err)
	}

	hosts, err := iohelper.ReadHosts(cfg.HostFile)
	if err != nil {
		return fmt.Errorf("read host file: %w", err)
	}

	results, err := scanner.Scan(ctx, ips, hosts, *cfg)
	if err != nil {
		return fmt.Errorf("execute scan: %w", err)
	}

	// Print results to stdout in a simple table format.
	if len(results) > 0 {
		fmt.Println("IP,HOST,STATUS,LENGTH,SIMILAR")
		for _, r := range results {
			fmt.Printf(
				"%s,%s,%d,%d,%d\n",
				r.IP,
				r.Host,
				r.Status,
				r.Length,
				r.Similar,
			)
		}
	} else {
		fmt.Println("No valid results found.")
	}

	if err := iohelper.WriteResults(cfg.OutputFile, results); err != nil {
		return fmt.Errorf("write results: %w", err)
	}

	return nil
}
