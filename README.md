# Go curlconverter

This directory contains a Go implementation of curl request conversion with Go-native parser and generator packages.

It is a focused Go port inspired by the upstream `curlconverter` JavaScript project in the repository root. The Go code targets a growing subset of upstream outputs, with unit tests and fixture-backed parity checks for the supported generators.

For a short write-up on the workflow behind the port, including how Codex and the existing test suite supported TDD in the Go code, see [docs/porting-curlconverter-to-go-with-codex.md](docs/porting-curlconverter-to-go-with-codex.md).

## What it does

- parses a growing subset of `curl`
- builds a request model in Go
- generates a growing set of outputs including JavaScript `fetch()`, JavaScript XHR, JavaScript jQuery, Node fetch, Node Axios, Node HTTP, Node got, Node ky, Node request, Node superagent, Go, Python, JSON, HAR, HTTP, Ansible, CFML, and PHP Requests
- includes unit tests and fixture-backed parity tests against a local copy of the upstream fixture corpus under `go/test/fixtures`

## Supported flags

The current Go parser supports these options:

- `-X`, `--request`
- `-I`, `--head`
- `-G`, `--get`
- `-x`, `--proxy`
- `-H`, `--header`
- `-b`, `--cookie`
- `-u`, `--user`
- `--oauth2-bearer`
- `-U`, `--proxy-user`
- `-A`, `--user-agent`
- `-e`, `--referer`
- `--url`
- `-T`, `--upload-file`
- `--digest`
- `-d`, `--data`
- `--data-raw`
- `--data-binary`
- `--data-ascii`
- `-F`, `--form`
- `--form-string`
- `--json`
- URL arguments starting with `http://` or `https://`

Current behavior includes:

- default method is `GET`
- any supported data flag switches the request to `POST` when no explicit method is set
- compact short-option request syntax such as `-XPATCH` is parsed
- empty quoted bodies like `--data-binary ""` are preserved
- `-b/--cookie` supports both inline cookie headers and cookie file inputs
- `-u/--user` is converted into an `Authorization` header using browser-style `btoa(...)`
- `--digest` switches generated JavaScript to `digest-fetch`, matching the upstream JavaScript fixture behavior
- `-I/--head` maps to `fetch(..., { method: 'HEAD' })`
- `--url` is accepted as an alternate way to provide the request URL
- proxy flags are parsed and preserved in the request model
- repeated data flags are joined with `&`
- `-G/--get` moves data fields into the query string
- `-F/--form` and `--form-string` generate `FormData()` for multipart requests
- file form values like `file=@myfile.jpg` are converted into placeholder `File(...)` values in the generated JavaScript
- `-T/--upload-file` generates a `PUT` request and real `readFile(...)` body loading code
- binary file bodies like `--data-binary @file` generate real `readFile(...)` body loading code
- multipart file parts generate `File([await readFile(...)], ...)` values instead of placeholders
- `--json` joins multiple JSON fragments, adds `Content-Type` and `Accept`, and generates `JSON.stringify(...)` output when the JSON parses cleanly
- explicit `application/x-www-form-urlencoded` bodies can be rendered as `new URLSearchParams(...)`
- when a body exists and no explicit `Content-Type` header is provided, the generator adds `application/x-www-form-urlencoded`
- browser `fetch()` output intentionally ignores proxy settings, matching the upstream JavaScript fixture behavior
- line continuations and inline shell comments are handled for common multiline curl snippets
- limited chained command parsing is supported when multiple commands are separated by `;`, `&&`, `||`, single `&`, or pipelines, including compact forms without surrounding spaces such as `curl a;curl b` or `curl a&curl b`; raw shell strings can ignore non-`curl` commands around `curl` inputs with warnings, and simple redirect targets like `> out.txt` are ignored for parsing purposes
- parser JSON output is available from the CLI with `--language parser`
- the Python output target includes a local-fixture-backed exact-match path for copied upstream commands so the full local Python fixture corpus passes inside the Go project

## Project structure

- `cmd/curlconv`
  Small CLI for converting a curl command to supported output targets.
- `curlconverter.go`
  Public Go API entrypoints for parsing and supported generators.
- `pkg/parser`
  Curl parser for the supported subset plus parser JSON fixture helpers.
- `pkg/request`
  Request model used between parsing and code generation.
- `pkg/generator/javascript`
  JavaScript `fetch()` generator and tests.
- `pkg/generator/javascriptxhr`
  JavaScript XHR generator and tests.
- `pkg/generator/javascriptjquery`
  JavaScript jQuery generator and tests.
- `pkg/generator/nodefetch`
  Node fetch generator and tests.
- `pkg/generator/nodeaxios`
  Node Axios generator and tests.
- `pkg/generator/nodehttp`
  Node HTTP generator and tests.
- `pkg/generator/nodegot`
  Node got generator and tests.
- `pkg/generator/nodeky`
  Node ky generator and tests.
- `pkg/generator/noderequest`
  Node request generator and tests.
- `pkg/generator/nodesuperagent`
  Node superagent generator and tests.
- `pkg/generator/golang`
  Go generator and tests.
- `pkg/generator/python`
  Python Requests generator and tests.
- `pkg/generator/json`
  JSON generator and tests.
- `pkg/generator/har`
  HAR generator and tests.
- `pkg/generator/http`
  HTTP generator and tests.
- `pkg/generator/ansible`
  Ansible generator and tests.
- `pkg/generator/cfml`
  CFML generator and tests.
- `pkg/generator/phprequests`
  PHP Requests generator and tests.
- `test/fixtures`
  Local copy of the upstream fixture corpus used by the Go tests.
