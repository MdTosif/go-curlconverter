package httpie

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var commands []string
	for _, url := range r.URLs {
		commands = append(commands, requestToHttpie(r, url))
	}
	return strings.Join(commands, "\n\n")
}

func requestToHttpie(r *request.Request, url request.RequestURL) string {
	flags := []string{}
	var method string
	urlArg := url.URL
	items := []string{}

	if r.BodyFile != "" || r.Body != "" || len(r.FormParts) > 0 {
		if strings.ToUpper(url.Method) != "POST" {
			method = reprStr(url.Method)
		}
	} else if strings.ToUpper(url.Method) != "GET" {
		method = reprStr(url.Method)
	}

	for _, h := range r.HeaderKV {
		if h.Value == "" {
			items = append(items, reprStr(escapeHeader(h.Key))+":")
		} else {
			items = append(items, reprStr(escapeHeader(h.Key))+":"+reprStr(escapeHeaderValue(h.Value)))
		}
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		if r.DigestAuth {
			flags = append(flags, "-A digest")
		}
		flags = append(flags, "-a "+reprStr(user+":"+pass))
	}

	if len(url.QueryList) > 0 {
		urlArg = url.URLWithoutQueryList
		for _, q := range url.QueryList {
			items = append(items, reprStr(escapeQueryName(q[0]))+"=="+reprStr(escapeQueryValue(q[1])))
		}
	}

	if r.BodyFile != "" {
		items = append(items, "@"+reprStr(r.BodyFile))
	} else if len(r.FormParts) > 0 {
		flags = append(flags, "--multipart")
		for _, part := range r.FormParts {
			if part.IsFile {
				items = append(items, reprStr(escapeFormName(part.Name))+"@"+reprStr(part.FileName))
			} else {
				items = append(items, reprStr(escapeFormName(part.Name))+"="+reprStr(part.Value))
			}
		}
	} else if r.Body != "" {
		formatData(&flags, &items, r.Body, r)
	}

	if r.FollowRedirects {
		flags = append(flags, "--follow")
	}

	if r.Insecure {
		flags = append(flags, "--verify=no")
	}
	if r.CACert != "" {
		flags = append(flags, "--verify="+reprStr(r.CACert))
	}
	if r.CAPath != "" {
		flags = append(flags, "--verify="+reprStr(r.CAPath))
	}

	if r.Cert != "" {
		flags = append(flags, "--cert="+reprStr(r.Cert))
	}
	if r.Key != "" {
		flags = append(flags, "--cert-key="+reprStr(r.Key))
	}

	if r.ConnectTimeout != "" {
		flags = append(flags, "--timeout="+reprStr(r.ConnectTimeout))
	}

	if r.Verbose {
		flags = append(flags, "--verbose")
	}
	if r.Silent {
		flags = append(flags, "--quiet")
	}

	command := "http"
	if strings.HasPrefix(urlArg, "https://") {
		command = "https"
		urlArg = localhostShorthand(urlArg[len("https://"):])
	} else if strings.HasPrefix(urlArg, "http://") {
		urlArg = localhostShorthand(urlArg[len("http://"):])
	}

	if method != "" {
		flags = append(flags, method)
	}

	needsDash := false
	for _, item := range items {
		if strings.HasPrefix(item, "-") {
			needsDash = true
			break
		}
	}
	if needsDash {
		items = append([]string{"--"}, items...)
	}

	args := append(flags, reprStr(urlArg))
	args = append(args, items...)

	multiline := len(args) > 3 || totalLength(args) > 75
	joiner := " "
	if multiline {
		joiner = " \\\n  "
	}

	return command + " " + strings.Join(args, joiner) + "\n"
}

func escapeHeader(s string) string {
	return strings.ReplaceAll(s, "=", "\\=")
}

func escapeHeaderValue(s string) string {
	if strings.HasPrefix(s, "=") || strings.HasPrefix(s, "@") {
		return "\\" + s
	}
	return s
}

