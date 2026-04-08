package request

// Request and RequestURL are small types representing the parsed curl request.
// This is intentionally minimal for the MVP: method, URL, headers and body.

type RequestURL struct {
	URL string
}

type Header struct {
	Key   string
	Value string
}

type Request struct {
	URLs     []RequestURL
	Method   string
	Headers  map[string]string
	HeaderKV []Header
	HasBody  bool
	Body     string
}
