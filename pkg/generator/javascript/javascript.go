package javascript

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
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
	if len(r.FormParts) > 0 && method == "GET" {
		method = "POST"
	}

	needsReadFile := requestNeedsReadFile(r)
	imports := make([]string, 0, 1)
	setup := make([]string, 0, 2)
	clientCall := "fetch"
	if needsReadFile {
		imports = append(imports, "import { readFile } from 'node:fs/promises';")
	}
	if len(r.FormParts) > 0 {
		setup = append(setup, renderFormData(r.FormParts))
	}
	if r.DigestAuth {
		user, pass := splitBasicAuth(r.BasicAuth)
		imports = append(imports, "import * as DigestFetch from 'digest-fetch';")
		setup = append(setup, "const client = new DigestFetch('"+escapeSingle(user)+"', '"+escapeSingle(pass)+"');")
		clientCall = "client.fetch"
	}

	lines := make([]string, 0, 4)
	if method != "GET" {
		lines = append(lines, "  method: '"+escapeSingle(method)+"'")
	}

	headers := append([]request.Header{}, r.HeaderKV...)
	if r.BearerToken != "" && !r.DigestAuth && !hasHeader(headers, "Authorization") {
		headers = append([]request.Header{{Key: "Authorization", Value: "Bearer " + r.BearerToken}}, headers...)
	} else if r.BasicAuth != "" && !r.DigestAuth && !hasHeader(headers, "Authorization") {
		headers = append([]request.Header{{Key: "Authorization", Value: ""}}, headers...)
	}
	if r.HasBody && r.BodyFile == "" && !hasHeader(headers, "Content-Type") {
		headers = append(headers, request.Header{
			Key:   "Content-Type",
			Value: "application/x-www-form-urlencoded",
		})
	}
	if len(headers) > 0 {
		headerLines := make([]string, 0, len(headers))
		for _, header := range headers {
			if header.Key == "Authorization" && r.BasicAuth != "" && header.Value == "" {
				headerLines = append(
					headerLines,
					"    'Authorization': 'Basic ' + btoa('"+escapeSingle(r.BasicAuth)+"')",
				)
				continue
			}
			headerLines = append(
				headerLines,
				"    '"+escapeSingle(header.Key)+"': '"+escapeSingle(header.Value)+"'",
			)
		}
		lines = append(lines, "  headers: {\n"+strings.Join(headerLines, ",\n")+"\n  }")
	}
	if len(r.FormParts) > 0 {
		lines = append(lines, "  body: form")
	} else if r.BodyFile != "" {
		lines = append(lines, "  body: await readFile('"+escapeSingle(r.BodyFile)+"')")
	} else if r.HasBody {
		if r.JSONBody {
			lines = append(lines, renderJSONBody(r.Body)...)
		} else if shouldUseURLSearchParams(r, headers, r.Body) {
			lines = append(lines, renderURLSearchParams(r.Body))
		} else {
			lines = append(lines, "  body: '"+escapeSingle(r.Body)+"'")
		}
	}

	body := clientCall + "('" + escapeSingle(u) + "');\n"
	if len(lines) > 0 {
		body = clientCall + "('" + escapeSingle(u) + "', {\n" + strings.Join(lines, ",\n") + "\n});\n"
	}

	sections := make([]string, 0, 3)
	if len(imports) > 0 {
		sections = append(sections, strings.Join(imports, "\n"))
	}
	mainSection := strings.TrimRight(body, "\n")
	if len(setup) > 0 {
		setupText := strings.Join(setup, "\n")
		separator := "\n\n"
		if r.DigestAuth && len(r.FormParts) == 0 {
			separator = "\n"
		}
		mainSection = setupText + separator + mainSection
	}
	sections = append(sections, mainSection)
	return strings.Join(sections, "\n\n") + "\n"
}

func escapeSingle(s string) string {
	// very small escape for single quotes and backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func hasHeader(headers []request.Header, name string) bool {
	for _, header := range headers {
		if strings.EqualFold(header.Key, name) {
			return true
		}
	}
	return false
}

func renderJSONBody(raw string) []string {
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return []string{"  body: '" + escapeSingle(raw) + "'"}
	}

	return []string{
		"  // body: '" + escapeSingle(raw) + "'",
		"  body: JSON.stringify(" + jsJSONValue(parsed, 1) + ")",
	}
}

func renderURLSearchParams(raw string) string {
	pairs := strings.Split(raw, "&")
	lines := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		key, value, _ := strings.Cut(pair, "=")
		lines = append(lines, "    '"+escapeSingle(key)+"': '"+escapeSingle(value)+"'")
	}
	return "  body: new URLSearchParams({\n" + strings.Join(lines, ",\n") + "\n  })"
}

func shouldUseURLSearchParams(r *request.Request, headers []request.Header, raw string) bool {
	if raw == "" {
		return false
	}
	if _, ok := r.Headers["Content-Type"]; !ok {
		return false
	}
	if !hasFormURLEncodedHeader(headers) {
		return false
	}
	pairs := strings.Split(raw, "&")
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || key == "" {
			return false
		}
		if strings.ContainsAny(key, " {}[]\"'") || strings.ContainsAny(value, " {}[]\"'") {
			return false
		}
	}
	return true
}

func hasFormURLEncodedHeader(headers []request.Header) bool {
	for _, header := range headers {
		if strings.EqualFold(header.Key, "Content-Type") {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(header.Value, ";")[0]))
			return mediaType == "application/x-www-form-urlencoded"
		}
	}
	return false
}

func renderFormData(parts []request.FormPart) string {
	lines := []string{"const form = new FormData();"}
	for _, part := range parts {
		if part.IsFile {
			lines = append(lines, "form.append('"+escapeSingle(part.Name)+"', new File([await readFile('"+escapeSingle(part.FileName)+"')], '"+escapeSingle(path.Base(part.FileName))+"'));")
		} else {
			lines = append(lines, "form.append('"+escapeSingle(part.Name)+"', '"+escapeSingle(part.Value)+"');")
		}
	}
	return strings.Join(lines, "\n")
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func requestNeedsReadFile(r *request.Request) bool {
	if r.BodyFile != "" {
		return true
	}
	for _, part := range r.FormParts {
		if part.IsFile {
			return true
		}
	}
	return false
}

func jsJSONValue(v any, indentLevel int) string {
	indent := strings.Repeat("  ", indentLevel)
	nextIndent := strings.Repeat("  ", indentLevel+1)

	switch value := v.(type) {
	case map[string]any:
		if len(value) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines := make([]string, 0, len(keys))
		for _, key := range keys {
			lines = append(lines, nextIndent+"'"+escapeSingle(key)+"': "+jsJSONValue(value[key], indentLevel+1))
		}
		return "{\n" + strings.Join(lines, ",\n") + "\n" + indent + "}"
	case []any:
		if len(value) == 0 {
			return "[]"
		}
		lines := make([]string, 0, len(value))
		for _, item := range value {
			lines = append(lines, nextIndent+jsJSONValue(item, indentLevel+1))
		}
		return "[\n" + strings.Join(lines, ",\n") + "\n" + indent + "]"
	case string:
		return "'" + escapeSingle(value) + "'"
	case bool:
		if value {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	case float64:
		return fmt.Sprintf("%v", value)
	default:
		return "'" + escapeSingle(fmt.Sprintf("%v", value)) + "'"
	}
}
