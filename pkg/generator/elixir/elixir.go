package elixir

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	methods := []string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"}
	bodyMethods := []string{"PATCH", "POST", "PUT"}

	isBodyMethod := contains(bodyMethods, strings.ToUpper(method))
	body := getBody(r)
	headers := getHeadersDict(r)
	params := getQueryDict(r)
	options, optionsWithoutParams := getOptions(r, params)

	if isBodyMethod || (body == `""` && contains(methods, strings.ToUpper(method))) {
		args := []string{}
		keepArgs := false
		keepArgs = keepArgs || options != "[]"
		if keepArgs {
			args = append(args, options)
		}
		keepArgs = keepArgs || headers != "[]"
		if keepArgs {
			args = append(args, headers)
		}
		keepArgs = keepArgs || body != `""`
		if keepArgs && isBodyMethod {
			args = append(args, body)
		}
		args = append(args, reprStr(r.URLs[0].URLWithoutQueryList))

		// Reverse args
		for i, j := 0, len(args)-1; i < j; i, j = i+1, j-1 {
			args[i], args[j] = args[j], args[i]
		}

		var s string
		s = "response = HTTPoison." + strings.ToLower(method) + "!("
		if len(args) == 1 {
			s += args[0]
		} else {
			s += "\n"
			s += "  " + strings.Join(args, ",\n  ")
			s += "\n"
		}
		return s + ")\n"
	}

	return fmt.Sprintf(`request = %%HTTPoison.Request{
  method: :%s,
  url: %s,
  body: %s,
  headers: %s,
  options: %s,
  params: %s
}

response = HTTPoison.request(request)
`, strings.ToLower(method), reprStr(r.URLs[0].URLWithoutQueryList), body, headers, optionsWithoutParams, params)
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]|#\{`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		if len(c) == 0 {
			return ""
		}
		switch c[0] {
		case 0:
			return "\\0"
		case 7:
			return "\\a"
		case 8:
			return "\\b"
		case 12:
			return "\\f"
		case 10:
			return "\\n"
		case 13:
			return "\\r"
		case 9:
			return "\\t"
		case 11:
			return "\\v"
		case 27:
			return "\\e"
		case '\\':
			return "\\\\"
		case '"':
			return "\\\""
		case '#':
			return "\\" + c
		default:
			if len(c) == 1 {
				hex := fmt.Sprintf("%04X", c[0])
				return "\\u" + hex
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "\\u{" + hex + "}"
		}
	})
	return "\"" + escaped + "\""
}

func addIndent(value string) string {
	lines := strings.Split(value, "\n")
	result := make([]string, len(lines))
	for i, line := range lines {
		if line != "" {
			result[i] = "  " + line
		} else {
			result[i] = line
		}
	}
	return strings.Join(result, "\n")
}

func getCookies(r *request.Request) string {
	if len(r.Cookies) == 0 {
		return ""
	}
	cookies := make([]string, len(r.Cookies))
	for i, c := range r.Cookies {
		cookies[i] = c[0] + "=" + c[1]
	}
	return "cookie: [" + reprStr(strings.Join(cookies, "; ")) + "]"
}

func getOptions(r *request.Request, params string) (string, string) {
	hackneyOptions := []string{}

	auth := getBasicAuth(r)
	if auth != "" {
		hackneyOptions = append(hackneyOptions, auth)
	}

	if r.Insecure {
		hackneyOptions = append(hackneyOptions, ":insecure")
	}

	cookies := getCookies(r)
	if cookies != "" {
		hackneyOptions = append(hackneyOptions, cookies)
	}

	var hackneyOptionsString string
	if len(hackneyOptions) > 1 {
		hackneyOptionsString = "hackney: [\n    " + strings.Join(hackneyOptions, ",\n    ") + "\n  ]"
	} else if len(hackneyOptions) > 0 {
		hackneyOptionsString = "hackney: [" + hackneyOptions[0] + "]"
	}

	optionsWithoutParams := "[" + hackneyOptionsString + "]"
	var options string
	if params != "[]" {
		options = "[\n"
		options += "    params: " + strings.TrimSpace(addIndent(params))
		if hackneyOptionsString != "" {
			options += ",\n"
			options += "    " + strings.TrimSpace(addIndent(hackneyOptionsString))
		}
		options += "\n  ]"
	} else {
		options = optionsWithoutParams
	}
	return options, optionsWithoutParams
}

func getBasicAuth(r *request.Request) string {
	if r.BasicAuth == "" {
		return ""
	}
	user, pass := splitBasicAuth(r.BasicAuth)
	return "basic_auth: {" + reprStr(user) + ", " + reprStr(pass) + "}"
}

func getQueryDict(r *request.Request) string {
	if len(r.URLs[0].QueryDict) == 0 {
		return "[]"
	}
	var queryDict string
	queryDict = "[\n"
	lines := []string{}
	for _, q := range r.URLs[0].QueryDict {
		lines = append(lines, "    {"+reprStr(q[0])+", "+reprStr(q[1])+"}")
	}
	queryDict += strings.Join(lines, ",\n")
	queryDict += "\n  ]"
	return queryDict
}

func getHeadersDict(r *request.Request) string {
	if len(r.HeaderKV) == 0 {
		return "[]"
	}
	var dict string
	dict = "[\n"
	lines := []string{}
	for _, h := range r.HeaderKV {
		lines = append(lines, "    {"+reprStr(h.Key)+", "+reprStr(h.Value)+"}")
	}
	dict += strings.Join(lines, ",\n")
	dict += "\n  ]"
	return dict
}

func getBody(r *request.Request) string {
	formData := getFormDataString(r)
	if formData != "" {
		return formData
	}
	return `""`
}

func getFormDataString(r *request.Request) string {
	if len(r.FormParts) > 0 {
		if len(r.FormParts) == 0 {
			return "{:multipart, []}"
		}

		formParams := []string{}
		for _, part := range r.FormParts {
			if part.IsFile {
				filename := part.FileName
				if filename == "" {
					filename = part.FileName
				}
				formParams = append(formParams, fmt.Sprintf("    {:file, %s, {\"form-data\", [{:name, %s}, {:filename, Path.basename(%s)}]}, []}", reprStr(part.FileName), reprStr(part.Name), reprStr(filename)))
			} else {
				formParams = append(formParams, "    {"+reprStr(part.Name)+", "+reprStr(part.Value)+"}")
			}
		}

		formStr := strings.Join(formParams, ",\n")
		if formStr != "" {
			return "{:multipart, [\n" + formStr + "\n  ]}"
		}
	}

	if r.Body != "" {
		return getDataString(r)
	}

	return ""
}

func getDataString(r *request.Request) string {
	if r.Body == "" {
		return ""
	}

	if strings.HasPrefix(r.Body, "@") {
		filePath := r.Body[1:]
		if r.IsDataBinary != nil && *r.IsDataBinary {
			return "File.read!(" + reprStr(filePath) + ")"
		} else {
			return "{:file, " + reprStr(filePath) + "}"
		}
	}

	values, err := url.ParseQuery(r.Body)
	if err == nil && len(values) > 0 {
		data := []string{}
		for k, vs := range values {
			for _, v := range vs {
				data = append(data, "    {"+reprStr(k)+", "+reprStr(v)+"}")
			}
		}
		return "{:form, [\n" + strings.Join(data, ",\n") + "\n  ]}"
	}

	if !strings.Contains(r.Body, "|") && len(strings.SplitN(r.Body, "\n", 4)) > 3 {
		return "~s|" + r.Body + "|"
	}
	return reprStr(r.Body)
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
