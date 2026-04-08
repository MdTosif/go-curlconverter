package nodeaxios

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	imports := []string{"import axios from 'axios';"}
	setup := []string{}
	body := ""
	url := renderJSExpr(r.URLs[0].URLWithoutQueryList)
	if len(r.URLs[0].QueryDict) == 0 {
		url = renderJSExpr(r.URLs[0].URL)
	}

	methodRaw := r.URLs[0].Method
	if methodRaw == "" {
		methodRaw = "GET"
	}
	methodLower := strings.ToLower(methodRaw)
	knownMethods := map[string]bool{
		"get": true, "delete": true, "head": true, "options": true,
		"post": true, "put": true, "patch": true,
	}
	dataMethods := map[string]bool{"post": true, "put": true, "patch": true}

	hasMultipart := len(r.FormParts) > 0
	if hasMultipart {
		imports = append(imports, "import FormData from 'form-data';", "import * as fs from 'fs';")
		setup = append(setup, "const form = new FormData();")
		for _, part := range r.FormParts {
			if part.IsFile {
				setup = append(setup, "form.append("+reprString(part.Name)+", fs.readFileSync("+renderJSExpr(part.FileName)+"), "+reprString(path.Base(part.FileName))+");")
			} else {
				setup = append(setup, "form.append("+reprString(part.Name)+", "+renderJSExpr(part.Value)+");")
			}
		}
	}

	configLines := []string{}
	if len(r.URLs[0].QueryDict) > 0 {
		configLines = append(configLines, "  params: "+renderDict(r.URLs[0].QueryDict))
	}
	if len(r.HeaderKV) > 0 || hasMultipart {
		headerLines := []string{}
		if hasMultipart {
			headerLines = append(headerLines, "    ...form.getHeaders()")
		}
		for _, header := range r.HeaderKV {
			headerLines = append(headerLines, "    "+reprString(header.Key)+": "+renderJSExpr(header.Value))
		}
		configLines = append(configLines, "  headers: {\n"+strings.Join(headerLines, ",\n")+"\n  }")
	}
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		authLines := []string{"    username: " + renderJSExpr(user)}
		if pass != "" {
			authLines = append(authLines, "    password: "+renderJSExpr(pass))
		}
		configLines = append(configLines, "  auth: {\n"+strings.Join(authLines, ",\n")+"\n  }")
	}
	if r.Proxy != "" {
		configLines = append(configLines, renderProxyConfig(r))
	}
	if !knownMethods[methodLower] {
		configLines = append([]string{"  method: " + reprString(methodLower)}, configLines...)
	}

	callPrefix := "const response = await axios"
	if knownMethods[methodLower] {
		callPrefix += "." + methodLower
	}

	if dataMethods[methodLower] {
		dataExpr, comment := renderDataExpr(r)
		lines := []string{callPrefix + "(", "  " + url + ","}
		if comment != "" {
			lines = append(lines, "  // "+comment+",")
		}
		if dataExpr == "" {
			dataExpr = "''"
		}
		lines = append(lines, "  "+dataExpr)
		if len(configLines) > 0 {
			lines[len(lines)-1] += ","
			lines = append(lines, indentBlock(renderConfig(configLines), "  "))
		}
		lines = append(lines, ");")
		body = strings.Join(lines, "\n")
	} else {
		if len(configLines) > 0 {
			body = callPrefix + "(" + url + ", " + renderConfig(configLines) + ");"
		} else {
			body = callPrefix + "(" + url + ");"
		}
	}

	sections := []string{strings.Join(imports, "\n")}
	if len(setup) > 0 {
		sections = append(sections, strings.Join(setup, "\n"))
	}
	sections = append(sections, body)
	return strings.Join(sections, "\n\n") + "\n"
}

func renderConfig(lines []string) string {
	return "{\n" + strings.Join(lines, ",\n") + "\n}"
}

func renderDict(pairs [][2]string) string {
	lines := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		lines = append(lines, "    "+reprString(pair[0])+": "+renderJSExpr(pair[1]))
	}
	return "{\n" + strings.Join(lines, ",\n") + "\n  }"
}

func renderProxyConfig(r *request.Request) string {
	proxy := r.Proxy
	if !strings.Contains(proxy, "://") {
		proxy = "http://" + proxy
	}
	parts := strings.SplitN(proxy, "://", 2)
	protocol := strings.ToLower(parts[0])
	hostPort := parts[1]
	host := hostPort
	port := "1080"
	if idx := strings.LastIndex(hostPort, ":"); idx != -1 {
		host = hostPort[:idx]
		port = hostPort[idx+1:]
	}
	lines := []string{
		"  proxy: {",
		"    protocol: " + reprString(protocol) + ",",
		"    host: " + renderJSExpr(host) + ",",
	}
	if portInt, err := strconv.Atoi(port); err == nil {
		lines = append(lines, fmt.Sprintf("    port: %d,", portInt))
	} else {
		lines = append(lines, "    port: "+renderJSExpr(port)+",")
	}
	if r.ProxyAuth != "" {
		user, pass := splitBasicAuth(r.ProxyAuth)
		lines = append(lines,
			"    auth: {",
			"      user: "+renderJSExpr(user)+",",
			"      password: "+renderJSExpr(pass),
			"    },",
		)
	}
	if strings.HasSuffix(lines[len(lines)-1], ",") {
		lines[len(lines)-1] = strings.TrimSuffix(lines[len(lines)-1], ",")
	}
	lines = append(lines, "  }")
	return strings.Join(lines, "\n")
}

func renderDataExpr(r *request.Request) (string, string) {
	if len(r.FormParts) > 0 {
		return "form", ""
	}
	if r.JSONBody && r.Body != "" {
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			return renderJSONValue(parsed, 1), reprString(r.Body)
		}
	}
	if r.Body != "" {
		if hasExplicitFormContentType(r) {
			if dict, ok := asSimpleDict(r.Body); ok {
				return "new URLSearchParams(" + dict + ")", ""
			}
		}
		return renderJSExpr(r.Body), ""
	}
	if r.BodyFile != "" {
		return renderJSExpr("@" + r.BodyFile), ""
	}
	return "", ""
}

func asSimpleDict(raw string) (string, bool) {
	parts := strings.Split(raw, "&")
	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return "", false
		}
		lines = append(lines, "    "+reprString(key)+": "+renderJSExpr(value))
	}
	return "{\n" + strings.Join(lines, ",\n") + "\n  }", true
}

func hasExplicitFormContentType(r *request.Request) bool {
	if r.AutoContentType {
		return false
	}
	for _, header := range r.HeaderKV {
		if strings.EqualFold(header.Key, "Content-Type") {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(header.Value, ";")[0]))
			return mediaType == "application/x-www-form-urlencoded"
		}
	}
	return false
}

func renderJSONValue(v any, indentLevel int) string {
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
			lines = append(lines, nextIndent+reprString(key)+": "+renderJSONValue(value[key], indentLevel+1))
		}
		return "{\n" + strings.Join(lines, ",\n") + "\n" + indent + "}"
	case []any:
		if len(value) == 0 {
			return "[]"
		}
		lines := make([]string, 0, len(value))
		for _, item := range value {
			lines = append(lines, nextIndent+renderJSONValue(item, indentLevel+1))
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
		ch := s[i]
		if ch == '$' && i+1 < len(s) && isEnvStart(s[i+1]) {
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
		cur.WriteByte(ch)
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
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_'
}

func isEnvPart(ch byte) bool {
	return isEnvStart(ch) || (ch >= '0' && ch <= '9')
}

func reprString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return "'" + s + "'"
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func indentBlock(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
