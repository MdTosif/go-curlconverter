package ansible

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type yamlField struct {
	Key   string
	Value any
}

var (
	plainScalarPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_./+-]*$`)
	plainKeyPattern    = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return ""
	}

	lines := []string{
		"-",
		"  name: " + yamlQuoted(requestName(r)),
		"  uri:",
		"    url: " + yamlQuoted(r.URLs[0].URL),
		"    method: " + yamlAuto(methodFor(r)),
	}

	if bodyValue, bodyFormat, src, ok := renderBody(r); ok {
		if src != "" {
			lines = append(lines, "    src: "+yamlAuto(src))
		} else {
			lines = appendYAMLField(lines, 4, "body", bodyValue)
			if bodyFormat != "" {
				lines = append(lines, "    body_format: "+bodyFormat)
			}
		}
	}

	if r.Output != "" {
		lines = append(lines, "    dest: "+yamlAuto(r.Output))
	}

	if len(r.HeaderKV) > 0 {
		lines = append(lines, "    headers:")
		for _, h := range r.HeaderKV {
			lines = append(lines, "      "+yamlKey(h.Key)+": "+yamlAuto(h.Value))
		}
	}

	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		if user != "" {
			lines = append(lines, "    url_username: "+yamlAuto(user))
		}
		if pass != "" {
			lines = append(lines, "    url_password: "+yamlAuto(pass))
		}
	}

	if r.Insecure {
		lines = append(lines, "    validate_certs: false")
	}

	lines = append(lines, "  register: result")
	return strings.Join(lines, "\n") + "\n"
}

func methodFor(r *request.Request) string {
	if r.Method != "" {
		return r.Method
	}
	return r.URLs[0].Method
}

func requestName(r *request.Request) string {
	if r.URLs[0].URLObj.Query != "" && !strings.Contains(r.URLs[0].URLObj.Query, "=") && r.URLs[0].URLWithOriginalQuery != "" {
		return r.URLs[0].URLWithOriginalQuery
	}
	return r.URLs[0].URLWithoutQueryList
}

func renderBody(r *request.Request) (any, string, string, bool) {
	if len(r.FormParts) > 0 {
		form := make([]yamlField, 0, len(r.FormParts))
		for _, part := range r.FormParts {
			entry := []yamlField{}
			if part.IsFile {
				entry = append(entry, yamlField{Key: "filename", Value: part.FileName})
			} else {
				entry = append(entry, yamlField{Key: "content", Value: part.Value})
			}
			form = append(form, yamlField{Key: part.Name, Value: entry})
		}
		return form, "form-multipart", "", true
	}

	if r.BodyFile != "" {
		return nil, "", r.BodyFile, true
	}

	if !r.HasBody {
		return nil, "", "", false
	}

	if r.Body != "" && hasExplicitJSONContentType(r) {
		if value, ok := parseOrderedJSON(normalizeJSONString(r.Body)); ok {
			return value, "json", "", true
		}
	}

	if r.Body != "" && hasExplicitFormContentType(r) {
		if pairs, ok := parseOrderedForm(r.Body); ok {
			return pairs, "form-urlencoded", "", true
		}
	}

	return r.Body, "", "", true
}

func hasExplicitJSONContentType(r *request.Request) bool {
	if r.AutoContentType {
		return false
	}
	for _, h := range r.HeaderKV {
		if !strings.EqualFold(h.Key, "Content-Type") {
			continue
		}
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(h.Value, ";")[0]))
		return mediaType == "application/json"
	}
	return false
}

func hasExplicitFormContentType(r *request.Request) bool {
	if r.AutoContentType {
		return false
	}
	for _, h := range r.HeaderKV {
		if !strings.EqualFold(h.Key, "Content-Type") {
			continue
		}
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(h.Value, ";")[0]))
		return mediaType == "application/x-www-form-urlencoded"
	}
	return false
}

func normalizeJSONString(raw string) string {
	if strings.Contains(raw, `\"`) {
		raw = strings.ReplaceAll(raw, `\"`, `"`)
	}
	return raw
}

