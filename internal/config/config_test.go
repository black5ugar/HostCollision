package config

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestParsePorts(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  []int
	}{
		{"", []int{}},
		{" 80,443,80, 8080 ,65535", []int{80, 443, 8080, 65535}},
	} {
		got, err := parsePorts(tt.value)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("parsePorts(%q) = %v, %v; want %v", tt.value, got, err, tt.want)
		}
	}
	for _, value := range []string{"0", "65536", "-1", "+80", "http", "80,", ",443", "80,,443", "80-90"} {
		if _, err := parsePorts(value); err == nil {
			t.Errorf("accepted invalid ports %q", value)
		}
	}
}

func TestPortFlags(t *testing.T) {
	oldArgs, oldFlags := os.Args, flag.CommandLine
	t.Cleanup(func() { os.Args, flag.CommandLine = oldArgs, oldFlags })
	os.Args = []string{"hostcollision", "-i", "ip.txt", "-d", "host.txt", "-o", "out.csv", "-p", "80,443,80", "-https-ports", "443,9443", "-insecure"}
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := FromFlags()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.Ports, []int{80, 443}) || !reflect.DeepEqual(cfg.HTTPSPorts, []int{443, 9443}) || !cfg.Insecure {
		t.Fatalf("unexpected port flags: %#v", cfg)
	}
}
