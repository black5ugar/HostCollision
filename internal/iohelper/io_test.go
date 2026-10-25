package iohelper

import (
	"encoding/csv"
	"hostCollision/internal/model"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestReadIPs verifies that ReadIPs correctly parses non-empty lines.
func TestReadIPs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ip.txt")

	content := "10.0.0.1\n\n 192.168.1.1 \n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	ips, err := ReadIPs(path)
	if err != nil {
		t.Fatalf("ReadIPs returned error: %v", err)
	}

	if len(ips) != 2 {
		t.Fatalf("expected 2 IPs, got %d", len(ips))
	}

	if ips[0] != model.IP("10.0.0.1") || ips[1] != model.IP("192.168.1.1") {
		t.Fatalf("unexpected IPs: %#v", ips)
	}
}

// TestReadHosts verifies that ReadHosts correctly parses non-empty lines.
func TestReadHosts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "host.txt")

	content := "www.example.com\n\n test.example.org \n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hosts, err := ReadHosts(path)
	if err != nil {
		t.Fatalf("ReadHosts returned error: %v", err)
	}

	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	if hosts[0] != model.Host("www.example.com") || hosts[1] != model.Host("test.example.org") {
		t.Fatalf("unexpected hosts: %#v", hosts)
	}
}

// TestWriteResults verifies the CSV schema and title escaping, including an
// absent title and a title containing Unicode, commas, and quotation marks.
func TestWriteResults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.txt")

	results := []model.Result{
		{
			IP:      model.IP("10.0.0.1"),
			Host:    model.Host("www.example.com"),
			Status:  200,
			Length:  123,
			Similar: 42,
			Title:   `用户中心, "Settings"`,
		},
		{
			IP:      model.IP("192.168.1.1"),
			Host:    model.Host("test.example.org"),
			Status:  301,
			Length:  456,
			Similar: 10,
		},
	}

	if err := WriteResults(path, results); err != nil {
		t.Fatalf("WriteResults returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV output: %v", err)
	}
	want := [][]string{
		{"ip", "host", "status", "length", "similar", "title"},
		{"10.0.0.1", "www.example.com", "200", "123", "42", `用户中心, "Settings"`},
		{"192.168.1.1", "test.example.org", "301", "456", "10", ""},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("unexpected CSV records: got %#v, want %#v", records, want)
	}
}
