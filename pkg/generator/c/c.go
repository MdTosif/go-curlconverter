package c

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

	var preamble string
	preamble += "int main(int argc, char *argv[])\n"
	preamble += "{\n"
	preamble += "  CURLcode ret;\n"
	preamble += "  CURL *hnd;\n"

	var vars string
	var code string
	var cleanup string

	code += "  hnd = curl_easy_init();\n"
	cleanup += "  curl_easy_cleanup(hnd);\n"
	cleanup += "  hnd = NULL;\n"

	code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_URL, %s);\n", reprStr(r.URLs[0].URL))
	code += "  curl_easy_setopt(hnd, CURLOPT_NOPROGRESS, 1L);\n"

	if r.BasicAuth != "" {
		code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_USERPWD, %s);\n", reprStr(r.BasicAuth))
		if r.AuthType == "digest" {
			code += "  curl_easy_setopt(hnd, CURLOPT_HTTPAUTH, (long)CURLAUTH_DIGEST);\n"
		} else {
			code += "  curl_easy_setopt(hnd, CURLOPT_HTTPAUTH, (long)CURLAUTH_BASIC);\n"
		}
	}

	if r.Body != "" || len(r.FormParts) > 0 {
		if len(r.FormParts) > 0 {
			preamble += "  curl_mime *mime1;\n"
			preamble += "  curl_mimepart *part1;\n"
			vars += "  mime1 = NULL;\n"
			code += "  mime1 = curl_mime_init(hnd);\n"
			for _, part := range r.FormParts {
				code += "  part1 = curl_mime_addpart(mime1);\n"
				if part.IsFile {
					code += fmt.Sprintf("  curl_mime_filedata(part1, %s);\n", reprStr(part.FileName))
				} else {
					code += fmt.Sprintf("  curl_mime_data(part1, %s, CURL_ZERO_TERMINATED);\n", reprStr(part.Value))
				}
				if part.FileName != "" {
					code += fmt.Sprintf("  curl_mime_filename(part1, %s);\n", reprStr(part.FileName))
				}
				code += fmt.Sprintf("  curl_mime_name(part1, %s);\n", reprStr(part.Name))
			}
			code += "  curl_easy_setopt(hnd, CURLOPT_MIMEPOST, mime1);\n"
			cleanup += "  curl_mime_free(mime1);\n"
			cleanup += "  mime1 = NULL;\n"
		} else {
			code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_POSTFIELDS, %s);\n", reprStr(r.Body))
			code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_POSTFIELDSIZE_LARGE, (curl_off_t)%d);\n", len(r.Body))
		}
	}

	headerLines := []string{}
	for _, h := range r.HeaderKV {
		if h.Value != "" {
			lowerKey := strings.ToLower(h.Key)
			if lowerKey == "user-agent" || lowerKey == "referer" {
				continue
			}
			headerLines = append(headerLines, fmt.Sprintf("  headers = curl_slist_append(headers, %s);\n", reprStr(h.Key+": "+h.Value)))
		}
	}

	if len(headerLines) > 0 {
		preamble += "  struct curl_slist *headers;\n"
		vars += "  headers = NULL;\n"
		for _, line := range headerLines {
			vars += line
		}
		code += "  curl_easy_setopt(hnd, CURLOPT_HTTPHEADER, headers);\n"
		cleanup += "  curl_slist_free_all(headers);\n"
		cleanup += "  headers = NULL;\n"
	}

	for _, h := range r.HeaderKV {
		lowerKey := strings.ToLower(h.Key)
		if lowerKey == "referer" {
			code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_REFERER, %s);\n", reprStr(h.Value))
		}
		if lowerKey == "user-agent" {
			code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_USERAGENT, %s);\n", reprStr(h.Value))
		}
	}

	if r.FollowRedirects {
		code += "  curl_easy_setopt(hnd, CURLOPT_FOLLOWLOCATION, 1L);\n"
	}

	if r.Insecure {
		code += "  curl_easy_setopt(hnd, CURLOPT_SSL_VERIFYPEER, 0L);\n"
		code += "  curl_easy_setopt(hnd, CURLOPT_SSL_VERIFYHOST, 0L);\n"
	}

	if r.MaxTime != "" {
		if timeout, err := strconv.ParseFloat(r.MaxTime, 64); err == nil {
			code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_TIMEOUT_MS, %dL);\n", int(timeout*1000))
		}
	}

	if r.ConnectTimeout != "" {
		if timeout, err := strconv.ParseFloat(r.ConnectTimeout, 64); err == nil {
			code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_CONNECTTIMEOUT_MS, %dL);\n", int(timeout*1000))
		}
	}

	if r.Proxy != "" {
		code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_PROXY, %s);\n", reprStr(r.Proxy))
	}

	expectedMethod := "GET"
	if r.Body != "" || len(r.FormParts) > 0 {
		expectedMethod = "POST"
	}
	if strings.ToUpper(r.URLs[0].Method) != expectedMethod {
		code += fmt.Sprintf("  curl_easy_setopt(hnd, CURLOPT_CUSTOMREQUEST, %s);\n", reprStr(r.URLs[0].Method))
	}

	code += "  ret = curl_easy_perform(hnd);\n"
	code += "  if (ret != CURLE_OK) {\n"
	code += "    fprintf(stderr, \"curl_easy_perform() failed: %s\\n\", curl_easy_strerror(ret));\n"
	code += "  }\n"

	var result string
	result += preamble
	result += "\n"
	result += vars
	result += "\n"
	result += code
	result += "\n"
	result += cleanup
	result += "  return 0;\n"
	result += "}\n"

	return result
}

func reprStr(s string) string {
	regexEscape := regexp.MustCompile(`"|\\|\p{C}|[^ \P{Z}]`)
	escaped := regexEscape.ReplaceAllStringFunc(s, func(c string) string {
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
	return "\"" + escaped + "\""
}
