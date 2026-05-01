package phpguzzle

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

	imports := []string{"GuzzleHttp\\Client"}

	requestURL := r.URLs[0].URL
	if len(r.URLs[0].QueryDict) > 0 {
		requestURL = r.URLs[0].URLWithoutQueryList
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	var guzzleCode string
	guzzleCode = "$client = new Client();\n\n"
	guzzleCode += "$response = $client->"

	methods := map[string]bool{"GET": true, "DELETE": true, "HEAD": true, "OPTIONS": true, "PATCH": true, "POST": true, "PUT": true}
	if methods[strings.ToUpper(method)] {
		guzzleCode += strings.ToLower(method) + "("
	} else {
		guzzleCode += "request(" + reprStr(method) + ", "
	}

	var options string

	if len(r.URLs[0].QueryDict) > 0 {
		options += "    'query' => [\n"
		for _, q := range r.URLs[0].QueryDict {
			options += "        " + reprStr(q[0]) + " => " + reprStr(q[1]) + ",\n"
		}
		options = removeTrailingComma(options)
		options += "    ],\n"
	}

	if len(r.HeaderKV) > 0 {
		headerReprs := make([][2]string, 0, len(r.HeaderKV))
		for _, h := range r.HeaderKV {
			headerReprs = append(headerReprs, [2]string{reprStr(h.Key), reprStr(h.Value)})
		}

		if len(headerReprs) > 0 {
			longestHeader := 0
			for _, h := range headerReprs {
				if len(h[0]) > longestHeader {
					longestHeader = len(h[0])
				}
			}
			options += "    'headers' => [\n"
			for _, h := range headerReprs {
				options += fmt.Sprintf("        %s => %s,\n", padRight(h[0], longestHeader), h[1])
			}
			options = removeTrailingComma(options)
			options += "    ],\n"
		}
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		options += fmt.Sprintf("    'auth' => [%s, %s", reprStr(user), reprStr(pass))
		authType := "basic"
		if r.DigestAuth {
			authType = "digest"
		}
		if authType == "digest" {
			options += ", 'digest'"
		}
		options += "],\n"
	}

	if r.BodyFile != "" {
		options += fmt.Sprintf("    'body' => Psr7\\Utils::tryFopen(%s, 'r')\n", reprStr(r.BodyFile))
		imports = append(imports, "GuzzleHttp\\Psr7")
	} else if len(r.FormParts) > 0 {
		options += "    'multipart' => [\n"
		for _, part := range r.FormParts {
			options += "        [\n"
			options += fmt.Sprintf("            'name' => %s,\n", reprStr(part.Name))
			if part.IsFile {
				options += fmt.Sprintf("            'contents' => Psr7\\Utils::tryFopen(%s, 'r'),\n", reprStr(part.FileName))
				imports = append(imports, "GuzzleHttp\\Psr7")
			} else {
				options += fmt.Sprintf("            'contents' => %s,\n", reprStr(part.Value))
			}
			options = removeTrailingComma(options)
			options += "        ],\n"
		}
		options = removeTrailingComma(options)
		options += "    ],\n"
	} else if r.Body != "" {
		contentType := getContentType(r)
		if contentType == "application/x-www-form-urlencoded" {
			values, err := url.ParseQuery(r.Body)
			if err == nil && len(values) > 0 {
				options += "    'form_params' => [\n"
				for k, vs := range values {
					for _, v := range vs {
						options += "        " + reprStr(k) + " => " + reprStr(v) + ",\n"
					}
				}
				options = removeTrailingComma(options)
				options += "    ],\n"
			} else {
				options += fmt.Sprintf("    'body' => %s,\n", reprStr(r.Body))
			}
		} else if contentType == "application/json" {
			var parsed any
			if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
				jsonCode := jsonToPhp(parsed, 4)
				options += fmt.Sprintf("    'json' => %s,\n", jsonCode)
			} else {
				options += fmt.Sprintf("    'body' => %s,\n", reprStr(r.Body))
			}
		} else {
			options += fmt.Sprintf("    'body' => %s,\n", reprStr(r.Body))
		}
	}

	if r.NoProxy != "" && r.Proxy != "" {
		options += "    'proxy' => [\n"
		options += fmt.Sprintf("        'http' => %s,\n", reprStr(r.Proxy))
		options += fmt.Sprintf("        'https' => %s,\n", reprStr(r.Proxy))
		noproxies := strings.Split(r.NoProxy, ",")
		for i, np := range noproxies {
			noproxies[i] = strings.TrimSpace(np)
		}
		npReprs := make([]string, 0, len(noproxies))
		for _, np := range noproxies {
			npReprs = append(npReprs, reprStr(np))
		}
		options += fmt.Sprintf("        'no' => [%s]\n", strings.Join(npReprs, ", "))
		options += "    ],\n"
	} else if r.Proxy != "" {
		options += fmt.Sprintf("    'proxy' => %s,\n", reprStr(r.Proxy))
	}

	if r.MaxTime != "" {
		if f, err := strconv.ParseFloat(r.MaxTime, 64); err == nil {
			options += fmt.Sprintf("    'timeout' => %f,\n", f)
		}
	}
	if r.ConnectTimeout != "" {
		if f, err := strconv.ParseFloat(r.ConnectTimeout, 64); err == nil {
			options += fmt.Sprintf("    'connect_timeout' => %f,\n", f)
		}
	}

	if !r.FollowRedirects {
		options += "    'allow_redirects' => false,\n"
	}

	if r.Insecure {
		options += "    'verify' => false,\n"
	} else if r.CACert != "" {
		options += fmt.Sprintf("    'verify' => %s,\n", reprStr(r.CACert))
	} else if r.CAPath != "" {
		options += fmt.Sprintf("    'verify' => %s,\n", reprStr(r.CAPath))
	}

	if r.Cert != "" {
		if r.Pass != "" {
			options += fmt.Sprintf("    'cert' => [%s, %s],\n", reprStr(r.Cert), reprStr(r.Pass))
		} else {
			options += fmt.Sprintf("    'cert' => %s,\n", reprStr(r.Cert))
		}
	}
	if r.Key != "" {
		options += fmt.Sprintf("    'ssl_key' => %s,\n", reprStr(r.Key))
	}

	if r.HTTP3 {
		options += "    'http_version' => 3.0,\n"
	} else if r.HTTP2 {
		options += "    'http_version' => 2.0,\n"
	}

	options = removeTrailingComma(options)

	guzzleCode += reprStr(requestURL)
	if options != "" {
		guzzleCode += ", [\n"
		guzzleCode += options
		guzzleCode += "]"
	}
	guzzleCode += ");"

	sort.Strings(imports)
	code := "<?php\n"
	code += "require 'vendor/autoload.php';\n\n"
	for _, imp := range imports {
		code += "use " + imp + ";\n"
	}
	code += "\n" + guzzleCode + "\n"

	return code
}

