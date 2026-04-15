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
	CookieJar     string       `json:"cookieJar,omitempty"`
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

	// Redirect options
	FollowRedirects        bool   `json:"-"`
	MaxRedirects           string `json:"-"`
	FollowRedirectsTrusted bool   `json:"-"`
	Post301                bool   `json:"-"`
	Post302                bool   `json:"-"`
	Post303                bool   `json:"-"`

	// Timeout options
	ConnectTimeout string `json:"-"`
	MaxTime        string `json:"-"`

	// TLS/SSL options
	Insecure     bool   `json:"-"`
	CACert       string `json:"-"`
	Cert         string `json:"-"`
	Key          string `json:"-"`
	CertType     string `json:"-"`
	KeyType      string `json:"-"`
	Pass         string `json:"-"`
	CAPath       string `json:"-"`
	CRLFile      string `json:"-"`
	PinnedPubKey string `json:"-"`

	// Transfer options
	Range      string `json:"-"`
	ContinueAt string `json:"-"`
	SpeedLimit string `json:"-"`
	SpeedTime  string `json:"-"`

	// Retry options
	Retry        string `json:"-"`
	RetryDelay   string `json:"-"`
	RetryMaxTime string `json:"-"`

	// Output options
	Output        string `json:"-"`
	RemoteName    bool   `json:"-"`
	RemoteNameAll bool   `json:"-"`
	Clobber       bool   `json:"-"`
	RemoteTime    bool   `json:"-"`
	Include       bool   `json:"-"`

	// Protocol options
	HTTP10              bool `json:"-"`
	HTTP11              bool `json:"-"`
	HTTP2               bool `json:"-"`
	HTTP3               bool `json:"-"`
	HTTP2PriorKnowledge bool `json:"-"`
	IPv4                bool `json:"-"`
	IPv6                bool `json:"-"`

	// Proxy options
	ProxyType      string `json:"-"`
	ProxyTunnel    bool   `json:"-"`
	NoProxy        string `json:"-"`
	PreProxy       string `json:"-"`
	ProxyCACert    string `json:"-"`
	ProxyCAPath    string `json:"-"`
	ProxyCert      string `json:"-"`
	ProxyCertType  string `json:"-"`
	ProxyKey       string `json:"-"`
	ProxyKeyType   string `json:"-"`
	ProxyPass      string `json:"-"`
	ProxyInsecure  bool   `json:"-"`
	ProxyDigest    bool   `json:"-"`
	ProxyBasic     bool   `json:"-"`
	ProxyNegotiate bool   `json:"-"`
	ProxyNtlm      bool   `json:"-"`

	// SOCKS proxy options
	SOCKS4         string `json:"-"`
	SOCKS4a        string `json:"-"`
	SOCKS5         string `json:"-"`
	SOCKS5Hostname string `json:"-"`

	// Misc
	Verbose             bool   `json:"-"`
	Silent              bool   `json:"-"`
	Fail                bool   `json:"-"`
	LocationTrusted     bool   `json:"-"`
	IgnoreContentLength bool   `json:"-"`
	Globoff             bool   `json:"-"`
	Netrc               string `json:"-"`
	NetrcFile           string `json:"-"`
	Stdin               bool   `json:"-"`
	StdinFile           string `json:"-"`
}
