package scanner

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"hostCollision/internal/model"
)

// endpoint identifies a service, keeping its baseline and hit limit separate.
type endpoint struct {
	IP     model.IP
	Port   int
	Scheme string
}

func (e endpoint) task(host model.Host) task {
	return task{IP: e.IP, Port: e.Port, Scheme: e.Scheme, Host: host}
}

func (t task) endpoint() endpoint {
	return endpoint{IP: t.IP, Port: t.Port, Scheme: t.Scheme}
}

func (s *Scanner) expandTargets(ips []model.IP) ([]endpoint, error) {
	ports := s.cfg.Ports
	if len(ports) == 0 {
		ports = []int{80}
	}
	for _, port := range append(append([]int(nil), ports...), s.cfg.HTTPSPorts...) {
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid port %d", port)
		}
	}
	var out []endpoint
	seen := make(map[endpoint]bool)
	for _, input := range ips {
		ip, explicitPort, err := parseTarget(string(input))
		if err != nil {
			return nil, err
		}
		targetPorts := ports
		if explicitPort != 0 {
			targetPorts = []int{explicitPort}
		}
		for _, port := range targetPorts {
			e := endpoint{IP: model.IP(ip), Port: port, Scheme: s.schemeForPort(port)}
			if !seen[e] {
				out = append(out, e)
				seen[e] = true
			}
		}
	}
	return out, nil
}

func parseTarget(input string) (string, int, error) {
	if ip, err := netip.ParseAddr(input); err == nil {
		return ip.String(), 0, nil
	}
	// Also accept bracketed IPv6 without an explicit port.
	if strings.HasPrefix(input, "[") && strings.HasSuffix(input, "]") {
		if ip, err := netip.ParseAddr(input[1 : len(input)-1]); err == nil && ip.Is6() {
			return ip.String(), 0, nil
		}
	}
	host, portText, err := net.SplitHostPort(input)
	if err != nil {
		return "", 0, fmt.Errorf("invalid target IP %q", input)
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return "", 0, fmt.Errorf("invalid target IP %q", input)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("invalid target port in %q", input)
	}
	return ip.String(), port, nil
}

func (s *Scanner) schemeForPort(port int) string {
	for _, p := range s.cfg.HTTPSPorts {
		if p == port {
			return "https"
		}
	}
	return "http"
}

func (s *Scanner) resolveTask(t task) (task, error) {
	ip, port, err := parseTarget(string(t.IP))
	if err != nil {
		return t, err
	}
	t.IP = model.IP(ip)
	if t.Port == 0 {
		t.Port = port
	}
	if t.Port == 0 {
		t.Port = 80
	}
	if t.Port < 1 || t.Port > 65535 {
		return t, fmt.Errorf("invalid target port %d", t.Port)
	}
	if t.Scheme == "" {
		t.Scheme = s.schemeForPort(t.Port)
	}
	if t.Scheme != "http" && t.Scheme != "https" {
		return t, fmt.Errorf("invalid scheme %q", t.Scheme)
	}
	return t, nil
}

func targetURL(t task) string {
	u := url.URL{Scheme: t.Scheme, Host: net.JoinHostPort(string(t.IP), strconv.Itoa(t.Port)), Path: "/"}
	return u.String()
}
