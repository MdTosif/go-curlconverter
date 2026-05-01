package pythonhttp

import (
	"encoding/base64"
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
		"http.client": true,
	}

	var code string
	urlObj := r.URLs[0].URLObj

	if urlObj.Scheme == "https" {
		code += "conn = http.client.HTTPSConnection("
	} else {
		code += "conn = http.client.HTTPConnection("
	}

	classArgs := []string{reprStr(urlObj.Host)}
	if r.MaxTime != "" {
		classArgs = append(classArgs, "timeout="+r.MaxTime)
	}
	if r.Insecure && urlObj.Scheme == "https" {
		classArgs = append(classArgs, "context=ssl._create_unverified_context()")
		imports["ssl"] = true
	}

	joinedClassArgs := strings.Join(classArgs, ", ")
	if len(joinedClassArgs) > 80 {
		joinedClassArgs = "\n    " + strings.Join(classArgs, ",\n    ") + "\n"
	}
	code += joinedClassArgs + ")\n"

	if r.BasicAuth != "" && r.AuthType == "basic" {
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(r.BasicAuth))
		r.HeaderKV = append(r.HeaderKV, request.Header{Key: "Authorization", Value: authHeader})
	}

	if len(r.HeaderKV) > 0 {
		code += formatHeaders(r.HeaderKV, imports)
	}

	var dataAsJson string
	var jsonRoundtrips bool
	contentType := getContentType(r.HeaderKV)
	isJson := r.Body != "" && strings.HasPrefix(contentType, "application/json")
	if isJson {
		dataAsJson, jsonRoundtrips = formatDataAsJson(r.Body, imports)
	}
	if dataAsJson != "" {
		code += dataAsJson
	}

	code += "conn.request("
	args := []string{reprStr(r.URLs[0].Method), reprStr(urlObj.Path + urlObj.Query)}
	if dataAsJson != "" {
		args = append(args, "json.dumps(json_data)")
		if !jsonRoundtrips && r.Body != "" {
			args = append(args, "# "+reprStr(r.Body))
		}
		imports["json"] = true
	} else if r.Body != "" {
		args = append(args, reprStr(r.Body))
	}
	if len(r.HeaderKV) > 0 {
		if len(args) == 2 {
			args = append(args, "headers=headers")
		} else {
			args = append(args, "headers")
		}
	}

	joinedArgs := strings.Join(args, ", ")
	if len(joinedArgs) > 80 {
		joinedArgs = "\n    " + strings.Join(args, ",\n    ") + "\n"
	}
	code += joinedArgs + ")\n"

	code += "response = conn.getresponse()\n"

	var importCode string
	importCode = "import http.client\n"
	importCode += printImports(imports)
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

func formatHeaders(headers []request.Header, imports map[string]bool) string {
	code := "headers = {\n"
	for _, h := range headers {
		if h.Value != "" {
			code += "    " + reprStr(h.Key) + ": " + reprStr(h.Value) + ",\n"
		}
	}
	if strings.HasSuffix(code, ",\n") {
		code = code[:len(code)-2] + "\n"
	}
	code += "}\n"
	return code
}

func formatDataAsJson(data string, imports map[string]bool) (string, bool) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return "", false
	}
	code := "json_data = " + jsonAsPython(jsonData, 0) + "\n"
	roundtrips, _ := json.Marshal(jsonData)
	return code, string(roundtrips) == data
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

func printImports(imports map[string]bool) string {
	if len(imports) == 0 {
		return ""
	}
	importList := make([]string, 0, len(imports))
	for imp := range imports {
		if imp != "" {
			importList = append(importList, "import "+imp)
		}
	}
	sort.Strings(importList)
	return strings.Join(importList, "\n") + "\n"
}

func getContentType(headers []request.Header) string {
	for _, h := range headers {
		if strings.ToLower(h.Key) == "content-type" {
			return h.Value
		}
	}
	return ""
}