func escapeJsonName(name string, isFirstKey bool) string {
	name = strings.ReplaceAll(name, "\\", "\\\\")
	name = strings.ReplaceAll(name, "[", "\\[")
	name = strings.ReplaceAll(name, "]", "\\]")
	name = strings.ReplaceAll(name, ":", "\\:")
	name = strings.ReplaceAll(name, "=", "\\=")

	if !isFirstKey {
		matched, _ := regexp.MatchString(`^\d+$`, name)
		if matched {
			name = "\\" + name
		}
	}
	return name
}

func escapeJsonStr(value string) string {
	if strings.HasPrefix(value, "\\=") {
		return value
	}
	if strings.HasPrefix(value, "\\@") {
		return value
	}
	if strings.HasPrefix(value, "=") || strings.HasPrefix(value, "@") {
		return "\\" + value
	}
	return value
}

func toJson(obj any, key string) []string {
	if obj == nil {
		return []string{reprStr(key) + ":=null"}
	}
	switch v := obj.(type) {
	case bool:
		return []string{reprStr(key) + ":=" + strconv.FormatBool(v)}
	case float64:
		return []string{reprStr(key) + ":=" + reprStr(strconv.FormatFloat(v, 'f', -1, 64))}
	case string:
		return []string{reprStr(key) + "=" + reprStr(escapeJsonStr(v))}
	case []any:
		if len(v) == 0 {
			return []string{reprStr(key) + ":=[]"}
		}
		result := []string{}
		for _, item := range v {
			result = append(result, toJson(item, key+"[]")...)
		}
		return result
	case map[string]any:
		if len(v) == 0 {
			return []string{reprStr(key) + ":={}"}
		}
		result := []string{}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			newKey := name
			if key != "" {
				newKey = key + "[" + escapeJsonName(name, false) + "]"
			} else {
				newKey = escapeJsonName(name, true)
			}
			result = append(result, toJson(v[name], newKey)...)
		}
		return result
	default:
		return []string{reprStr(key) + "=" + reprStr(fmt.Sprintf("%v", v))}
	}
}

func jsonAsHttpie(flags *[]string, items *[]string, data string) {
	var parsed any
	if err := json.Unmarshal([]byte(data), &parsed); err == nil {
		if obj, ok := parsed.(map[string]any); ok && len(obj) > 0 {
			jsonItems := toJson(obj, "")
			*items = append(*items, jsonItems...)
			return
		}
		if arr, ok := parsed.([]any); ok && len(arr) > 0 {
			jsonItems := toJson(arr, "")
			*items = append(*items, jsonItems...)
			return
		}
	}
	*flags = append(*flags, "--raw "+reprStr(data))
}

func escapeQueryName(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ":", "\\:")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

func escapeQueryValue(s string) string {
	if strings.HasPrefix(s, "=") || strings.HasPrefix(s, "@") {
		return "\\" + s
	}
	return s
}

func urlencodedAsHttpie(flags *[]string, items *[]string, data string) {
	values, err := url.ParseQuery(data)
	if err != nil || len(values) == 0 {
		*flags = append(*flags, "--raw "+reprStr(data))
		return
	}

	*flags = append(*flags, "--form")
	for k, vs := range values {
		for _, v := range vs {
			*items = append(*items, reprStr(escapeQueryName(k))+"=="+reprStr(escapeQueryValue(v)))
		}
	}
}

func formatData(flags *[]string, items *[]string, data string, r *request.Request) {
	contentType := getContentType(r)
	if contentType == "application/json" {
		jsonAsHttpie(flags, items, data)
	} else if contentType == "application/x-www-form-urlencoded" {
		urlencodedAsHttpie(flags, items, data)
	} else {
		*flags = append(*flags, "--raw "+reprStr(data))
	}
}

func escapeFormName(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "@", "\\@")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

func localhostShorthand(s string) string {
	if strings.HasPrefix(s, "localhost:") {
		return s[len("localhost"):]
	}
	if strings.HasPrefix(s, "localhost/") || s == "localhost" {
		return ":" + s[len("localhost"):]
	}
	return s
}

func reprStr(s string) string {
	// Simple shell escaping for HTTPie
	if s == "" {
		return "''"
	}
	if strings.ContainsAny(s, " \t\n\r'\"\\$|&;<>()[]{}*?") {
		return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	}
	return s
}

func totalLength(args []string) int {
	total := 0
	for _, arg := range args {
		total += len(arg)
	}
	return total
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
