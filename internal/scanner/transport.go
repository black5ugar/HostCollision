package scanner

import (
	"crypto/tls"
	"net"
	"net/http"
)

type probeTransport struct {
	http     *http.Transport
	insecure bool
}

func (t *probeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" {
		return t.http.RoundTrip(req)
	}
	host := req.Host
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	transport := t.http.Clone()
	// Each HTTPS probe uses its own TLS connection so a previous candidate's
	// SNI cannot select the virtual host for a later candidate on the same IP.
	transport.DisableKeepAlives = true
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	transport.TLSClientConfig.ServerName = host
	// The synthetic baseline hostname cannot have a trusted certificate.
	// Candidate certificates are verified unless -insecure is requested.
	transport.TLSClientConfig.InsecureSkipVerify = t.insecure || req.Host == baselineHost
	return transport.RoundTrip(req)
}

func (t *probeTransport) CloseIdleConnections() { t.http.CloseIdleConnections() }
