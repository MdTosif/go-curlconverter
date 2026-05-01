package rhttr2

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var steps []string

	url := r.URLs[0].URL
	if len(r.URLs[0].QueryDict) > 0 {
		url = r.URLs[0].URLWithoutQueryList
	}

	steps = append(steps, "request("+reprStr(url)+")")

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}
	if method != "GET" {
		steps = addCurlStep(steps, "req_method", []string{reprStr(method)}, nil)
	}

	queryList := getQueryList(r)
	steps = addCurlStep(steps, "req_url_query", nil, queryList)

	headerList := getHeaderList(r)
	steps = addCurlStep(steps, "req_headers", nil, headerList)

	steps = addBodyStep(steps, r)

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		steps = addCurlStep(steps, "req_auth_basic", []string{reprStr(user), reprStr(pass)}, nil)
	}
	if r.BearerToken != "" {
		steps = addCurlStep(steps, "req_auth_bearer_token", []string{reprStr(r.BearerToken)}, nil)
	}

	if r.Proxy != "" {
		steps = addCurlStep(steps, "req_proxy", []string{reprStr(r.Proxy)}, nil)
	}

	if r.MaxTime != "" {
		steps = addCurlStep(steps, "req_timeout", []string{toNumeric(r.MaxTime)}, nil)
	}

	curlOptions := getCurlOptions(r)
	steps = addCurlStep(steps, "req_options", nil, curlOptions)

	retryOptions := getRetryOptions(r)
	steps = addCurlStep(steps, "req_retry", nil, retryOptions)

	performArgs := getPerformArgs(r)
	steps = addCurlStep(steps, "req_perform", nil, performArgs, true)

	code := "library(httr2)\n\n"
	code += strings.Join(steps, " |>\n  ")

	return code + "\n"
}

func getQueryList(r *request.Request) [][2]string {
	if len(r.URLs) == 0 {
		return nil
	}
	return r.URLs[0].QueryDict
}

func getHeaderList(r *request.Request) [][2]string {
	if len(r.HeaderKV) == 0 {
		return nil
	}
	result := make([][2]string, 0, len(r.HeaderKV))
	for _, h := range r.HeaderKV {
		result = append(result, [2]string{h.Key, h.Value})
	}
	return result
}

func addBodyStep(steps []string, r *request.Request) []string {
	if len(r.FormParts) > 0 {
		params := getMultipartParams(r)
		return addCurlStep(steps, "req_body_multipart", nil, params)
	}

	if r.Body == "" {
		return steps
	}

	if strings.HasPrefix(r.Body, "@") && (r.IsDataRaw == nil || !*r.IsDataRaw) {
		filePath := r.Body[1:]
		return addCurlStep(steps, "req_body_file", []string{reprStr(filePath)}, nil)
	}

	values, err := url.ParseQuery(r.Body)
	if err == nil && len(values) > 0 {
		formData := make([][2]string, 0, len(values))
		for k, vs := range values {
			for _, v := range vs {
				formData = append(formData, [2]string{k, v})
			}
		}
		return addCurlStep(steps, "req_body_form", nil, formData)
	}

	contentType := "application/x-www-form-urlencoded"
	return addCurlStep(steps, "req_body_raw", []string{reprStr(r.Body)}, [][2]string{{"type", reprStr(contentType)}})
}

func getMultipartParams(r *request.Request) [][2]string {
	params := make([][2]string, 0, len(r.FormParts))
	for _, part := range r.FormParts {
		if part.IsFile {
			params = append(params, [2]string{part.Name, "curl::form_file(" + reprStr(part.FileName) + ")"})
		} else {
			params = append(params, [2]string{part.Name, reprStr(part.Value)})
		}
	}
	return params
}

func getCurlOptions(r *request.Request) [][2]string {
	options := make([][2]string, 0)
	if r.Insecure {
		options = append(options, [2]string{"ssl_verifypeer", "0"})
	}
	if r.MaxRedirects != "" {
		options = append(options, [2]string{"maxredirs", toNumeric(r.MaxRedirects)})
	}
	if r.ConnectTimeout != "" {
		options = append(options, [2]string{"connecttimeout", toNumeric(r.ConnectTimeout)})
	}
	return options
}

func getRetryOptions(r *request.Request) [][2]string {
	options := make([][2]string, 0)
	if r.Retry != "" {
		options = append(options, [2]string{"max_tries", toNumeric(r.Retry)})
	}
	if r.RetryMaxTime != "" {
		options = append(options, [2]string{"max_seconds", toNumeric(r.RetryMaxTime)})
	}
	return options
}

func getPerformArgs(r *request.Request) [][2]string {
	args := make([][2]string, 0)
	if r.Verbose {
		args = append(args, [2]string{"verbosity", "1"})
	}
	return args
}

func addCurlStep(steps []string, f string, mainArgs []string, dots [][2]string, keepIfEmpty ...bool) []string {
	if len(mainArgs) == 0 && len(dots) == 0 && len(keepIfEmpty) == 0 {
		return steps
	}

	dotArgs := make([]string, 0, len(dots))
	for _, dot := range dots {
		name := dot[0]
		value := dot[1]
		if name == "" {
			dotArgs = append(dotArgs, value)
		} else {
			dotArgs = append(dotArgs, reprBacktick(name)+" = "+value)
		}
	}

	args := append(mainArgs, dotArgs...)
	if len(args) == 0 && len(keepIfEmpty) == 0 {
		return steps
	}

	var newStep string
	if len(dots) == 0 || len(args) == 1 {
		newStep = f + "(" + strings.Join(args, ", ") + ")"
	} else {
		indent := "    "
		argsStr := strings.Join(args, ",\n"+indent)
		newStep = f + "(\n" + indent + argsStr + "\n  )"
	}

	return append(steps, newStep)
}

func reprBacktick(s string) string {
	// Check if it can be used without backticks
	if matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9._]*$`, s); matched {
		if !reservedWords[s] {
			return s
		}
	}
	if matched, _ := regexp.MatchString(`^\.[a-zA-Z][a-zA-Z0-9._]*$`, s); matched {
		if !reservedWords[s] {
			return s
		}
	}
	return "`" + s + "`"
}

func reprStr(s string) string {
	quote := `"`
	if strings.Contains(s, `"`) && !strings.Contains(s, `'`) {
		quote = `'`
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, quote, `\`+quote)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return quote + s + quote
}

func toNumeric(s string) string {
	// Try to parse as float
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s
	}
	return "as.numeric(" + reprStr(s) + ")"
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

var (
	reservedWords = map[string]bool{
		"if": true, "else": true, "repeat": true, "while": true,
		"function": true, "for": true, "in": true, "next": true,
		"break": true, "TRUE": true, "FALSE": true, "NULL": true,
		"Inf": true, "NaN": true, "NA": true, "NA_integer_": true,
		"NA_real_": true, "NA_complex_": true, "NA_character_": true,
	}
)
