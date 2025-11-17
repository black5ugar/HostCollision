package iohelper

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"hostCollision/internal/model"
	"os"
	"strconv"
	"strings"
)

// ReadIPs reads IP addresses from the specified file.
// Each non-empty trimmed line is treated as a single IP entry.
func ReadIPs(path string) ([]model.IP, error) {
	lines, err := readNonEmptyLines(path)
	if err != nil {
		return nil, err
	}

	ips := make([]model.IP, 0, len(lines))
	for _, line := range lines {
		ips = append(ips, model.IP(line))
	}
	return ips, nil
}

// ReadHosts reads hostnames from the specified file.
// Each non-empty trimmed line is treated as a single host entry.
func ReadHosts(path string) ([]model.Host, error) {
	lines, err := readNonEmptyLines(path)
	if err != nil {
		return nil, err
	}

	hosts := make([]model.Host, 0, len(lines))
	for _, line := range lines {
		hosts = append(hosts, model.Host(line))
	}
	return hosts, nil
}

// WriteResults writes scan results to the specified file in CSV format.
// The first row contains the header: ip, host, status, length, similar.
func WriteResults(path string, results []model.Result) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open output file %q: %w", path, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header row.
	if err := writer.Write([]string{"ip", "host", "status", "length", "similar"}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, r := range results {
		record := []string{
			string(r.IP),
			string(r.Host),
			strconv.Itoa(r.Status),
			strconv.Itoa(r.Length),
			strconv.Itoa(r.Similar),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write result for %s %s: %w", r.IP, r.Host, err)
		}
	}

	return nil
}

// readNonEmptyLines reads a file line by line and returns a slice of
// trimmed non-empty lines.
func readNonEmptyLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}

	return lines, nil
}
