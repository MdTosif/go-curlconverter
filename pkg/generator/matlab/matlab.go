package matlab

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var code string
	url := r.URLs[0]

	code += "%% HTTP Interface\n"
	code += "import matlab.net.*\n"
	code += "import matlab.net.http.*\n"

	if r.Body != "" || len(r.FormParts) > 0 {
		code += "import matlab.net.http.io.*\n"
	}
	code += "\n"

	// Prepare query string
	if len(url.QueryList) > 0 {
		code += "params = {"
		for _, q := range url.QueryList {
			code += reprStr(q[0]) + " " + reprStr(q[1]) + " "
		}
		code = strings.TrimSuffix(code, " ")
		code += "};\n"
	}

	// Prepare cookies
	if len(r.Cookies) > 0 {
		code += "cookies = {"
		for _, c := range r.Cookies {
			code += reprStr(c[0]) + " " + reprStr(c[1]) + " "
		}
		code = strings.TrimSuffix(code, " ")
		code += "};\n"
	}

	// Prepare headers
	if len(r.HeaderKV) > 0 {
		code += "header = [\n"
		for _, h := range r.HeaderKV {
			if h.Value != "" {
				code += "    HeaderField(" + reprStr(h.Key) + ", " + reprStr(h.Value) + ")\n"
			}
		}
		code += "]';\n"
	}

	// Prepare URI
	if len(url.QueryList) > 0 {
		code += "uri = URI(" + reprStr(url.URLWithoutQueryList) + ", QueryParameter(params'));\n"
	} else {
		code += "uri = URI(" + reprStr(url.URL) + ");\n"
	}

	// Prepare auth
	if r.BasicAuth != "" {
		parts := strings.SplitN(r.BasicAuth, ":", 2)
		if len(parts) == 2 {
			code += "cred = Credentials('Username', " + reprStr(parts[0]) + ", 'Password', " + reprStr(parts[1]) + ");\n"
		}
	}

	// Prepare options
	if r.BasicAuth != "" || r.Insecure {
		code += "options = HTTPOptions("
		options := []string{}
		if r.BasicAuth != "" {
			options = append(options, "'Credentials', cred")
		}
		if r.Insecure {
			options = append(options, "'VerifyServerName', false")
		}
		code += strings.Join(options, ", ")
		code += ");\n"
	}

	// Prepare body
	if r.Body != "" {
		contentType := getContentType(r.HeaderKV)
		if strings.HasPrefix(contentType, "application/json") {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(r.Body), &jsonData); err == nil {
				code += "body = JSONProvider(" + structify(jsonData, 1) + ");\n"
			} else {
				code += "body = StringProvider(" + reprStr(r.Body) + ");\n"
			}
		} else {
			code += "body = StringProvider(" + reprStr(r.Body) + ");\n"
		}
	}

	// Prepare request message
	method := strings.ToLower(url.Method)
	reqMessage := []string{reprStr(method)}
	if len(r.HeaderKV) > 0 {
		reqMessage = append(reqMessage, "header")
	}
	if r.Body != "" {
		if len(reqMessage) == 1 {
			reqMessage = append(reqMessage, "[]")
		}
		reqMessage = append(reqMessage, "body")
	}

	sendParams := []string{"uri.EncodedURI"}
	if r.BasicAuth != "" || r.Insecure {
		sendParams = append(sendParams, "options")
	}

	code += "\n"
	code += "response = RequestMessage([" + strings.Join(reqMessage, ", ") + "]).send(" + strings.Join(sendParams, ", ") + ");\n"

	return code
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`\p{C}|[^ \P{Z}]`)
	escaped := strings.ReplaceAll(s, "'", "''")

	parts := []string{}
	for _, c := range escaped {
		if regexEscape.MatchString(string(c)) {
			switch c {
			case '\x07':
				parts = append(parts, "sprintf('\\a')")
			case '\b':
				parts = append(parts, "sprintf('\\b')")
			case '\f':
				parts = append(parts, "sprintf('\\f')")
			case '\n':
				parts = append(parts, "newline")
			case '\r':
				parts = append(parts, "sprintf('\\r')")
			case '\t':
				parts = append(parts, "sprintf('\\t')")
			case '\v':
				parts = append(parts, "sprintf('\\v')")
			default:
				parts = append(parts, fmt.Sprintf("char(%d)", c))
			}
		} else {
			parts = append(parts, "'"+string(c)+"'")
		}
	}

	if len(parts) == 0 {
		return "''"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func structify(obj interface{}, indent int) string {
	indentStr := strings.Repeat(" ", indent*4)
	prevIndentStr := strings.Repeat(" ", (indent-1)*4)

	switch v := obj.(type) {
	case string:
		return reprStr(v)
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "string(nan)"
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		allNumbers := true
		for _, item := range v {
			if _, ok := item.(float64); !ok {
				allNumbers = false
				break
			}
		}
		if allNumbers {
			numStrs := make([]string, len(v))
			for i, item := range v {
				numStrs[i] = fmt.Sprintf("%v", item)
			}
			return "[" + strings.Join(numStrs, " ") + "]"
		}
		code := "{{\n"
		for _, item := range v {
			code += indentStr + structify(item, indent+1) + "\n"
		}
		code += prevIndentStr + "}}"
		return code
	case map[string]interface{}:
		if len(v) == 0 {
			return "struct()"
		}
		code := "struct(...\n"
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				code += ",...\n"
			}
			code += indentStr + "'" + k + "', " + structify(v[k], indent+1) + "\n"
		}
		code += prevIndentStr + ")"
		return code
	default:
		return "string(nan)"
	}
}

func getContentType(headers []request.Header) string {
	for _, h := range headers {
		if strings.ToLower(h.Key) == "content-type" {
			return h.Value
		}
	}
	return ""
}
