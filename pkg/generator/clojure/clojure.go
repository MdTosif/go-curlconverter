package clojure

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := []string{"(require '[clj-http.client :as client])"}

	methods := map[string]bool{"GET": true, "HEAD": true, "POST": true, "PUT": true, "DELETE": true, "OPTIONS": true, "COPY": true, "MOVE": true, "PATCH": true}
	dataMethods := map[string]bool{"POST": true, "PUT": true, "PATCH": true, "DELETE": true}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	params := make(map[string]string)
	var fn string

	if methods[strings.ToUpper(method)] {
		fn = "client/" + strings.ToLower(method)
	} else {
		fn = "client/request"
		params["url"] = ""
		params["method"] = reprStr(method)
	}

	url := r.URLs[0].URL
	queryParams := getQueryParams(r)
	if queryParams != "" {
		params["query-params"] = queryParams
		url = r.URLs[0].URLWithoutQueryList
	}

	params["headers"] = ""

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		authParam := "basic-auth"
		if r.DigestAuth {
			authParam = "digest-auth"
		}
		params[authParam] = "[" + reprStr(user) + " " + reprStr(pass) + "]"
	}

	if len(r.FormParts) > 0 {
		params["multipart"] = reprMultipart(r)
	} else if r.Body != "" {
		addData(params, r)
		if !dataMethods[strings.ToUpper(method)] {
			// Warning would go here
		}
	}

	if len(r.HeaderKV) > 0 {
		params["headers"] = reprHeaders(r)
	} else {
		delete(params, "headers")
	}

	if !r.Compressed {
		params[":decompress-body"] = "false"
	}

	if r.Insecure {
		params["insecure?"] = "true"
	}

	if !r.FollowRedirects {
		params["redirect-strategy"] = ":none"
	}

	if r.MaxRedirects != "" {
		params["max-redirects"] = r.MaxRedirects
	}

	if r.MaxTime != "" {
		timeout := times1000(r.MaxTime)
		params["socket-timeout"] = timeout
		params["connection-timeout"] = timeout
	}
	if r.ConnectTimeout != "" {
		params["connection-timeout"] = times1000(r.ConnectTimeout)
	}

	code := "(" + fn

	if fn == "client/request" {
		params["url"] = reprStr(url)
	} else {
		code += " " + reprStr(url)
	}

	paramLines := make([]string, 0, len(params))
	for param, value := range params {
		key := ":" + param + " "
		paramStr := key + indent(value, len(key))
		paramLines = append(paramLines, paramStr)
	}

	if len(paramLines) > 0 {
		paramStart := len(code) + 1
		if len(code) > 70 {
			paramStart = 1 + len(fn) + 1
			code += "\n" + strings.Repeat(" ", paramStart)
		} else {
			code += " "
		}
		code += indent("{" + strings.Join(paramLines, "\n") + "}", paramStart+1)
	}

	sort.Strings(imports)
	return strings.Join(imports, "\n") + "\n\n" + code + ")\n"
}

func getQueryParams(r *request.Request) string {
	if len(r.URLs) == 0 || len(r.URLs[0].QueryDict) == 0 {
		return ""
	}

	lines := []string{"{"}
	for _, q := range r.URLs[0].QueryDict {
		key := q[0]
		if !safeAsKeyword(key) {
			return ""
		}
		lines = append(lines, " :"+key+" "+reprStr(q[1]))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n ")
}

func reprHeaders(r *request.Request) string {
	lines := make([]string, 0, len(r.HeaderKV))
	for _, h := range r.HeaderKV {
		lines = append(lines, reprStr(h.Key)+" "+reprStr(h.Value))
	}
	return "{" + strings.Join(lines, ",\n ") + "}"
}

func reprMultipart(r *request.Request) string {
	parts := make([]string, 0, len(r.FormParts))
	for _, part := range r.FormParts {
		partStr := "{:name " + reprStr(part.Name)
		if part.IsFile {
			partStr += " :content (clojure.java.io/file " + reprStr(part.FileName) + ")"
		} else {
			partStr += " :content " + reprStr(part.Value)
		}
		parts = append(parts, partStr+"}")
	}
	return "[" + strings.Join(parts, "\n ") + "]"
}

func addData(params map[string]string, r *request.Request) {
	contentType := getContentType(r)
	if contentType == "application/json" {
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			if obj, ok := parsed.(map[string]any); ok {
				params["form-params"] = reprJson(obj)
				params["content-type"] = ":json"
				return
			}
		}
	}

	if contentType == "application/x-www-form-urlencoded" {
		values, err := url.ParseQuery(r.Body)
		if err == nil && len(values) > 0 {
			lines := []string{"{"}
			for k, vs := range values {
				if !safeAsKeyword(k) {
					continue
				}
				for _, v := range vs {
					lines = append(lines, " :"+k+" "+reprStr(v))
				}
			}
			lines = append(lines, "}")
			params["form-params"] = strings.Join(lines, "\n ")
			return
		}
	}

	params["body"] = reprStr(r.Body)
}

func reprJson(obj any) string {
	switch v := obj.(type) {
	case nil:
		return "nil"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return reprStr(v)
	case []any:
		if len(v) == 0 {
			return "[]"
		}
		reprs := make([]string, 0, len(v))
		for _, item := range v {
			reprs = append(reprs, reprJson(item))
		}
		return "[" + strings.Join(reprs, " ") + "]"
	case map[string]any:
		if len(v) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			if !safeAsKeyword(k) {
				return "{}"
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		reprs := make([]string, 0, len(keys))
		for _, k := range keys {
			reprs = append(reprs, ":"+k+" "+reprJson(v[k]))
		}
		return "{" + strings.Join(reprs, " ") + "}"
	default:
		return reprStr(fmt.Sprintf("%v", v))
	}
}

func reprStr(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}

func safeAsKeyword(s string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9*+!\-_'?<>=]*$`, s)
	return matched
}

func indent(s string, spaces int) string {
	indentStr := "\n" + strings.Repeat(" ", spaces)
	return strings.ReplaceAll(s, "\n", indentStr)
}

func times1000(s string) string {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return strconv.FormatFloat(f*1000, 'f', -1, 64)
	}
	return "(* (Float/parseFloat " + reprStr(s) + ") 1000)"
}

func getContentType(r *request.Request) string {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return strings.TrimSpace(strings.Split(h.Value, ";")[0])
		}
	}
	return ""
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
