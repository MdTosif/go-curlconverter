package julia

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

	imports := map[string]bool{
		"HTTP": true,
	}

	var code string
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"}
	method := r.URLs[0].Method
	fn := "HTTP.request"
	args := []string{}

	methodUpper := strings.ToUpper(method)
	isStandardMethod := false
	for _, m := range methods {
		if m == methodUpper {
			isStandardMethod = true
			fn = "HTTP." + strings.ToLower(method)
			break
		}
	}
	if !isStandardMethod {
		args = append(args, reprStr(method))
	}

	url := reprStr(r.URLs[0].URL)
	hasQuery := false
	if len(r.URLs[0].QueryList) > 0 {
		code += "query = [\n"
		for _, q := range r.URLs[0].QueryList {
			code += "    " + reprStr(q[0]) + " => " + reprStr(q[1]) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += "]\n\n"
		url = reprStr(r.URLs[0].URLWithoutQueryList)
		hasQuery = true
	}
	args = append(args, url)

	hasHeaders := len(r.HeaderKV) > 0 || r.BasicAuth != ""
	if hasHeaders {
		code += "headers = Dict(\n"
		for _, h := range r.HeaderKV {
			if h.Value != "" {
				code += "    " + reprStr(h.Key) + " => " + reprStr(h.Value) + ",\n"
			}
		}
		if r.BasicAuth != "" {
			code += "    \"Authorization\" => \"Basic \" * base64encode(" + reprStr(r.BasicAuth) + "),\n"
			imports["Base64"] = true
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += ")\n\n"
		args = append(args, "headers")
	}

	bodyArg := ""
	if r.Body != "" || len(r.FormParts) > 0 {
		if len(r.FormParts) > 0 {
			code += "form = HTTP.Form(\n"
			code += "    Dict(\n"
			for _, part := range r.FormParts {
				code += "        " + reprStr(part.Name) + " => "
				if part.IsFile {
					code += "open(" + reprStr(part.FileName) + "),\n"
				} else {
					code += reprStr(part.Value) + ",\n"
				}
			}
			if strings.HasSuffix(code, ",\n") {
				code = code[:len(code)-2] + "\n"
			}
			code += "    )\n"
			code += ")\n\n"
			bodyArg = "form"
		} else {
			contentType := getContentType(r.HeaderKV)
			if contentType == "application/json" {
				var jsonData interface{}
				if err := json.Unmarshal([]byte(r.Body), &jsonData); err == nil {
					code += "body = " + jsonAsJulia(jsonData, 0) + "\n\n"
					imports["JSON"] = true
					bodyArg = "JSON.json(body)"
				} else {
					code += "body = " + reprStr(r.Body) + "\n\n"
					bodyArg = "body"
				}
			} else if contentType == "application/x-www-form-urlencoded" {
				values := parseQueryString(r.Body)
				if len(values) > 0 {
					code += "body = Dict(\n"
					for k, v := range values {
						code += "    " + reprStr(k) + " => " + reprStr(v) + ",\n"
					}
					if strings.HasSuffix(code, ",\n") {
						code = code[:len(code)-2] + "\n"
					}
					code += ")\n\n"
					bodyArg = "body"
				} else {
					code += "body = " + reprStr(r.Body) + "\n\n"
					bodyArg = "body"
				}
			} else {
				code += "body = " + reprStr(r.Body) + "\n\n"
				bodyArg = "body"
			}
		}
	}

	if bodyArg != "" {
		if !hasHeaders {
			args = append(args, "[]")
		}
		args = append(args, bodyArg)
	}

	if hasQuery {
		args = append(args, "query=query")
	}

	if r.MaxTime != "" {
		args = append(args, "readtimeout="+r.MaxTime)
	}

	if !r.Compressed {
		args = append(args, "decompress=false")
	}

	if !r.FollowRedirects {
		args = append(args, "redirect=false")
	}

	if r.Insecure {
		args = append(args, "require_ssl_verification=false")
	}

	code += "resp = " + fn + "("
	oneLineArgs := strings.Join(args, ", ")
	if ((fn == "HTTP.request" && len(args) > 2) || (fn != "HTTP.request" && len(args) > 1)) && len(oneLineArgs) > 70 {
		code += "\n    " + strings.Join(args, ",\n    ")
		code += "\n)\n"
	} else {
		code += oneLineArgs + ")\n"
	}

	var importCode string
	if len(imports) > 0 {
		importList := make([]string, 0, len(imports))
		for imp := range imports {
			importList = append(importList, imp)
		}
		sort.Strings(importList)
		importCode = "using " + strings.Join(importList, ", ") + "\n\n"
	}

	return importCode + code
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`\$|\\|"|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "$":
			return "\\$"
		case "\\":
			return "\\\\"
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
		case "\"":
			return "\\\""
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
		return "\\U" + hex
	})
	return "\"" + escaped + "\""
}

func jsonAsJulia(obj interface{}, indent int) string {
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
		return "nothing"
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		code := "[\n"
		for _, item := range v {
			code += strings.Repeat(" ", indent+4) + jsonAsJulia(item, indent+4) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat(" ", indent) + "]"
		return code
	case map[string]interface{}:
		if len(v) == 0 {
			return "Dict()"
		}
		code := "Dict(\n"
		for k, val := range v {
			code += strings.Repeat(" ", indent+4) + reprStr(k) + " => " + jsonAsJulia(val, indent+4) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat(" ", indent) + ")"
		return code
	default:
		return "nothing"
	}
}

func parseQueryString(s string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(s, "&")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func getContentType(headers []request.Header) string {
	for _, h := range headers {
		if strings.ToLower(h.Key) == "content-type" {
			return h.Value
		}
	}
	return ""
}
