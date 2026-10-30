// Command nanogpt-proxy is a transparent reverse proxy that forwards every
// request to the NanoGPT API as-is: same method, path, query, headers and body.
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// Headers that httputil.ReverseProxy strips in Rewrite mode; we put the
// client's originals back so nothing the client sent is dropped.
var forwardedHeaders = []string{"Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto"}

func main() {
	target, err := url.Parse(getenv("TARGET_URL", "https://nano-gpt.com"))
	if err != nil {
		log.Fatalf("invalid TARGET_URL: %v", err)
	}
	addr := ":" + getenv("PORT", "8080")

	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Don't add Accept-Encoding or transparently decompress responses;
	// pass compression through exactly as client and upstream negotiate it.
	transport.DisableCompression = true

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			// Points scheme/host at the target and sets the Host header to the
			// target host (required, otherwise upstream rejects the request).
			pr.SetURL(target)
			for _, h := range forwardedHeaders {
				if v, ok := pr.In.Header[h]; ok {
					pr.Out.Header[h] = v
				}
			}
		},
		Transport: transport,
		// Flush immediately so streaming (SSE) responses aren't buffered.
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("proxy error: %s %s: %v", r.Method, r.URL.Path, err)
			w.WriteHeader(http.StatusBadGateway)
		},
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           proxy,
		ReadHeaderTimeout: 30 * time.Second,
	}

	log.Printf("proxying %s -> %s", addr, target)
	log.Fatal(srv.ListenAndServe())
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
