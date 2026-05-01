package nodefetch

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := []string{}
	if requestNeedsFileFromSync(r) {
		imports = append(imports, "import fetch, { fileFromSync } from 'node-fetch';")
	} else {
		imports = append(imports, "import fetch from 'node-fetch';")
	}

	setup := []string{}
	if len(r.FormParts) > 0 {
		setup = append(setup, renderFormData(r.FormParts))
	}

	lines := []string{}
	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	if method != "GET" {
		lines = append(lines, "  method: "+reprString(method))
	}

	headers := append([]request.Header{}, r.HeaderKV...)
	if r.BearerToken != "" && !hasHeader(headers, "Authorization") {
		headers = append([]request.Header{{Key: "Authorization", Value: "Bearer " + r.BearerToken}}, headers...)
	} else if r.BasicAuth != "" && !hasHeader(headers, "Authorization") {
		headers = append([]request.Header{{Key: "Authorization", Value: ""}}, headers...)
	}
	if r.HasBody && r.BodyFile == "" && !hasHeader(headers, "Content-Type") {
		headers = append(headers, request.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"})
	}
	if len(headers) > 0 {
		headerLines := make([]string, 0, len(headers))
		for _, h := range headers {
			if h.Key == "Authorization" && r.BasicAuth != "" && h.Value == "" {
				headerLines = append(headerLines, "    'Authorization': 'Basic ' + btoa("+reprString(r.BasicAuth)+")")
				continue
			}
			headerLines = append(headerLines, "    "+reprString(h.Key)+": "+renderJSExpr(h.Value))
		}
		lines = append(lines, "  headers: {\n"+strings.Join(headerLines, ",\n")+"\n  }")
	}

	if r.Proxy != "" {
		proxyImport, proxyLine := renderProxy(r.Proxy)
		if proxyImport != "" {
			imports = append(imports, proxyImport)
		}
		if proxyLine != "" {
			lines = append(lines, "  "+proxyLine)
		}
	}

	switch {
	case len(r.FormParts) > 0:
		lines = append(lines, "  body: form")
	case r.BodyFile != "":
		lines = append(lines, "  body: fileFromSync("+renderJSExpr(r.BodyFile)+")")
	case r.HasBody:
		lines = append(lines, renderBody(r)...)
	}

	urlExpr := renderJSExpr(r.URLs[0].URL)
	body := "fetch(" + urlExpr + ");\n"
	if len(lines) > 0 {
		body = "fetch(" + urlExpr + ", {\n" + strings.Join(lines, ",\n") + "\n});\n"
	}

	sections := []string{strings.Join(imports, "\n")}
	if len(setup) > 0 {
		sections = append(sections, strings.Join(setup, "\n"))
	}
	sections = append(sections, strings.TrimRight(body, "\n"))
	return strings.Join(sections, "\n\n") + "\n"
}

func requestNeedsFileFromSync(r *request.Request) bool {
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

func renderProxy(proxy string) (string, string) {
	lower := strings.ToLower(proxy)
	if strings.HasPrefix(lower, "socks") {
		return "import { SocksProxyAgent } from 'socks-proxy-agent';", "agent: new SocksProxyAgent(" + renderJSExpr(proxy) + ")"
	}
	return "import HttpsProxyAgent from 'https-proxy-agent';", "agent: new HttpsProxyAgent(" + renderJSExpr(proxy) + ")"
}

func renderFormData(parts []request.FormPart) string {
	lines := []string{"const form = new FormData();"}
	for _, part := range parts {
		if part.IsFile {
			lines = append(lines, "form.append("+reprString(part.Name)+", fileFromSync("+renderJSExpr(part.FileName)+"));")
		} else {
			lines = append(lines, "form.append("+reprString(part.Name)+", "+renderJSExpr(part.Value)+");")
		}
	}
	return strings.Join(lines, "\n")
}

func renderBody(r *request.Request) []string {
	if r.JSONBody {
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			return []string{
				"  // body: " + reprString(r.Body),
				"  body: JSON.stringify(" + jsJSONValue(parsed, 1) + ")",
			}
		}
	}
	if shouldUseURLSearchParams(r) {
		return []string{renderURLSearchParams(r.Body)}
	}
	return []string{"  body: " + renderJSExpr(r.Body)}
}

func shouldUseURLSearchParams(r *request.Request) bool {
	if r.Body == "" {
		return false
	}
	if r.AutoContentType {
		return false
	}
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(h.Value, ";")[0]))
			if mediaType != "application/x-www-form-urlencoded" {
				return false
			}
			parts := strings.Split(r.Body, "&")
			for _, part := range parts {
				key, value, ok := strings.Cut(part, "=")
				if !ok || key == "" || strings.ContainsAny(key, " {}[]\"'") || strings.ContainsAny(value, " {}[]\"'") {
					return false
				}
			}
			return true
		}
	}
	return false
}

func renderURLSearchParams(raw string) string {
	parts := strings.Split(raw, "&")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		key, value, _ := strings.Cut(part, "=")
		lines = append(lines, "    "+reprString(key)+": "+renderJSExpr(value))
	}
	return "  body: new URLSearchParams({\n" + strings.Join(lines, ",\n") + "\n  })"
}

func hasHeader(headers []request.Header, key string) bool {
	for _, h := range headers {
		if strings.EqualFold(h.Key, key) {
			return true
		}
	}
	return false
}

func renderJSExpr(s string) string {
	parts := splitEnvInterpolations(s)
	if len(parts) == 1 {
		return reprString(parts[0])
	}
	rendered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "$") && len(part) > 1 {
			rendered = append(rendered, "process.env["+reprString(part[1:])+"]")
		} else {
			rendered = append(rendered, reprString(part))
		}
	}
	return strings.Join(rendered, " + ")
}

func splitEnvInterpolations(s string) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '$' && i+1 < len(s) && isEnvStart(s[i+1]) {
			if cur.Len() > 0 {
				parts = append(parts, cur.String())
				cur.Reset()
			}
			j := i + 2
			for j < len(s) && isEnvPart(s[j]) {
				j++
			}
			parts = append(parts, s[i:j])
			i = j - 1
			continue
		}
		cur.WriteByte(s[i])
	}
	if len(parts) == 0 {
		return []string{s}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

func isEnvStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isEnvPart(ch byte) bool {
	return isEnvStart(ch) || (ch >= '0' && ch <= '9')
}

func reprString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return "'" + s + "'"
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
			lines = append(lines, nextIndent+reprString(key)+": "+jsJSONValue(value[key], indentLevel+1))
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
		return reprString(value)
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
		return reprString(fmt.Sprintf("%v", value))
	}
}
