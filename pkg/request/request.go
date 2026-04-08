package request

// Request and related types represent the parsed curl request used by the Go
// converter pipeline. The exported JSON shape intentionally mirrors the
// upstream parser fixtures for the currently supported subset, while hidden
// fields retain convenience state for generators.

type URLObject struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Path     string `json:"path"`
	Query    string `json:"query"`
	Fragment string `json:"fragment"`
}

type RequestURL struct {
	OriginalURL          string      `json:"originalUrl"`
	URLWithoutQueryList  string      `json:"urlWithoutQueryList"`
	URL                  string      `json:"url"`
	URLObj               URLObject   `json:"urlObj"`
	URLWithOriginalQuery string      `json:"urlWithOriginalQuery"`
	URLWithoutQueryArray string      `json:"urlWithoutQueryArray"`
	Method               string      `json:"method"`
	QueryList            [][2]string `json:"queryList,omitempty"`
	QueryDict            [][2]string `json:"queryDict,omitempty"`
	QueryArray           []string    `json:"queryArray,omitempty"`
	URLQueryArray        []string    `json:"urlQueryArray,omitempty"`
}

type Header struct {
	Key   string
	Value string
}

type HeaderSet struct {
	Lowercase bool        `json:"lowercase"`
	Headers   [][2]string `json:"headers"`
}

type FormPart struct {
	Name     string
	Value    string
	FileName string
	IsFile   bool
}

type Request struct {
	URLs          []RequestURL `json:"urls"`
	AuthType      string       `json:"authType"`
	ProxyAuthType string       `json:"proxyAuthType"`
	HeadersOut    HeaderSet    `json:"headers"`
	ProxyHeaders  HeaderSet    `json:"proxyHeaders"`
	Cookies       [][2]string  `json:"cookies,omitempty"`
	CookieFiles   []string     `json:"cookieFiles,omitempty"`
	Compressed    bool         `json:"compressed,omitempty"`
	Data          string       `json:"data,omitempty"`
	DataArray     []string     `json:"dataArray,omitempty"`
	IsDataRaw     *bool        `json:"isDataRaw,omitempty"`
	IsDataBinary  *bool        `json:"isDataBinary,omitempty"`

	Headers         map[string]string `json:"-"`
	HeaderKV        []Header          `json:"-"`
	Method          string            `json:"-"`
	BasicAuth       string            `json:"-"`
	BearerToken     string            `json:"-"`
	DigestAuth      bool              `json:"-"`
	FormParts       []FormPart        `json:"-"`
	Proxy           string            `json:"-"`
	ProxyAuth       string            `json:"-"`
	BodyFile        string            `json:"-"`
	HasBody         bool              `json:"-"`
	JSONBody        bool              `json:"-"`
	Body            string            `json:"-"`
	AutoContentType bool              `json:"-"`
}