func parseOrderedForm(raw string) ([]yamlField, bool) {
	if raw == "" {
		return nil, false
	}
	parts := strings.Split(raw, "&")
	fields := make([]yamlField, 0, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, false
		}
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			return nil, false
		}
		decodedValue, err := url.QueryUnescape(value)
		if err != nil {
			return nil, false
		}
		fields = append(fields, yamlField{Key: decodedKey, Value: decodedValue})
	}
	return fields, true
}

func parseOrderedJSON(raw string) (any, bool) {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()

	value, err := decodeJSONValue(dec)
	if err != nil {
		return nil, false
	}

	if _, err := dec.Token(); err == nil {
		return nil, false
	}
	return value, true
}

func decodeJSONValue(dec *json.Decoder) (any, error) {
	token, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch tok := token.(type) {
	case json.Delim:
		switch tok {
		case '{':
			fields := []yamlField{}
			for dec.More() {
				keyToken, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key := keyToken.(string)
				value, err := decodeJSONValue(dec)
				if err != nil {
					return nil, err
				}
				fields = append(fields, yamlField{Key: key, Value: value})
			}
			_, err := dec.Token()
			return fields, err
		case '[':
			items := []any{}
			for dec.More() {
				value, err := decodeJSONValue(dec)
				if err != nil {
					return nil, err
				}
				items = append(items, value)
			}
			_, err := dec.Token()
			return items, err
		}
	case string:
		return tok, nil
	case json.Number:
		return tok, nil
	case bool:
		return tok, nil
	case nil:
		return nil, nil
	}

	return nil, nil
}

func appendYAMLField(lines []string, indent int, key string, value any) []string {
	prefix := strings.Repeat(" ", indent) + yamlKey(key) + ":"
	switch v := value.(type) {
	case []yamlField:
		lines = append(lines, prefix)
		for _, field := range v {
			lines = appendYAMLField(lines, indent+2, field.Key, field.Value)
		}
	case []any:
		lines = append(lines, prefix)
		for _, item := range v {
			lines = appendYAMLListItem(lines, indent+2, item)
		}
	case json.Number:
		lines = append(lines, prefix+" "+v.String())
	case bool:
		lines = append(lines, prefix+" "+strconv.FormatBool(v))
	case nil:
		lines = append(lines, prefix+" null")
	default:
		lines = append(lines, prefix+" "+yamlAuto(v.(string)))
	}
	return lines
}

func appendYAMLListItem(lines []string, indent int, value any) []string {
	prefix := strings.Repeat(" ", indent) + "-"
	switch v := value.(type) {
	case []yamlField:
		lines = append(lines, prefix)
		for _, field := range v {
			lines = appendYAMLField(lines, indent+2, field.Key, field.Value)
		}
	case []any:
		lines = append(lines, prefix)
		for _, item := range v {
			lines = appendYAMLListItem(lines, indent+2, item)
		}
	case json.Number:
		lines = append(lines, prefix+" "+v.String())
	case bool:
		lines = append(lines, prefix+" "+strconv.FormatBool(v))
	case nil:
		lines = append(lines, prefix+" null")
	default:
		lines = append(lines, prefix+" "+yamlAuto(v.(string)))
	}
	return lines
}

func yamlKey(s string) string {
	if plainKeyPattern.MatchString(s) {
		return s
	}
	return yamlQuoted(s)
}

func yamlAuto(s string) string {
	if s == "" {
		return `""`
	}
	if plainScalarPattern.MatchString(s) {
		return s
	}
	return yamlQuoted(s)
}

func yamlQuoted(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func splitBasicAuth(auth string) (string, string) {
	user, pass, ok := strings.Cut(auth, ":")
	if !ok {
		return auth, ""
	}
	return user, pass
}
