package har

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/mdtosif/go-curlconverter/pkg/request"
)

type HAR struct {
	Log Log `json:"log"`
}

type Log struct {
	Version string  `json:"version"`
	Entries []Entry `json:"entries"`
}

type Entry struct {
	Request RequestEntry `json:"request"`
}

type RequestEntry struct {
	Method      string       `json:"method"`
	URL         string       `json:"url"`
	HTTPVersion string       `json:"httpVersion"`
	Headers     []Header     `json:"headers"`
	Cookies     []Cookie     `json:"cookies"`
	QueryString []QueryParam `json:"queryString"`
	PostData    *PostData    `json:"postData,omitempty"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type QueryParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PostData struct {
	MimeType string  `json:"mimeType"`
	Text     string  `json:"text,omitempty"`
	Params   []Param `json:"params,omitempty"`
}

type Param struct {
	Name     string `json:"name"`
	Value    string `json:"value,omitempty"`
	FileName string `json:"fileName,omitempty"`
}

func Generate(r *request.Request) string {
	if r == nil || len(r.URLs) == 0 {
		return "{}"
	}

	u := r.URLs[0].URL
	method := r.Method
	if method == "" {
		method = "GET"
	}

	entry := Entry{
		Request: RequestEntry{
			Method:      method,
			URL:         u,
			HTTPVersion: "HTTP/1.1",
			Headers:     []Header{},
			Cookies:     []Cookie{},
			QueryString: []QueryParam{},
		},
	}

	// Parse URL for query params
	if parsedURL, err := url.Parse(u); err == nil {
		if query := parsedURL.Query(); len(query) > 0 {
			for key, vals := range query {
				for _, val := range vals {
					entry.Request.QueryString = append(entry.Request.QueryString, QueryParam{
						Name:  key,
						Value: val,
					})
				}
			}
		}
	}

	// Headers
	for _, h := range r.HeaderKV {
		entry.Request.Headers = append(entry.Request.Headers, Header{
			Name:  h.Key,
			Value: h.Value,
		})
	}

	// Cookies from CookieJar
	if r.CookieJar != "" {
		pairs := strings.Split(r.CookieJar, "; ")
		for _, pair := range pairs {
			if parts := strings.SplitN(pair, "=", 2); len(parts) == 2 {
				entry.Request.Cookies = append(entry.Request.Cookies, Cookie{
					Name:  strings.TrimSpace(parts[0]),
					Value: parts[1],
				})
			}
		}
	}

	// Basic auth
	if r.BasicAuth != "" {
		entry.Request.Headers = append(entry.Request.Headers, Header{
			Name:  "Authorization",
			Value: "Basic " + base64.StdEncoding.EncodeToString([]byte(r.BasicAuth)),
		})
	}

	// Post data
	if len(r.FormParts) > 0 {
		entry.Request.PostData = &PostData{
			MimeType: "multipart/form-data",
			Params:   []Param{},
		}
		for _, part := range r.FormParts {
			p := Param{Name: part.Name}
			if part.IsFile {
				p.FileName = part.FileName
			} else {
				p.Value = part.Value
			}
			entry.Request.PostData.Params = append(entry.Request.PostData.Params, p)
		}
	} else if r.HasBody && r.Body != "" {
		contentType := "text/plain"
		for _, h := range r.HeaderKV {
			if strings.EqualFold(h.Key, "Content-Type") {
				contentType = h.Value
				break
			}
		}
		entry.Request.PostData = &PostData{
			MimeType: contentType,
			Text:     r.Body,
		}
	}

	har := HAR{
		Log: Log{
			Version: "1.2",
			Entries: []Entry{entry},
		},
	}

	result, _ := json.MarshalIndent(har, "", "  ")
	return string(result)
}
