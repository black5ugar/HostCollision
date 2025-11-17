package main

import (
	"fmt"
	"log"
	"net/http"
)

// handler returns different responses based on the Host header.
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Host {
	case "www.aaa.com":
		fmt.Fprintln(w, "<html><body><h1>Site AAA</h1><p>Welcome to AAA</p></body></html>")
	case "www.bbb.com":
		fmt.Fprintln(w, "<html><body><h1>Site BBB</h1><p>Welcome to BBB - secret site</p></body></html>")
	default:
		fmt.Fprintln(w, "<html><body><h1>Default Site</h1><p>This is default vhost</p></body></html>")
	}
}

func main() {
	http.HandleFunc("/", handler)

	// Listen on port 80. This usually requires root privileges.
	addr := ":80"
	log.Printf("Test server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
