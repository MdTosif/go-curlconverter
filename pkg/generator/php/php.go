package php

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	var cookieString string
	for _, h := range r.HeaderKV {
		if strings.ToLower(h.Key) == "cookie" {
			cookieString = h.Value
			break
		}
	}

	var phpCode string
	phpCode += "<?php\n"
	phpCode += "$ch = curl_init();\n"
	phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_URL, %s);\n", reprStr(r.URLs[0].URL))
	phpCode += "curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);\n"
	phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_CUSTOMREQUEST, %s);\n", reprStr(r.URLs[0].Method))

	if r.Compressed {
		hasAcceptEncoding := false
		for _, h := range r.HeaderKV {
			if strings.ToLower(h.Key) == "accept-encoding" {
				hasAcceptEncoding = true
				break
			}
		}
		if !hasAcceptEncoding {
			r.HeaderKV = append(r.HeaderKV, request.Header{Key: "Accept-Encoding", Value: "gzip"})
		}
	}

	if len(r.HeaderKV) > 0 {
		headersArrayCode := "[\n"
		for _, h := range r.HeaderKV {
			if h.Value == "" {
				continue
			}
			headersArrayCode += fmt.Sprintf("    %s,\n", reprStr(h.Key+": "+h.Value))
		}
		headersArrayCode += "]"
		phpCode += "curl_setopt($ch, CURLOPT_HTTPHEADER, " + headersArrayCode + ");\n"
	}

	if cookieString != "" {
		phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_COOKIE, %s);\n", reprStr(cookieString))
	}

	if r.BasicAuth != "" && (r.AuthType == "basic" || r.AuthType == "digest") {
		authType := "CURLAUTH_BASIC"
		if r.AuthType == "digest" {
			authType = "CURLAUTH_DIGEST"
		}
		phpCode += "curl_setopt($ch, CURLOPT_HTTPAUTH, " + authType + ");\n"
		phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_USERPWD, %s);\n", reprStr(r.BasicAuth))
	}

	if r.Body != "" || len(r.FormParts) > 0 {
		requestDataCode := ""
		if len(r.FormParts) > 0 {
			requestDataCode = "[\n"
			for _, part := range r.FormParts {
				requestDataCode += "    " + reprStr(part.Name) + " => "
				if part.IsFile {
					requestDataCode += "new CURLFile(" + reprStr(part.FileName)
					requestDataCode += ")"
				} else {
					if part.FileName != "" {
						requestDataCode += "new CURLStringFile(" + reprStr(part.Value)
						requestDataCode += ", " + reprStr(part.FileName)
						requestDataCode += ")"
					} else {
						requestDataCode += reprStr(part.Value)
					}
				}
				requestDataCode += ",\n"
			}
			requestDataCode += "]"
		} else if r.IsDataBinary != nil && *r.IsDataBinary && strings.HasPrefix(r.Body, "@") {
			requestDataCode = "file_get_contents(" + reprStr(r.Body[1:]) + ")"
		} else {
			requestDataCode = reprStr(r.Body)
		}
		phpCode += "curl_setopt($ch, CURLOPT_POSTFIELDS, " + requestDataCode + ");\n"
	}

	if r.Proxy != "" {
		phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_PROXY, %s);\n", reprStr(r.Proxy))
		if r.ProxyAuth != "" {
			phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_PROXYUSERPWD, %s);\n", reprStr(r.ProxyAuth))
		}
	}

	if r.MaxTime != "" {
		if timeout, err := strconv.Atoi(r.MaxTime); err == nil {
			phpCode += fmt.Sprintf("curl_setopt($ch, CURLOPT_TIMEOUT, %d);\n", timeout)
		}
	}

	if r.FollowRedirects {
		phpCode += "curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);\n"
	}

	if r.Insecure {
		phpCode += "curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);\n"
		phpCode += "curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);\n"
	}

	phpCode += "\n$response = curl_exec($ch);\n\n"
	phpCode += "curl_close($ch);\n"
	return phpCode
}

func reprStr(s string) string {
	regexSingleEscape := regexp.MustCompile(`'|\\`)
	regexDoubleEscape := regexp.MustCompile(`"|\$|\\|\p{C}|[^ \P{Z}]`)

	quote := "'"
	regex := regexSingleEscape
	if (strings.Contains(s, "'") && !strings.Contains(s, "\"")) || regexp.MustCompile(`[^\x20-\x7E]`).MatchString(s) {
		quote = "\""
		regex = regexDoubleEscape
	}

	escaped := regex.ReplaceAllStringFunc(s, func(c string) string {
		switch c {
		case "$":
			if quote == "'" {
				return "$"
			}
			return "\\$"
		case "\\":
			return "\\\\"
		case "'":
			if c == quote {
				return "\\" + c
			}
			return c
		case "\"":
			if c == quote {
				return "\\" + c
			}
			return c
		case "\n":
			return "\\n"
		case "\r":
			return "\\r"
		case "\t":
			return "\\t"
		case "\v":
			return "\\v"
		case "\x1b":
			return "\\e"
		case "\f":
			return "\\f"
		}

		if len(c) == 0 {
			return ""
		}
		hex := fmt.Sprintf("%X", c[0])
		if len(hex) > 2 {
			return "\\u{" + hex + "}"
		}
		return "\\x" + hex
	})
	return quote + escaped + quote
}
