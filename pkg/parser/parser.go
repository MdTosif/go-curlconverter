package parser

import (
	"errors"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

// Parse a simple curl command into a Request. This parser is intentionally
// minimal and supports a subset of curl useful for basic conversions:
// -X / --request
// -H / --header
// -d / --data / --data-raw
// URL (http:// or https://)

func tokenize(s string) []string {
	var out []string
	var cur strings.Builder
	inSingle, inDouble, esc := false, false, false
	tokenStarted := false
	for _, r := range s {
		if esc {
			cur.WriteRune(r)
			esc = false
			tokenStarted = true
			continue
		}
		if r == '\\' {
			esc = true
			tokenStarted = true
			continue
		}
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			tokenStarted = true
			continue
		}
		if r == '"' && !inSingle {
			inDouble = !inDouble
			tokenStarted = true
			continue
		}
		if !inSingle && !inDouble && (r == ' ' || r == '\t' || r == '\n') {
			if tokenStarted {
				out = append(out, cur.String())
				cur.Reset()
				tokenStarted = false
			}
			continue
		}
		cur.WriteRune(r)
		tokenStarted = true
	}
	if tokenStarted {
		out = append(out, cur.String())
	}
	return out
}

func Parse(cmd string) (*request.Request, error) {
	toks := tokenize(strings.TrimSpace(cmd))
	if len(toks) == 0 {
		return nil, errors.New("empty command")
	}
	if toks[0] != "curl" {
		return nil, errors.New("command must start with 'curl'")
	}

	r := &request.Request{
		Method:   "GET",
		Headers:  map[string]string{},
		HeaderKV: []request.Header{},
	}
	dataParts := []string{}
	jsonParts := []string{}
	appendQuery := false
	explicitMethod := false

	for i := 1; i < len(toks); i++ {
		t := toks[i]
		if t == "-X" || t == "--request" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -X/--request")
			}
			r.Method = strings.ToUpper(toks[i+1])
			explicitMethod = true
			i++
			continue
		}
		if t == "-G" || t == "--get" {
			appendQuery = true
			continue
		}
		if t == "-H" || t == "--header" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -H/--header")
			}
			h := toks[i+1]
			// split on first ':'
			if idx := strings.Index(h, ":"); idx != -1 {
				k := strings.TrimSpace(h[:idx])
				v := strings.TrimSpace(h[idx+1:])
				setHeader(r, k, v)
			} else {
				setHeader(r, h, "")
			}
			i++
			continue
		}
		if t == "-b" || t == "--cookie" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -b/--cookie")
			}
			setHeader(r, "Cookie", toks[i+1])
			i++
			continue
		}
		if t == "-d" || t == "--data" || t == "--data-raw" || t == "--data-binary" || t == "--data-ascii" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -d/--data")
			}
			r.HasBody = true
			dataParts = append(dataParts, toks[i+1])
			// if method is GET, curl will use POST when data is provided
			if r.Method == "GET" {
				r.Method = "POST"
			}
			i++
			continue
		}
		if t == "--json" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for --json")
			}
			r.HasBody = true
			r.JSONBody = true
			jsonParts = append(jsonParts, toks[i+1])
			if r.Method == "GET" {
				r.Method = "POST"
			}
			i++
			continue
		}
		// basic URL detection
		if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
			r.URLs = append(r.URLs, request.RequestURL{URL: t})
			continue
		}
		// ignore unknown flags for MVP
	}

	if len(r.URLs) == 0 {
		return nil, errors.New("no URL found in command")
	}
	if len(jsonParts) > 0 {
		r.Body = strings.Join(jsonParts, "")
		if !hasHeader(r, "Content-Type") {
			setHeader(r, "Content-Type", "application/json")
		}
		if !hasHeader(r, "Accept") {
			setHeader(r, "Accept", "application/json")
		}
	} else if len(dataParts) > 0 {
		if appendQuery {
			for i, u := range r.URLs {
				r.URLs[i].URL = appendQueryString(u.URL, strings.Join(dataParts, "&"))
			}
			r.HasBody = false
			r.Body = ""
			if !explicitMethod {
				r.Method = "GET"
			}
		} else {
			r.Body = strings.Join(dataParts, "&")
		}
	}
	return r, nil
}

func setHeader(r *request.Request, key, value string) {
	r.Headers[key] = value
	for i, header := range r.HeaderKV {
		if strings.EqualFold(header.Key, key) {
			r.HeaderKV[i] = request.Header{Key: key, Value: value}
			return
		}
	}
	r.HeaderKV = append(r.HeaderKV, request.Header{Key: key, Value: value})
}

func hasHeader(r *request.Request, key string) bool {
	for _, header := range r.HeaderKV {
		if strings.EqualFold(header.Key, key) {
			return true
		}
	}
	return false
}

func appendQueryString(url, query string) string {
	if query == "" {
		return url
	}
	if strings.Contains(url, "?") {
		if strings.HasSuffix(url, "?") || strings.HasSuffix(url, "&") {
			return url + query
		}
		return url + "&" + query
	}
	return url + "?" + query
}
