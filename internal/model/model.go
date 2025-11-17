package model

// IP represents an IP address used as a scan target.
type IP string

// Host represents a hostname used for collision attempts.
type Host string

// Target represents a single IP and host combination to be scanned.
type Target struct {
	IP   IP
	Host Host
}

// Result represents the outcome of a scan for a single IP and host pair.
type Result struct {
	IP      IP   // Scanned IP.
	Host    Host // Scanned host.
	Status  int  // HTTP status code returned by the target.
	Length  int  // Length of the HTTP response body in bytes.
	Similar int  // Similarity score in percentage (0-100).

	BodyHash string // Optional hash of the response body for deduplication.
	Duration int64  // Request duration in milliseconds.
	Error    error  // Error encountered during the request, if any.
}
