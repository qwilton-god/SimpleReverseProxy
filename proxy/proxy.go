package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxRequestBodySize  = 10 << 20
	defaultProxyTimeout = 30 * time.Second
)

type Proxy struct {
	client *http.Client
}

func NewProxy() *Proxy {
	return &Proxy{
		client: &http.Client{
			Timeout: defaultProxyTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, backend *Backend) {
	backendURL := backend.GetURL()

	proxyURL, err := p.buildProxyURL(backendURL, r)
	if err != nil {
		p.logAndError(w, "Failed to build proxy URL", err, http.StatusBadGateway)
		return
	}

	proxyReq, err := p.createProxyRequest(r, proxyURL)
	if err != nil {
		p.logAndError(w, "Failed to create proxy request", err, http.StatusBadGateway)
		return
	}

	p.copyHeaders(proxyReq.Header, r.Header, false)

	if r.Body != nil && r.ContentLength > 0 {
		if r.ContentLength > maxRequestBodySize {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		proxyReq.Body = r.Body
	}

	response, err := p.client.Do(proxyReq)
	if err != nil {
		p.logAndError(w, "Failed to reach backend", err, http.StatusBadGateway)
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	p.copyHeaders(w.Header(), response.Header, true)

	w.WriteHeader(response.StatusCode)

	if _, err := io.Copy(w, response.Body); err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

func (p *Proxy) buildProxyURL(backendURL *url.URL, r *http.Request) (*url.URL, error) {
	targetURL := *backendURL
	targetURL.Path = r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL.RawQuery = r.URL.RawQuery
	}
	return &targetURL, nil
}

func (p *Proxy) createProxyRequest(r *http.Request, targetURL *url.URL) (*http.Request, error) {
	ctx := r.Context()

	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	proxyReq.Host = r.Host
	proxyReq.ContentLength = r.ContentLength

	return proxyReq, nil
}

func (p *Proxy) copyHeaders(dst, src http.Header, skipHopByHop bool) {
	userAgent := ""
	if ua := src.Get("User-Agent"); ua != "" {
		userAgent = ua
	}

	for key, values := range src {
		if skipHopByHop && isHopByHopHeader(key) {
			continue
		}

		if key == "User-Agent" && userAgent != "" {
			dst.Set("User-Agent", userAgent)
			continue
		}

		for _, value := range values {
			dst.Add(key, value)
		}
	}

	if skipHopByHop {
		dst.Set("X-Forwarded-For", getRemoteIP(src.Get("X-Forwarded-For")))
		dst.Set("X-Forwarded-Host", src.Get("Host"))
		dst.Set("X-Forwarded-Proto", getScheme(src))
	}
}

func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	header = strings.ToLower(header)
	for _, h := range hopByHopHeaders {
		if strings.ToLower(h) == header {
			return true
		}
	}
	return false
}

func getRemoteIP(forwardedFor string) string {
	if forwardedFor != "" {
		return forwardedFor
	}
	return "unknown"
}

func getScheme(headers http.Header) string {
	if headers.Get("X-Forwarded-Proto") != "" {
		return headers.Get("X-Forwarded-Proto")
	}
	return "http"
}

func (p *Proxy) logAndError(w http.ResponseWriter, message string, err error, status int) {
	log.Printf("%s: %v", message, err)
	http.Error(w, http.StatusText(status), status)
}
