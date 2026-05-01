package ruby

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

	imports := map[string]bool{}
	var code string
	url := r.URLs[0]

	code += "require 'net/http'\n"

	// URI and query parameters
	if len(url.QueryList) > 0 {
		code += "uri = URI(" + reprStr(url.URLWithoutQueryList) + ")\n"
		code += "params = {\n"
		for _, q := range url.QueryList {
			code += "  " + reprStr(q[0]) + " => " + reprStr(q[1]) + ",\n"
		}
		code += "}\n"
		code += "uri.query = URI.encode_www_form(params)\n\n"
	} else {
		code += "uri = URI(" + reprStr(url.URL) + ")\n"
	}

	// Method handling
	methods := map[string]string{
		"GET":    "Get",
		"HEAD":   "Head",
		"POST":   "Post",
		"PATCH":  "Patch",
		"PUT":    "Put",
		"DELETE": "Delete",
	}

	method := strings.ToUpper(url.Method)
	if m, ok := methods[method]; ok {
		if method == "GET" && len(r.HeaderKV) == 0 && r.BasicAuth == "" && r.Body == "" && !r.Insecure {
			code += "res = Net::HTTP.get_response(uri)\n"
			return code
		}
		code += "req = Net::HTTP::" + m + ".new(uri)\n"
	} else {
		code += "req = Net::HTTPGenericRequest.new(" + reprStr(method) + ", true, true, uri)\n"
	}

	// Basic auth
	if r.BasicAuth != "" {
		parts := strings.SplitN(r.BasicAuth, ":", 2)
		if len(parts) == 2 {
			code += "req.basic_auth " + reprStr(parts[0]) + ", " + reprStr(parts[1]) + "\n"
		}
	}

	// Body
	if r.Body != "" {
		contentType := getContentType(r.HeaderKV)
		if strings.HasPrefix(contentType, "application/json") {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(r.Body), &jsonData); err == nil {
				code += "req.body = " + objToRuby(jsonData, 0) + ".to_json\n"
				imports["json"] = true
			} else {
				code += "req.body = " + reprStr(r.Body) + "\n"
			}
		} else {
			code += "req.body = " + reprStr(r.Body) + "\n"
		}
	}

	// Content type
	contentType := getContentType(r.HeaderKV)
	if contentType != "" {
		code += "req.content_type = " + reprStr(contentType) + "\n"
	}

	// Headers
	for _, h := range r.HeaderKV {
		if h.Value != "" && strings.ToLower(h.Key) != "content-type" {
			code += "req[" + reprStr(h.Key) + "] = " + reprStr(h.Value) + "\n"
		}
	}

	code += "\n"

	// Request options
	code += "req_options = {\n"
	code += "  use_ssl: uri.scheme == 'https'"
	if r.Insecure {
		imports["openssl"] = true
		code += ",\n"
		code += "  verify_mode: OpenSSL::SSL::VERIFY_NONE\n"
	} else {
		code += "\n"
	}
	code += "}\n"

	code += "res = Net::HTTP.start(uri.hostname, uri.port, req_options) do |http|\n"
	code += "  http.request(req)\n"
	code += "end\n"

	// Add imports
	var prelude string
	for imp := range imports {
		prelude += "require '" + imp + "'\n"
	}
	if prelude != "" {
		prelude += "\n"
	}

	return prelude + code
}

func reprStr(s string) string {
	regexSingleEscape := regexp.MustCompile(`'|\\`)
	regexDoubleEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]|#[{@$]`)

	quote := "'"
	if regexDoubleEscape.MatchString(s) || (strings.Contains(s, "'") && !strings.Contains(s, "\"")) {
		quote = "\""
	}

	regex := regexSingleEscape
	if quote == "\"" {
		regex = regexDoubleEscape
	}

	escaped := regex.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\x07":
			return "\\a"
		case "\b":
			return "\\b"
		case "\f":
			return "\\f"
		case "\n":
			return "\\n"
		case "\r":
			return "\\r"
		case "\t":
			return "\\t"
		case "\v":
			return "\\v"
		case "\x1B":
			return "\\e"
		case "\\":
			return "\\\\"
		case "'":
			return "\\'"
		case "\"":
			return "\\\""
		case "#":
			return "\\" + c
		}
		if len(c) == 0 {
			return ""
		}
		hex := fmt.Sprintf("%X", c[0])
		if len(hex) <= 2 {
			return "\\x" + hex
		}
		if len(hex) <= 4 {
			return "\\u" + hex
		}
		return "\\u{" + hex + "}"
	})
	return quote + escaped + quote
}

func objToRuby(obj interface{}, indent int) string {
	indentStr := strings.Repeat(" ", indent+2)
	prevIndentStr := strings.Repeat(" ", indent)

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
		return "nil"
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		code := "[\n"
		for i, item := range v {
			code += indentStr + objToRuby(item, indent+2)
			if i < len(v)-1 {
				code += ",\n"
			} else {
				code += "\n"
			}
		}
		code += prevIndentStr + "]"
		return code
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		code := "{\n"
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			code += indentStr + reprStr(k) + " => " + objToRuby(v[k], indent+2)
			if i < len(keys)-1 {
				code += ",\n"
			} else {
				code += "\n"
			}
		}
		code += prevIndentStr + "}"
		return code
	default:
		return "nil"
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
