package parser

import (
	"bytes"
	"encoding/json"
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
	commentLine := false
	for _, r := range s {
		if commentLine {
			if r == '\n' {
				commentLine = false
				canStartComment = true
			}
			continue
		}
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
			commentLine = true
			continue
		}
		if !inSingle && !inDouble && r == '\r' {
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

func ParseAll(cmd string) ([]*request.Request, error) {
	toks := tokenize(strings.TrimSpace(cmd))
	if len(toks) == 0 {
		return nil, errors.New("empty command")
	}
	if toks[0] != "curl" {
		return nil, errors.New("command must start with 'curl'")
	}

	r := &request.Request{
		Method:        "GET",
		AuthType:      "basic",
		ProxyAuthType: "basic",
		Headers:       map[string]string{},
		HeaderKV:      []request.Header{},
		HeadersOut:    request.HeaderSet{Lowercase: false, Headers: [][2]string{}},
		ProxyHeaders:  request.HeaderSet{Lowercase: false, Headers: [][2]string{}},
		Cookies:       nil,
		CookieFiles:   nil,
		Compressed:    false,
		DataArray:     nil,
		IsDataRaw:     nil,
		IsDataBinary:  nil,
		URLs:          nil,
		FormParts:     nil,
	}
	dataParts := []string{}
	jsonParts := []string{}
	appendQuery := false
	explicitMethod := false
	pendingReferer := ""
	lastDataFlag := ""

	for i := 1; i < len(toks); i++ {
		t := toks[i]
		if strings.HasPrefix(t, "-X") && len(t) > 2 && t != "-X" {
			r.Method = t[2:]
			explicitMethod = true
			continue
		}
		switch t {
		case "-X", "--request":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -X/--request")
			}
			r.Method = toks[i+1]
			explicitMethod = true
			i++
			continue
		case "-I", "--head":
			r.Method = "HEAD"
			explicitMethod = true
			continue
		case "-T", "--upload-file":
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
		case "-x", "--proxy":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -x/--proxy")
			}
			r.Proxy = toks[i+1]
			i++
			continue
		case "-A", "--user-agent":
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
		case "-e", "--referer":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -e/--referer")
			}
			pendingReferer = toks[i+1]
			i++
			continue
		case "--oauth2-bearer":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for --oauth2-bearer")
			}
			r.BearerToken = toks[i+1]
			i++
			continue
		case "--digest":
			r.DigestAuth = true
			continue
		case "-U", "--proxy-user":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -U/--proxy-user")
			}
			r.ProxyAuth = toks[i+1]
			i++
			continue
		case "-G", "--get":
			appendQuery = true
			continue
		case "--compressed":
			r.Compressed = true
			continue
		case "--url":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for --url")
			}
			r.URLs = append(r.URLs, request.RequestURL{OriginalURL: toks[i+1]})
			i++
			continue
		case "-H", "--header":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -H/--header")
			}
			h := toks[i+1]
			if idx := strings.Index(h, ":"); idx != -1 {
				k := strings.TrimSpace(h[:idx])
				v := strings.TrimSpace(h[idx+1:])
				setHeader(r, k, v)
			} else {
				setHeader(r, h, "")
			}
			i++
			continue
		case "-b", "--cookie":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -b/--cookie")
			}
			cookieArg := toks[i+1]
			if looksLikeCookieFile(cookieArg) {
				r.CookieFiles = append(r.CookieFiles, cookieArg)
			} else {
				setHeader(r, "Cookie", cookieArg)
			}
			i++
			continue
		case "-u", "--user":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -u/--user")
			}
			r.BasicAuth = toks[i+1]
			i++
			continue
		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -d/--data")
			}
			r.HasBody = true
			value := toks[i+1]
			lastDataFlag = t
			if (t == "--data-binary" || t == "-d") && strings.HasPrefix(value, "@") {
				r.BodyFile = stripAtPrefix(value)
				r.Body = ""
				r.JSONBody = false
			} else {
				dataParts = append(dataParts, value)
			}
			if !explicitMethod && r.Method == "GET" {
				r.Method = "POST"
			}
			i++
			continue
		case "-F", "--form", "--form-string":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for -F/--form")
			}
			part, err := parseFormPart(toks[i+1], t == "--form-string")
			if err != nil {
				return nil, err
			}
			r.FormParts = append(r.FormParts, part)
			if !explicitMethod && r.Method == "GET" {
				r.Method = "POST"
			}
			i++
			continue
		case "--json":
			if i+1 >= len(toks) {
				return nil, errors.New("missing argument for --json")
			}
			r.HasBody = true
			r.JSONBody = true
			jsonParts = append(jsonParts, toks[i+1])
			if !explicitMethod && r.Method == "GET" {
				r.Method = "POST"
			}
			i++
			continue
		}

		if !strings.HasPrefix(t, "-") {
			r.URLs = append(r.URLs, request.RequestURL{OriginalURL: t})
		}
	}

	if len(r.URLs) == 0 {
		return nil, errors.New("no URL found in command")
	}
	if pendingReferer != "" && !hasHeader(r, "Referer") {
		setHeader(r, "Referer", pendingReferer)
	}
	if r.BodyFile != "" && methodUsesUploadTarget(r.Method) {
		for i, u := range r.URLs {
			r.URLs[i].URL = appendUploadFileToURL(normalizeURL(u.OriginalURL), r.BodyFile)
		}
	}
	if len(jsonParts) > 0 {
		r.Body = strings.Join(jsonParts, "")
		r.Data = r.Body
		r.DataArray = []string{r.Body}
		boolFalse := false
		r.IsDataRaw = &boolFalse
		r.IsDataBinary = &boolFalse
		if !hasHeader(r, "Content-Type") {
			setHeader(r, "Content-Type", "application/json")
		}
		if !hasHeader(r, "Accept") {
			setHeader(r, "Accept", "application/json")
		}
	} else if len(dataParts) > 0 && r.BodyFile == "" {
		joined := strings.Join(dataParts, "&")
		if appendQuery {
			for i, u := range r.URLs {
				r.URLs[i].URL = appendQueryString(normalizeURL(firstNonEmpty(u.URL, u.OriginalURL)), joined)
			}
			r.HasBody = false
			r.Body = ""
			if !explicitMethod {
				r.Method = "GET"
			}
		} else {
			if !hasHeader(r, "Content-Type") {
				setHeader(r, "Content-Type", "application/x-www-form-urlencoded")
				r.AutoContentType = true
			}
			r.Body = joined
			r.Data = joined
			r.DataArray = []string{joined}
			boolRaw := lastDataFlag == "--data-raw"
			boolBinary := lastDataFlag == "--data-binary"
			r.IsDataRaw = &boolRaw
			r.IsDataBinary = &boolBinary
		}
	}

	for i := range r.URLs {
		if r.URLs[i].URL == "" {
			r.URLs[i].URL = normalizeURL(r.URLs[i].OriginalURL)
		}
		populateURLFields(&r.URLs[i], r.Method)
		if r.URLs[i].QueryList != nil {
			r.URLs[i].QueryArray = pairsToQueryArray(r.URLs[i].QueryList)
			r.URLs[i].URLQueryArray = pairsToQueryArray(r.URLs[i].QueryList)
		}
		if appendQuery && len(dataParts) > 0 {
			r.URLs[i].QueryArray = append([]string{}, pairsToQueryArray(r.URLs[i].QueryList)...)
		}
	}

	if cookieHeader, ok := getHeader(r, "Cookie"); ok {
		r.Cookies = parseCookies(cookieHeader)
	}

	return []*request.Request{r}, nil
}

