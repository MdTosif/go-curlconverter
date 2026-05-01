package r

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

var (
	reservedWords = map[string]bool{
		"if": true, "else": true, "repeat": true, "while": true,
		"function": true, "for": true, "in": true, "next": true,
		"break": true, "TRUE": true, "FALSE": true, "NULL": true,
		"Inf": true, "NaN": true, "NA": true, "NA_integer_": true,
		"NA_real_": true, "NA_complex_": true, "NA_character_": true,
	}

	backtickEscapeRegex = regexp.MustCompile("[`\\\\\\pC\\pZ]")
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var sections []string
	sections = append(sections, "library(httr)")

	cookieDict := getCookieDict(r)
	if cookieDict != "" {
		sections = append(sections, cookieDict)
	}

	headerDict := getHeaderDict(r)
	if headerDict != "" {
		sections = append(sections, headerDict)
	}

	queryList := getQueryList(r)
	if queryList != "" {
		sections = append(sections, queryList)
	}

	dataString, filesString := getDataString(r)
	if dataString != "" {
		sections = append(sections, dataString)
	} else if filesString != "" {
		sections = append(sections, filesString)
	}

	requestLine := getRequestLine(r, headerDict != "", queryList != "", cookieDict != "", dataString != "", filesString != "")
	sections = append(sections, requestLine)

	return strings.Join(sections, "\n\n") + "\n"
}

func getCookieDict(r *request.Request) string {
	if len(r.Cookies) == 0 {
		return ""
	}

	lines := []string{"cookies = c("}
	for _, cookie := range r.Cookies {
		// httr percent-encodes cookie values
		decoded, err := url.QueryUnescape(strings.ReplaceAll(cookie[1], "+", " "))
		if err != nil {
			return ""
		}
		lines = append(lines, "  "+reprBacktick(cookie[0])+" = "+reprStr(decoded)+",")
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func getHeaderDict(r *request.Request) string {
	if len(r.HeaderKV) == 0 {
		return ""
	}

	lines := []string{"headers = c("}
	for _, h := range r.HeaderKV {
		lines = append(lines, "  "+reprBacktick(h.Key)+" = "+reprStr(h.Value)+",")
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func getQueryList(r *request.Request) string {
	if len(r.URLs) == 0 || len(r.URLs[0].QueryDict) == 0 {
		return ""
	}

	lines := []string{"params = list("}
	for _, q := range r.URLs[0].QueryDict {
		lines = append(lines, "  "+reprBacktick(q[0])+" = "+reprStr(q[1])+",")
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

func getDataString(r *request.Request) (string, string) {
	if len(r.FormParts) > 0 {
		lines := []string{"files = list("}
		for _, part := range r.FormParts {
			if part.IsFile {
				lines = append(lines, "  "+reprBacktick(part.Name)+" = upload_file("+reprStr(part.FileName)+"),")
			} else {
				lines = append(lines, "  "+reprBacktick(part.Name)+" = "+reprStr(part.Value)+",")
			}
		}
		lines = append(lines, ")")
		return "", strings.Join(lines, "\n")
	}

	if r.Body != "" {
		if strings.HasPrefix(r.Body, "@") && (r.IsDataRaw == nil || !*r.IsDataRaw) {
			filePath := r.Body[1:]
			return "data = upload_file(" + reprStr(filePath) + ")", ""
		}

		// Try to parse as form data
		values, err := url.ParseQuery(r.Body)
		if err == nil && len(values) > 0 {
			lines := []string{"data = list("}
			keys := make([]string, 0, len(values))
			for k := range values {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				for _, v := range values[k] {
					lines = append(lines, "  "+reprBacktick(k)+" = "+reprStr(v)+",")
				}
			}
			lines = append(lines, ")")
			return strings.Join(lines, "\n"), ""
		}

		return "data = " + reprStr(r.Body), ""
	}

	return "", ""
}

func getRequestLine(r *request.Request, hasHeaders, hasQuery, hasCookies, hasData, hasFiles bool) string {
	url := r.URLs[0].URL
	if hasQuery {
		url = r.URLs[0].URLWithoutQueryList
	}

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	var requestLine string
	simpleMethods := map[string]bool{"GET": true, "HEAD": true, "PATCH": true, "PUT": true, "DELETE": true, "POST": true}

	if simpleMethods[method] {
		requestLine = "res <- httr::" + method + "("
	} else {
		requestLine = "res <- httr::VERB(" + reprStr(method) + ", "
	}

	requestLine += "url = " + reprStr(url)

	if hasHeaders {
		requestLine += ", httr::add_headers(.headers=headers)"
	}
	if hasQuery {
		requestLine += ", query = params"
	}
	if hasCookies {
		requestLine += ", httr::set_cookies(.cookies = cookies)"
	}
	if hasFiles {
		requestLine += `, body = files, encode = "multipart"`
	} else if hasData {
		requestLine += ", body = data"
	}
	if r.Insecure {
		requestLine += ", config = httr::config(ssl_verifypeer = FALSE)"
	}
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		requestLine += ", httr::authenticate(" + reprStr(user) + ", " + reprStr(pass) + ")"
	}

	requestLine += ")"
	return requestLine
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

	// Escape and wrap in backticks
	escaped := backtickEscapeRegex.ReplaceAllStringFunc(s, func(c string) string {
		if len(c) == 0 {
			return c
		}
		r := []rune(c)[0]
		switch r {
		case '\a':
			return "\\a"
		case '\b':
			return "\\b"
		case '\f':
			return "\\f"
		case '\n':
			return "\\n"
		case '\r':
			return "\\r"
		case '\t':
			return "\\t"
		case '\v':
			return "\\v"
		case '\\':
			return "\\\\"
		case '`':
			return "\\`"
		default:
			// Unicode escape
			hex := ""
			if r <= 0xFF {
				hex = "\\x" + formatHex(r, 2)
			} else if r <= 0xFFFF {
				hex = "\\u" + formatHex(r, 4)
			} else {
				hex = "\\U" + formatHex(r, 8)
			}
			return hex
		}
	})
	return "`" + escaped + "`"
}

func reprStr(s string) string {
	// R prefers double quotes, use single if string contains double quotes
	quote := `"`
	if strings.Contains(s, `"`) && !strings.Contains(s, `'`) {
		quote = `'`
	}

	// Simple Python-style escaping
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, quote, `\`+quote)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)

	return quote + s + quote
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}

func formatHex(r rune, width int) string {
	return fmt.Sprintf("%0*X", width, r)
}
