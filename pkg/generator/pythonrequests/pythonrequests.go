package pythonrequests

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
		"requests": true,
	}

	var code string
	url := r.URLs[0]

	if len(r.HeaderKV) > 0 {
		code += "headers = {\n"
		for _, h := range r.HeaderKV {
			if h.Value != "" {
				code += "    " + reprStr(h.Key) + ": " + reprStr(h.Value) + ",\n"
			}
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += "}\n"
	}

	var dataArg string
	var jsonArg string
	contentType := getContentType(r.HeaderKV)

	if r.Body != "" {
		if strings.HasPrefix(contentType, "application/json") {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(r.Body), &jsonData); err == nil {
				code += "json_data = " + jsonAsPython(jsonData, 0) + "\n"
				jsonArg = "json_data"
			} else {
				dataArg = reprStr(r.Body)
			}
		} else {
			dataArg = reprStr(r.Body)
		}
	}

	paramsArg := ""
	if len(url.QueryList) > 0 {
		code += "params = {\n"
		for _, q := range url.QueryList {
			code += "    " + reprStr(q[0]) + ": " + reprStr(q[1]) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += "}\n"
		paramsArg = "params=params"
	}

	authArg := ""
	if r.BasicAuth != "" {
		parts := strings.SplitN(r.BasicAuth, ":", 2)
		if len(parts) == 2 {
			code += "auth = (" + reprStr(parts[0]) + ", " + reprStr(parts[1]) + ")\n"
			authArg = "auth=auth"
		}
	}

	code += "response = requests." + strings.ToLower(url.Method) + "(" + reprStr(url.URL)
	args := []string{}
	if paramsArg != "" {
		args = append(args, paramsArg)
	}
	if dataArg != "" {
		args = append(args, "data="+dataArg)
	}
	if jsonArg != "" {
		args = append(args, "json="+jsonArg)
	}
	if authArg != "" {
		args = append(args, authArg)
	}
	if len(r.HeaderKV) > 0 {
		args = append(args, "headers=headers")
	}
	if r.MaxTime != "" {
		args = append(args, "timeout="+r.MaxTime)
	}
	if !r.FollowRedirects {
		args = append(args, "allow_redirects=False")
	}
	if r.Insecure {
		args = append(args, "verify=False")
	}

	if len(args) > 0 {
		code += ",\n    " + strings.Join(args, ",\n    ")
	}
	code += ")\n"

	var importCode string
	importCode = "import requests\n"
	if len(imports) > 1 {
		importList := make([]string, 0, len(imports))
		for imp := range imports {
			if imp != "requests" {
				importList = append(importList, "import "+imp)
			}
		}
		sort.Strings(importList)
		if len(importList) > 0 {
			importCode += strings.Join(importList, "\n") + "\n"
		}
	}
	importCode += "\n"

	return importCode + code
}

func reprStr(s string) string {
	regexSingleEscape := regexp.MustCompile(`'|\\|\p{C}|[^ \P{Z}]`)
	regexDoubleEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)

	quote := "'"
	if strings.Contains(s, "'") && !strings.Contains(s, "\"") {
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
		case "\\":
			return "\\\\"
		case "'":
			return "\\'"
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
	return quote + escaped + quote
}

func jsonAsPython(obj interface{}, indent int) string {
	switch v := obj.(type) {
	case string:
		return reprStr(v)
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "True"
		}
		return "False"
	case nil:
		return "None"
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		code := "[\n"
		for _, item := range v {
			code += strings.Repeat(" ", indent+4) + jsonAsPython(item, indent+4) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat(" ", indent) + "]"
		return code
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		code := "{\n"
		for k, val := range v {
			code += strings.Repeat(" ", indent+4) + reprStr(k) + ": " + jsonAsPython(val, indent+4) + ",\n"
		}
		if strings.HasSuffix(code, ",\n") {
			code = code[:len(code)-2] + "\n"
		}
		code += strings.Repeat(" ", indent) + "}"
		return code
	default:
		return "None"
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
