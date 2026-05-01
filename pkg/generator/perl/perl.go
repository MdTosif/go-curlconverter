package perl

import (
	"fmt"
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

	// Simple case for GET/HEAD with no extra options
	if (strings.ToUpper(method) == "GET" || strings.ToUpper(method) == "HEAD") &&
		len(r.HeaderKV) == 0 &&
		r.BasicAuth == "" &&
		r.Body == "" &&
		len(r.FormParts) == 0 &&
		!r.Insecure &&
		r.MaxTime == "" &&
		r.MaxRedirects == "" {
		code := "use LWP::Simple;\n"
		code += fmt.Sprintf("$content = %s(%s);\n", strings.ToLower(method), reprStr(r.URLs[0].URL))
		return code
	}

	methods := []string{"DELETE", "GET", "HEAD", "PATCH", "POST", "PUT"}
	helperFunction := contains(methods, strings.ToUpper(method))

	var code string
	code += "use LWP::UserAgent;\n"
	if r.BasicAuth != "" {
		code += "use MIME::Base64;\n"
	}
	if !helperFunction {
		code += "require HTTP::Request;\n"
	}
	if r.BodyFile != "" {
		code += "use File::Slurp;\n"
	}
	code += "\n"

	code += "$ua = LWP::UserAgent->new("
	uaArgs := []string{}
	if r.Insecure {
		uaArgs = append(uaArgs, "ssl_opts => { verify_hostname => 0 }")
	}
	if r.MaxTime != "" {
		uaArgs = append(uaArgs, "timeout => "+r.MaxTime)
	}
	if r.MaxRedirects != "" {
		uaArgs = append(uaArgs, "max_redirect => "+r.MaxRedirects)
	}
	if len(uaArgs) > 1 {
		code += "\n    " + strings.Join(uaArgs, ",\n    ") + "\n"
	} else {
		code += strings.Join(uaArgs, ", ")
	}
	code += ");\n"

	args := []string{}
	if !helperFunction {
		code += "$request = HTTP::Request->new("
		args = append(args, reprStr(method))
		args = append(args, reprStr(r.URLs[0].URL))
	} else {
		code += fmt.Sprintf("$response = $ua->%s(", strings.ToLower(method))
		args = append(args, reprStr(r.URLs[0].URL))
		if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
			for _, h := range r.HeaderKV {
				if h.Value == "" {
					continue
				}
				args = append(args, reprHashKey(h.Key)+" => "+reprStr(h.Value))
			}
			if r.BasicAuth != "" {
				user, pass := splitBasicAuth(r.BasicAuth)
				authValue := user + ":" + pass
				args = append(args, `Authorization => "Basic " . MIME::Base64::encode(`+reprStr(authValue)+`)`)
			}
		}
		if r.BodyFile != "" {
			args = append(args, "Content => read_file("+reprStr(r.BodyFile)+")")
		} else if len(r.FormParts) > 0 {
			args = append(args, "Content_Type => 'form-data'")
			lines := []string{}
			for _, part := range r.FormParts {
				if part.IsFile {
					line := reprHashKey(part.Name) + " => [" + reprStr(part.FileName)
					lines = append(lines, line+"]")
				} else {
					lines = append(lines, reprHashKey(part.Name)+" => "+reprStr(part.Value))
				}
			}
			args = append(args, "Content => [\n        "+strings.Join(lines, ",\n        ")+"\n    ]")
		} else if r.Body != "" {
			args = append(args, "Content => "+reprStr(r.Body))
		}
	}

	if (!helperFunction && len(args) > 2) || (helperFunction && len(args) > 1) {
		code += "\n    " + strings.Join(args, ",\n    ") + "\n"
	} else {
		code += strings.Join(args, ", ")
	}
	code += ");\n"

	if !helperFunction {
		code += "$response = $ua->request($request);\n"
	}

	return code
}

func reprStr(s string) string {
	needsEscaping := regexp.MustCompile(`\p{C}|[^ \P{Z}]`)
	if !needsEscaping.MatchString(s) {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "'", "\\'") + "'"
	}

	regexEscape := regexp.MustCompile(`\$|@|"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "$":
			return "\\$"
		case "@":
			return "\\@"
		case "\\":
			return "\\\\"
		case "\a":
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
		default:
			if len(c) == 1 {
				hex := fmt.Sprintf("%02X", c[0])
				return "\\x" + hex
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "\\x{" + hex + "}"
		}
	})
	return "\"" + escaped + "\""
}

func reprHashKey(s string) string {
	hashKeySafe := regexp.MustCompile(`^[a-zA-Z]+$`)
	if hashKeySafe.MatchString(s) {
		return s
	}
	return reprStr(s)
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
