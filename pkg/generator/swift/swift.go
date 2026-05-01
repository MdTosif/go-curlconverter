package swift

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var code string
	code += "import Foundation\n\n"

	var dataCode, dataArgCode string
	if r.Body != "" {
		dataCode, dataArgCode = formatData(r, r.Body)
		if dataCode != "" {
			code += dataCode + "\n"
		}
	}

	code += "let url = URL(string: " + reprStr(r.URLs[0].URL) + ")!\n"
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		code += "\n"
		code += "let username = " + reprStr(user) + "\n"
		code += "let password = " + reprStr(pass) + "\n"
		code += "let loginString = String(format: \"\\(username):\\(password)\", username, password)\n"
		code += "let loginData = loginString.data(using: String.Encoding.utf8)!\n"
		code += "let base64LoginString = loginData.base64EncodedString()\n"
		code += "\n"
	}

	if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
		code += "let headers = [\n"
		for _, h := range r.HeaderKV {
			if h.Value != "" {
				code += "    " + reprStr(h.Key) + ": " + reprStr(h.Value) + ",\n"
			}
		}
		if r.BasicAuth != "" {
			code += "    \"Authorization\": \"Basic \\(base64LoginString)\",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += "]\n"
	}
	code += "\n"

	code += "var request = URLRequest(url: url"
	if r.MaxTime != "" && r.MaxTime != "60.0" {
		code += ", timeoutInterval: " + r.MaxTime
	}
	code += ")\n"

	if strings.ToUpper(r.URLs[0].Method) != "GET" {
		code += "request.httpMethod = " + reprStr(r.URLs[0].Method) + "\n"
	}

	if len(r.HeaderKV) > 0 {
		code += "request.allHTTPHeaderFields = headers\n"
	}

	if dataArgCode != "" {
		code += dataArgCode
	}

	code += "\n"
	code += "let task = URLSession.shared.dataTask(with: request) { (data, response, error) in\n"
	code += "    if let error = error {\n"
	code += "        print(error)\n"
	code += "    } else if let data = data {\n"
	code += "        let str = String(data: data, encoding: .utf8)\n"
	code += "        print(str ?? \"\")\n"
	code += "    }\n"
	code += "}\n"
	code += "\n"
	code += "task.resume()\n"

	return code
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\x00":
			return "\\0"
		case "\\":
			return "\\\\"
		case "\t":
			return "\\t"
		case "\n":
			return "\\n"
		case "\r":
			return "\\r"
		case "\"":
			return "\\\""
		}
		if len(c) == 0 {
			return ""
		}
		hex := fmt.Sprintf("%X", c[0])
		return "\\u{" + hex + "}"
	})
	return "\"" + escaped + "\""
}

func reprJson(data interface{}, indent int) string {
	switch v := data.(type) {
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
		for _, item := range v {
			code += strings.Repeat("    ", indent+1) + reprJson(item, indent+1) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat("    ", indent) + "]"
		return code
	case map[string]interface{}:
		if len(v) == 0 {
			return "[:]"
		}
		code := "[\n"
		for key, value := range v {
			code += strings.Repeat("    ", indent+1) + reprStr(key) + ": " + reprJson(value, indent+1) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat("    ", indent) + "]"
		return code
	default:
		return "nil"
	}
}

func formatData(r *request.Request, data string) (string, string) {
	contentType := getContentType(r.HeaderKV)
	if contentType == "application/x-www-form-urlencoded" {
		parts := strings.Split(data, "&")
		if len(parts) > 0 {
			code := "let data = NSMutableData(data: " + reprStr(parts[0]) + ".data(using: .utf8)!)\n"
			for _, part := range parts[1:] {
				code += "data.append(" + reprStr("&"+part) + ".data(using: .utf8)!)\n"
			}
			return code, "request.httpBody = data as Data\n"
		}
	} else if contentType == "application/json" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(data), &parsed); err == nil {
			roundtrips, _ := json.Marshal(parsed)
			if string(roundtrips) == data {
				var code string
				switch parsed.(type) {
				case []interface{}:
					code += "let jsonData = " + reprJson(parsed, 0) + " as [Any]\n"
				case map[string]interface{}:
					code += "let jsonData = " + reprJson(parsed, 0) + " as [String : Any]\n"
				default:
					code += "let jsonData = " + reprJson(parsed, 0) + "\n"
				}
				code += "let data = try! JSONSerialization.data(withJSONObject: jsonData, options: [])\n"
				return code, "request.httpBody = data as Data\n"
			}
		}
	}
	return "", "request.httpBody = " + reprStr(data) + ".data(using: .utf8)\n"
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func getContentType(headers []request.Header) string {
	for _, h := range headers {
		if strings.ToLower(h.Key) == "content-type" {
			return h.Value
		}
	}
	return ""
}
