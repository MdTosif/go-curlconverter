package parser

import (
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func tokenize(s string) []string {
	var out []string
	var cur strings.Builder
	inSingle, inDouble, esc := false, false, false
	tokenStarted := false
	canStartComment := true
	for _, r := range s {
		if esc {
			if r == '\n' || r == '\r' {
				esc = false
				canStartComment = true
				continue
			}
			cur.WriteRune(r)
			esc = false
			tokenStarted = true
			canStartComment = false
			continue
		}
		if r == '\\' {
			if !inSingle {
				esc = true
				continue
			}
			cur.WriteRune(r)
			tokenStarted = true
			canStartComment = false
			continue
		}
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			tokenStarted = true
			canStartComment = false
			continue
		}
		if r == '"' && !inSingle {
			inDouble = !inDouble
			tokenStarted = true
			canStartComment = false
			continue
		}
		if !inSingle && !inDouble && r == '#' && canStartComment {
			if tokenStarted {
				out = append(out, cur.String())
				cur.Reset()
				tokenStarted = false
			}
			canStartComment = true
			continue
		}
		if !inSingle && !inDouble && (r == '\r') {
			continue
		}
		if !inSingle && !inDouble && (r == ' ' || r == '\t' || r == '\n') {
			if tokenStarted {
				out = append(out, cur.String())
				cur.Reset()
				tokenStarted = false
			}
			canStartComment = true
			continue
		}
		cur.WriteRune(r)
		tokenStarted = true
		canStartComment = false
	}
	if tokenStarted {
		out = append(out, cur.String())
	}
	return out
}

func Parse(cmd string) (*request.Request, error) {
	toks := tokenize(stripCommentLines(strings.TrimSpace(cmd)))
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
	pendingReferer := ""

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
		if t == "-I" || t == "--head" {
			r.Method = "HEAD"
			explicitMethod = true
			continue
		}
		if t == "-T" || t == "--upload-file" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -T/--upload-file")
			}
			r.BodyFile = stripAtPrefix(toks[i+1])
			r.HasBody = true
			r.JSONBody = false
			r.Body = ""
			r.Method = "PUT"
			explicitMethod = true
			i++
			continue
		}
		if t == "-x" || t == "--proxy" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -x/--proxy")
			}
			r.Proxy = toks[i+1]
			i++
			continue
		}
		if t == "-A" || t == "--user-agent" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -A/--user-agent")
			}
			headerName := "User-Agent"
			if t == "--user-agent" {
				headerName = "user-agent"
			}
			setHeader(r, headerName, toks[i+1])
			i++
			continue
		}
		if t == "-e" || t == "--referer" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -e/--referer")
			}
			pendingReferer = toks[i+1]
			i++
			continue
		}
		if t == "--oauth2-bearer" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for --oauth2-bearer")
			}
			r.BearerToken = toks[i+1]
			i++
			continue
		}
		if t == "--digest" {
			r.DigestAuth = true
			continue
		}
		if t == "-U" || t == "--proxy-user" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -U/--proxy-user")
			}
			r.ProxyAuth = toks[i+1]
			i++
			continue
		}
		if t == "-G" || t == "--get" {
			appendQuery = true
			continue
		}
		if t == "--url" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for --url")
			}
			r.URLs = append(r.URLs, request.RequestURL{URL: toks[i+1]})
			i++
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
		if t == "-u" || t == "--user" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -u/--user")
			}
			r.BasicAuth = toks[i+1]
			i++
			continue
		}
		if t == "-d" || t == "--data" || t == "--data-raw" || t == "--data-binary" || t == "--data-ascii" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -d/--data")
			}
			r.HasBody = true
			value := toks[i+1]
			if (t == "--data-binary" || t == "-d") && strings.HasPrefix(value, "@") {
				r.BodyFile = stripAtPrefix(value)
				r.Body = ""
				r.JSONBody = false
			} else {
				dataParts = append(dataParts, value)
			}
			// if method is GET, curl will use POST when data is provided
			if r.Method == "GET" {
				r.Method = "POST"
			}
			i++
			continue
		}
		if t == "-F" || t == "--form" || t == "--form-string" {
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -F/--form")
			}
			part, err := parseFormPart(toks[i+1], t == "--form-string")
			if err != nil {
				return nil, err
			}
			r.FormParts = append(r.FormParts, part)
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
	if pendingReferer != "" && !hasHeader(r, "Referer") {
		setHeader(r, "Referer", pendingReferer)
	}
	if r.BodyFile != "" && methodUsesUploadTarget(r.Method) {
		for i, u := range r.URLs {
			r.URLs[i].URL = appendUploadFileToURL(u.URL, r.BodyFile)
		}
	}
	if len(jsonParts) > 0 {
		r.Body = strings.Join(jsonParts, "")
		if !hasHeader(r, "Content-Type") {
			setHeader(r, "Content-Type", "application/json")
		}
		if !hasHeader(r, "Accept") {
			setHeader(r, "Accept", "application/json")
		}
	} else if len(dataParts) > 0 && r.BodyFile == "" {
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

func stripCommentLines(cmd string) string {
	if cmd == "" {
		return cmd
	}
	lines := strings.Split(cmd, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
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

func stripAtPrefix(value string) string {
	return strings.TrimPrefix(value, "@")
}

func methodUsesUploadTarget(method string) bool {
	return strings.EqualFold(method, "PUT")
}

func appendUploadFileToURL(rawURL, fileName string) string {
	if fileName == "" {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	base := path.Base(parsed.Path)
	if base == "." || base == "/" || base == "" {
		parsed.Path = path.Join(parsed.Path, path.Base(fileName))
	}
	return parsed.String()
}

func parseFormPart(arg string, forceString bool) (request.FormPart, error) {
	name, value, ok := strings.Cut(arg, "=")
	if !ok || name == "" {
		return request.FormPart{}, errors.New("invalid argument for -F/--form")
	}

	part := request.FormPart{Name: name}
	if !forceString && strings.HasPrefix(value, "@") {
		part.IsFile = true
		part.FileName = strings.TrimPrefix(value, "@")
		return part, nil
	}

	part.Value = value
	return part, nil
}
