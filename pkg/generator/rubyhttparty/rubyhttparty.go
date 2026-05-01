package rubyhttparty

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var code string
	var partyCode string
	var imports []string

	methods := map[string]bool{
		"GET": true, "HEAD": true, "POST": true, "PATCH": true,
		"PUT": true, "PROPPATCH": true, "LOCK": true, "UNLOCK": true,
		"OPTIONS": true, "PROPFIND": true, "DELETE": true,
		"MOVE": true, "COPY": true, "MKCOL": true, "TRACE": true,
	}

	url := r.URLs[0].URL
	code += "url = " + reprStr(url) + "\n"

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	methodUpper := strings.ToUpper(method)

	if methods[methodUpper] {
		partyCode += "res = HTTParty." + strings.ToLower(method) + "(url"
	} else {
		partyCode += "res = HTTParty.get(url"
	}

	if r.BasicAuth != "" {
		authType := "basic"
		if r.DigestAuth {
			authType = "digest"
		}
		user, pass := splitBasicAuth(r.BasicAuth)
		code += fmt.Sprintf("auth = { username: %s, password: %s }\n", reprStr(user), reprStr(pass))
		if authType == "basic" {
			partyCode += ", basic_auth: auth"
		} else {
			partyCode += ", digest_auth: auth"
		}
	}

	var reqBody string
	explicitMultipart := false

	if r.BodyFile != "" {
		if r.BodyFile == "-" || r.BodyFile == "." {
			reqBody = "body = STDIN.read\n"
		} else {
			reqBody = "body = File.read(" + reprStr(r.BodyFile) + ")\n"
		}
	} else if r.Body != "" {
		var importJson bool
		reqBody, importJson = getDataString(r)
		if importJson {
			imports = append(imports, "json")
		}
	} else if len(r.FormParts) > 0 {
		reqBody, explicitMultipart = getFilesString(r)
	}

	if len(r.HeaderKV) > 0 {
		partyCode += ", headers: headers"
		code += "headers = {\n"
		for _, h := range r.HeaderKV {
			lowerKey := strings.ToLower(h.Key)
			if lowerKey == "accept-encoding" || lowerKey == "content-length" {
				code += "# "
			}
			code += fmt.Sprintf("  %s: %s,\n", reprStr(h.Key), reprStr(h.Value))
		}
		code += "}\n"
	}

	if reqBody != "" {
		code += reqBody
		if explicitMultipart {
			partyCode += ", multipart: true"
		}
		partyCode += ", body: body"
	}

	if r.Insecure {
		partyCode += ", verify: false"
	}

	partyCode += ")\n"
	code += partyCode

	return code
}

func getDataString(r *request.Request) (string, bool) {
	if r.Body == "" {
		return "", false
	}

	// Check if JSON
	contentType := ""
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			contentType = strings.TrimSpace(strings.Split(h.Value, ";")[0])
			break
		}
	}

	if contentType == "application/json" {
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			if obj, ok := parsed.(map[string]any); ok {
				return "body = " + objToRuby(obj) + ".to_json\n", true
			}
		}
	}

	// Try to parse as form data
	values, err := url.ParseQuery(r.Body)
	if err == nil && len(values) > 0 && (r.IsDataBinary == nil || !*r.IsDataBinary) {
		return "body = " + queryToRubyDict(values) + "\n", false
	}

	return "body = " + reprStr(r.Body) + "\n", false
}

func getFilesString(r *request.Request) (string, bool) {
	if len(r.FormParts) == 0 {
		return "", false
	}

	explicitMultipart := true
	body := "body = {\n"

	for _, part := range r.FormParts {
		body += "  :" + part.Name + ": "
		if part.IsFile {
			body += "File.open(" + reprStr(part.FileName) + ")"
			explicitMultipart = false
		} else {
			body += reprStr(part.Value)
		}
		body += "\n"
	}
	body += "}\n"
	return body, explicitMultipart
}

func queryToRubyDict(values url.Values) string {
	if len(values) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := []string{"{"}
	for _, k := range keys {
		for _, v := range values[k] {
			lines = append(lines, "  "+reprStr(k)+" => "+reprStr(v)+",")
		}
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func objToRuby(obj any, indent ...int) string {
	indentLevel := 0
	if len(indent) > 0 {
		indentLevel = indent[0]
	}
	indentStr := strings.Repeat("  ", indentLevel)

	switch v := obj.(type) {
	case string:
		return reprStr(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "nil"
	case map[string]any:
		if len(v) == 0 {
			return "{}"
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines := []string{"{"}
		for _, k := range keys {
			lines = append(lines, indentStr+"  "+reprStr(k)+" => "+objToRuby(v[k], indentLevel+1)+",")
		}
		lines = append(lines, indentStr+"}")
		return strings.Join(lines, "\n")
	case []any:
		if len(v) == 0 {
			return "[]"
		}
		lines := []string{"["}
		for _, item := range v {
			lines = append(lines, indentStr+"  "+objToRuby(item, indentLevel+1)+",")
		}
		lines = append(lines, indentStr+"]")
		return strings.Join(lines, "\n")
	default:
		return reprStr(fmt.Sprintf("%v", v))
	}
}

func reprStr(s string) string {
	quote := "'"
	if strings.ContainsAny(s, "\x00-\x1F\x7F-\x9F") || (strings.Contains(s, "'") && !strings.Contains(s, "\"")) {
		quote = "\""
	}

	escaped := escapeRubyString(s, quote)
	return quote + escaped + quote
}

func escapeRubyString(s, quote string) string {
	var result strings.Builder
	for i, r := range s {
		switch r {
		case '\a':
			result.WriteString("\\a")
		case '\b':
			result.WriteString("\\b")
		case '\f':
			result.WriteString("\\f")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		case '\v':
			result.WriteString("\\v")
		case '\x1b':
			result.WriteString("\\e")
		case '\\':
			result.WriteString("\\\\")
		case '\'':
			if quote == "'" {
				result.WriteString("\\'")
			} else {
				result.WriteRune(r)
			}
		case '"':
			if quote == "\"" {
				result.WriteString("\\\"")
			} else {
				result.WriteRune(r)
			}
		case '#', '$', '@':
			if quote == "\"" || quote == "{}" {
				result.WriteString("\\")
				result.WriteRune(r)
			} else {
				result.WriteRune(r)
			}
		case '}':
			if quote == "{}" {
				result.WriteString("\\}")
			} else {
				result.WriteRune(r)
			}
		case '\x00':
			if i+1 < len(s) && !isDigit(s[i+1]) {
				result.WriteString("\\0")
			} else {
				result.WriteRune(r)
			}
		default:
			if r < 32 || r == 127 || (r >= 0x7F && r <= 0x9F) {
				hex := fmt.Sprintf("\\x%02X", r)
				result.WriteString(hex)
			} else if r > 0xFFFF {
				hex := fmt.Sprintf("\\u{%X}", r)
				result.WriteString(hex)
			} else if r > 0x7F || (quote == "\"" && strings.ContainsAny(string(r), "#{$@")) {
				hex := fmt.Sprintf("\\u%04X", r)
				result.WriteString(hex)
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
