package request

// Request and RequestURL represent the parsed curl request used by the Go
// converter pipeline.

type RequestURL struct {
	URL string
}

type Header struct {
	Key   string
	Value string
}

type FormPart struct {
	Name     string
	Value    string
	FileName string
	IsFile   bool
}

type Request struct {
	URLs        []RequestURL
	Method      string
	Headers     map[string]string
	HeaderKV    []Header
	BasicAuth   string
	BearerToken string
	DigestAuth  bool
	FormParts   []FormPart
	Proxy       string
	ProxyAuth   string
	BodyFile    string
	HasBody     bool
	JSONBody    bool
	Body        string
}