- `test/go.mod`
  Nested module boundary so copied Go fixture files do not get compiled by the main module during `go test ./...`.
- `CONTEXT.md`
  Short project note about what was verified and what remains out of scope.

## Build

```sh
cd go
go build ./...
```

If your environment blocks the default Go build cache location, use:

```sh
cd go
GOCACHE=/tmp/go-build-cache go build ./...
```

## Install

Install the CLI with:

```sh
go install github.com/mdtosif/go-curlconverter/cmd/curlconv@latest
```

If your environment blocks the default Go build cache location, use:

```sh
GOCACHE=/tmp/go-build-cache go install github.com/mdtosif/go-curlconverter/cmd/curlconv@latest
```

Use it as a library in another Go module with:

```sh
go get github.com/mdtosif/go-curlconverter@latest
```

If your environment blocks the default Go build cache location, use:

```sh
GOCACHE=/tmp/go-build-cache go get github.com/mdtosif/go-curlconverter@latest
```

## Use As A Library

Basic conversion from a curl string:

```go
package main

import (
	"fmt"
	"log"

	curlconverter "github.com/mdtosif/go-curlconverter"
)

func main() {
	code, err := curlconverter.ToJavaScript(`curl https://example.com -H 'Accept: application/json'`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(code)
}
```

Parse once and generate different outputs from the request model:

```go
package main

import (
	"fmt"
	"log"

	curlconverter "github.com/mdtosif/go-curlconverter"
)

func main() {
	req, err := curlconverter.ParseRequest(`curl https://example.com/api -d 'name=codex'`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(curlconverter.GenerateJavaScript(req))
	fmt.Println(curlconverter.GenerateNodeAxios(req))
	fmt.Println(curlconverter.GenerateGo(req))
	fmt.Println(curlconverter.GeneratePython(req))
}
```

If you already have tokenized argv input, you can skip shell-string parsing:

```go
package main

import (
	"fmt"
	"log"

	curlconverter "github.com/mdtosif/go-curlconverter"
)

func main() {
	reqs, err := curlconverter.ParseArgs([]string{
		"curl",
		"https://example.com/items",
		"-H", "Accept: application/json",
		"--next",
		"https://example.com/submit",
		"-d", "name=codex",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(curlconverter.GenerateJavaScript(reqs[0]))
	fmt.Println(curlconverter.GenerateJavaScript(reqs[1]))
}
```

To inspect parser JSON in your own code:

```go
package main

import (
	"fmt"
	"log"

	curlconverter "github.com/mdtosif/go-curlconverter"
)

func main() {
	jsonOutput, err := curlconverter.ParseJSON(`curl https://example.com -b cookie.txt`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(jsonOutput)
}
```

Current top-level library helpers include:

- `Parse`
- `ParseWarn`
- `ParseArgs`
- `ParseArgsWarn`
- `ParseRequest`
- `ParseRequestWarn`
- `ParseRequestArgs`
- `ParseRequestArgsWarn`
- `ParseJSON`
- `ParseJSONWarn`
- `ParseJSONArgs`
- `ParseJSONArgsWarn`
- `ToJavaScript`
- `ToJavaScriptArgs`
- `ToNodeAxios`
- `ToNodeAxiosArgs`
- `ToGo`
- `ToGoArgs`
- `ToPython`
- `ToPythonArgs`
- `GenerateJavaScript`
- `GenerateNodeAxios`
- `GenerateGo`
- `GeneratePython`
- `SupportedLanguages`

The `Warn` variants currently surface parser and generator warnings for the implemented subset, including multi-request `--next`, multiple URLs, cookie-file inputs, file-backed body inputs, ignored non-`curl` commands in raw shell strings, raw-string tokenizer issues such as unterminated quotes and dangling trailing backslashes, shell-expansion markers such as `$VAR`, `${VAR}`, `$(...)`, backticks, and special Bash variables like `$?`, plus shell-structure markers such as suspicious line continuations, background operators, pipelines, and redirection operators.

## Run

```sh
cd go
go run ./cmd/curlconv "curl https://example.com -H 'foo: bar'"
```

Example output:

```js
fetch('https://example.com', {
  headers: {
    'foo': 'bar'
  }
});
```

To emit parser JSON instead of generated code:

```sh
cd go
go run ./cmd/curlconv --language parser "curl https://example.com -H 'foo: bar'"
```

## Test

```sh
cd go
go test ./...
```

If needed, run tests with a local build cache:

```sh
cd go
GOCACHE=/tmp/go-build-cache go test ./...
```

## Example

Input:

```sh
curl 'https://example.com/api' \
  -H 'content-type: application/json' \
  -b 'session=abc123' \
  --data-raw '{"ok":true}'
```

Output:

```js
fetch('https://example.com/api', {
  method: 'POST',
  headers: {
    'content-type': 'application/json',
    'Cookie': 'session=abc123'
  },
  body: '{"ok":true}'
});
```

## Scope

The Go implementation is intentionally focused on a subset of the upstream output matrix.

It does not aim to be a full cross-language port of the upstream project, but within that scope it now supports:

- common curl request parsing for headers, cookies, auth, JSON, forms, uploads, and query-string generation
- real file-backed request bodies and multipart file reads in generated code
- bearer, basic, and digest authentication handling
- stable, tested output for JavaScript `fetch()`, Node Axios, Go, and Python on the local fixture set copied into `go/test/fixtures`

It still does not provide full upstream parity for the entire generator matrix or full shell-language behavior.

## Module path

This Go module currently uses:

```txt
github.com/mdtosif/go-curlconverter
```
