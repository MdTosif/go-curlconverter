package json

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type JSONOutput struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Cookies map[string]string `json:"cookies,omitempty"`
	Data    any               `json:"data,omitempty"`
	Auth    *Auth             `json:"auth,omitempty"`
	Files   []File            `json:"files,omitempty"`
	RawBody string            `json:"raw_body,omitempty"`
}

type Auth struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type File struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
	Data string `json:"data,omitempty"`
}

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return "{}"
	}

	output := JSONOutput{
		URL:    r.URLs[0].URL,
		Method: r.Method,
	}

	if output.Method == "" {
		output.Method = "GET"
	}

	// Headers
	if len(r.HeaderKV) > 0 {
		output.Headers = make(map[string]string)
		for _, h := range r.HeaderKV {
			output.Headers[h.Key] = h.Value
		}
	}

	// Cookies from CookieJar
	if r.CookieJar != "" {
		output.Cookies = make(map[string]string)
		pairs := strings.Split(r.CookieJar, "; ")
		for _, pair := range pairs {
			if parts := strings.SplitN(pair, "=", 2); len(parts) == 2 {
				output.Cookies[strings.TrimSpace(parts[0])] = parts[1]
			}
		}
	}

	// Auth
	if r.BasicAuth != "" {
		user, pass := splitBasicAuth(r.BasicAuth)
		output.Auth = &Auth{
			Type:     "basic",
			Username: user,
			Password: pass,
		}
	}

	// Body/Data
	if len(r.FormParts) > 0 {
		for _, part := range r.FormParts {
			f := File{Name: part.Name}
			if part.IsFile {
				f.Path = part.FileName
			} else {
				f.Data = part.Value
			}
			output.Files = append(output.Files, f)
		}
	} else if r.HasBody && r.Body != "" {
		// Try parse as JSON
		var parsed any
		if err := json.Unmarshal([]byte(r.Body), &parsed); err == nil {
			output.Data = parsed
		} else {
			// Try parse as query string
			if values, err := url.ParseQuery(r.Body); err == nil && len(values) > 0 {
				data := make(map[string]string)
				for key, vals := range values {
					if len(vals) > 0 {
						data[key] = vals[0]
					}
				}
				output.Data = data
			} else {
				output.RawBody = r.Body
			}
		}
	}

	result, _ := json.MarshalIndent(output, "", "  ")
	return string(result)
}

func splitBasicAuth(auth string) (string, string) {
	parts := strings.SplitN(auth, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return auth, ""
}