func reprStr(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return "'" + s + "'"
}

func removeTrailingComma(s string) string {
	if strings.HasSuffix(s, ",\n") {
		return s[:len(s)-2] + "\n"
	}
	return s
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func jsonToPhp(obj any, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	switch v := obj.(type) {
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return reprStr(v)
	case []any:
		if len(v) == 0 {
			return "[]"
		}
		lines := []string{"[\n"}
		for _, item := range v {
			lines = append(lines, indentStr+"    "+jsonToPhp(item, indent+4)+",\n")
		}
		lines = append(lines, indentStr+"]")
		return strings.Join(lines, "")
	case map[string]any:
		if len(v) == 0 {
			return "new stdClass()"
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines := []string{"[\n"}
		for _, k := range keys {
			lines = append(lines, indentStr+"    "+reprStr(k)+" => "+jsonToPhp(v[k], indent+4)+",\n")
		}
		lines = append(lines, indentStr+"]")
		return strings.Join(lines, "")
	default:
		return reprStr(fmt.Sprintf("%v", v))
	}
}

func getContentType(r *request.Request) string {
	for _, h := range r.HeaderKV {
		if strings.EqualFold(h.Key, "Content-Type") {
			return strings.TrimSpace(strings.Split(h.Value, ";")[0])
		}
	}
	return ""
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
