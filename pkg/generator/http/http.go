package http

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	u := r.URLs[0].URL
	method := r.Method
	if method == "" {
		method = "GET"
	}

	var code strings.Builder

	// Parse URL to get path
	parsedURL, err := url.Parse(u)
	if err != nil {
		parsedURL = &url.URL{Path: "/"}
	}

	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	// Request line
	code.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", method, path))

	// Host header
	if parsedURL.Host != "" {
		code.WriteString(fmt.Sprintf("Host: %s\r\n", parsedURL.Host))
	}

	// Headers
	for _, h := range r.HeaderKV {
		code.WriteString(fmt.Sprintf("%s: %s\r\n", h.Key, h.Value))
	}

	// Content-Type if body exists
	if r.HasBody && r.Body != "" {
		hasContentType := false
		for _, h := range r.HeaderKV {
			if strings.EqualFold(h.Key, "Content-Type") {
				hasContentType = true
				break
			}
		}
		if !hasContentType {
			code.WriteString("Content-Type: application/x-www-form-urlencoded\r\n")
		}
	}

	// Cookie header
	if r.CookieJar != "" {
		code.WriteString(fmt.Sprintf("Cookie: %s\r\n", r.CookieJar))
	}

	// Auth
	if r.BasicAuth != "" {
		code.WriteString(fmt.Sprintf("Authorization: Basic %s\r\n", r.BasicAuth))
	}

	// Empty line before body
	code.WriteString("\r\n")

	// Body
	if r.HasBody && r.Body != "" {
		code.WriteString(r.Body)
	}

	return code.String()
}