func Parse(cmd string) (*request.Request, error) {
	reqs, err := ParseAll(cmd)
	if err != nil {
		return nil, err
	}
	return reqs[0], nil
}

func MarshalJSON(reqs []*request.Request) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(reqs); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func setHeader(r *request.Request, key, value string) {
	r.Headers[key] = value
	for i, header := range r.HeaderKV {
		if strings.EqualFold(header.Key, key) {
			r.HeaderKV[i] = request.Header{Key: key, Value: value}
			r.HeadersOut.Headers[i] = [2]string{key, value}
			return
		}
	}
	r.HeaderKV = append(r.HeaderKV, request.Header{Key: key, Value: value})
	r.HeadersOut.Headers = append(r.HeadersOut.Headers, [2]string{key, value})
}

func getHeader(r *request.Request, key string) (string, bool) {
	for _, header := range r.HeaderKV {
		if strings.EqualFold(header.Key, key) {
			return header.Value, true
		}
	}
	return "", false
}

func hasHeader(r *request.Request, key string) bool {
	_, ok := getHeader(r, key)
	return ok
}

func appendQueryString(rawURL, query string) string {
	if query == "" {
		return rawURL
	}
	if strings.Contains(rawURL, "?") {
		if strings.HasSuffix(rawURL, "?") || strings.HasSuffix(rawURL, "&") {
			return rawURL + query
		}
		return rawURL + "&" + query
	}
	return rawURL + "?" + query
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

func normalizeURL(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if raw == "" {
		return raw
	}
	return "http://" + raw
}

func populateURLFields(u *request.RequestURL, method string) {
	parsed, err := url.Parse(u.URL)
	if err != nil {
		u.Method = method
		return
	}
	queryWithPrefix := ""
	if parsed.RawQuery != "" {
		queryWithPrefix = "?" + parsed.RawQuery
	}
	urlWithoutQuery := strings.TrimSuffix(u.URL, queryWithPrefix)
	u.URLObj = request.URLObject{
		Scheme:   parsed.Scheme,
		Host:     parsed.Host,
		Port:     "",
		Path:     parsed.EscapedPath(),
		Query:    queryWithPrefix,
		Fragment: parsed.Fragment,
	}
	u.URLWithoutQueryList = urlWithoutQuery
	u.URLWithOriginalQuery = u.URL
	u.URLWithoutQueryArray = urlWithoutQuery
	u.Method = method

	if parsed.RawQuery == "" {
		return
	}
	list := parseQueryString(parsed.RawQuery)
	if len(list) == 0 {
		return
	}
	u.QueryList = list
	if queryDictSequential(list) {
		u.QueryDict = append([][2]string{}, list...)
	}
}

func parseQueryString(raw string) [][2]string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "&")
	ret := make([][2]string, 0, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			key = part
			value = ""
		}
		ret = append(ret, [2]string{key, value})
	}
	return ret
}

func pairsToQueryArray(pairs [][2]string) []string {
	ret := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		ret = append(ret, pair[0]+"="+pair[1])
	}
	return ret
}

func queryDictSequential(pairs [][2]string) bool {
	seen := map[string]bool{}
	lastKey := ""
	for _, pair := range pairs {
		if seen[pair[0]] && pair[0] != lastKey {
			return false
		}
		seen[pair[0]] = true
		lastKey = pair[0]
	}
	return true
}

func parseCookies(raw string) [][2]string {
	parts := strings.Split(raw, ";")
	ret := make([][2]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			ret = append(ret, [2]string{part, ""})
			continue
		}
		ret = append(ret, [2]string{strings.TrimSpace(key), strings.TrimSpace(value)})
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func looksLikeCookieFile(value string) bool {
	return !strings.Contains(value, "=") && !strings.Contains(value, ";")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
