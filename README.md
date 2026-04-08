# Go curlconverter

This directory contains a Go implementation of curl request conversion with Go-native parser and generator packages.

It is a focused Go port inspired by the upstream `curlconverter` JavaScript project in the repository root. The Go code targets a growing subset of upstream outputs, with unit tests and fixture-backed parity checks for the supported generators.

## What it does

- parses a limited subset of `curl`
- builds a request model in Go
- generates JavaScript `fetch()`, Node Axios, Go, and Python code
- includes unit tests and fixture-backed parity tests against the checked-in upstream project fixtures

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
- empty quoted bodies like `--data-binary ""` are preserved
- `-b/--cookie` is converted into a `Cookie` header
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

## Project structure

- `cmd/curlconv`
  Small CLI for converting a curl command to supported output targets.
- `pkg/parser`
  Curl parser for the supported subset.
- `pkg/request`
  Request model used between parsing and code generation.
- `pkg/generator/javascript`
  JavaScript `fetch()` generator and tests.
- `pkg/generator/nodeaxios`
  Node Axios generator and tests.
- `pkg/generator/golang`
  Go generator and tests.
- `pkg/generator/python`
  Python Requests generator and tests.
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
- stable, tested output for JavaScript `fetch()`, Node Axios, Go, and Python on the supported fixture set

## Module path

This Go module currently uses:

```txt
github.com/mdtosif/go-curlconverter
```
