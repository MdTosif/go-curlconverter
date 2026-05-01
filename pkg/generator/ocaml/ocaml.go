package ocaml

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

	var code string
	code += "open Lwt\n"
	code += "open Cohttp\n"
	code += "open Cohttp_lwt_unix\n"
	code += "\n"

	method := r.URLs[0].Method
	if method == "" {
		method = "GET"
	}

	methodConstants := []string{"GET", "POST", "HEAD", "DELETE", "PATCH", "PUT", "OPTIONS", "TRACE", "CONNECT"}
	methodFns := []string{"HEAD", "GET", "DELETE", "POST", "PUT", "PATCH"}
	bodyMethods := []string{"POST", "PUT", "PATCH"}

	code += "let uri = Uri.of_string " + reprStr(r.URLs[0].URL) + " in\n"

	if !contains(methodConstants, strings.ToUpper(method)) {
		code += "let meth = Code.method_of_string " + reprStr(method) + " in\n"
	}

	hasBody := r.Body != ""

	if len(r.HeaderKV) == 1 && r.BasicAuth == "" {
		h := r.HeaderKV[0]
		if h.Value != "" {
			code += "let headers = Header.init_with " + reprStr(h.Key) + " " + reprStr(h.Value) + " in\n"
		}
	} else if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
		code += "let headers = Header.init ()"
		for _, h := range r.HeaderKV {
			if h.Value != "" {
				code += "\n  |> fun h -> Header.add h " + reprStr(h.Key) + " " + reprStr(h.Value)
			}
		}
		if r.BasicAuth != "" {
			user, pass := splitBasicAuth(r.BasicAuth)
			code += "\n  |> fun h -> Header.add_authorization h (`Basic (" + reprStr(user) + ", " + reprStr(pass) + "))"
		}
		code += " in\n"
	}

	if r.Body != "" {
		code += "let body = Cohttp_lwt.Body.of_string " + reprStr(r.Body) + " in\n"
	}

	fn := "Client.call"
	args := []string{}
	if len(r.HeaderKV) > 0 || r.BasicAuth != "" {
		args = append(args, "~headers")
	}
	if hasBody {
		args = append(args, "~body")
	}
	if !contains(methodConstants, strings.ToUpper(method)) {
		args = append(args, "meth")
	} else {
		if contains(methodFns, strings.ToUpper(method)) && (!hasBody || contains(bodyMethods, strings.ToUpper(method))) {
			fn = "Client." + strings.ToLower(method)
		} else {
			args = append(args, "`"+strings.ToUpper(method))
		}
	}
	args = append(args, "uri")

	code += fn + " " + strings.Join(args, " ") + "\n"
	code += ">>= fun (resp, body) ->\n"
	code += "  (* Do stuff with the result *)\n"

	return code
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "\\":
			return "\\\\"
		case "\"":
			return "\\\""
		case "\n":
			return "\\n"
		case "\r":
			return "\\r"
		case "\t":
			return "\\t"
		case "\b":
			return "\\b"
		default:
			if len(c) == 1 {
				hex := fmt.Sprintf("%02X", c[0])
				return "\\x" + hex
			}
			hex := fmt.Sprintf("%04X", c[0])
			return "\\u{" + hex + "}"
		}
	})
	return "\"" + escaped + "\""
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
