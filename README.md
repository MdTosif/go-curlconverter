# Go curlconverter

This directory contains a small Go implementation of a `curl` to JavaScript `fetch()` converter.

It is an MVP inspired by the upstream `curlconverter` JavaScript project in the repository root. It is useful for basic conversions, local experiments, and incremental parity work, but it is not yet a full port of the upstream parser or generators.

## What it does

- parses a limited subset of `curl`
- builds a small request model in Go
- generates JavaScript `fetch()` code
- includes unit tests and a few fixture-backed parity tests against the checked-in upstream project fixtures

## Supported flags

The current Go parser supports these options:

- `-X`, `--request`
- `-G`, `--get`
- `-H`, `--header`
- `-b`, `--cookie`
- `-d`, `--data`
- `--data-raw`
- `--data-binary`
- `--data-ascii`
- `--json`
- URL arguments starting with `http://` or `https://`

Current behavior includes:

- default method is `GET`
- any supported data flag switches the request to `POST` when no explicit method is set
- empty quoted bodies like `--data-binary ""` are preserved
- `-b/--cookie` is converted into a `Cookie` header
- repeated data flags are joined with `&`
- `-G/--get` moves data fields into the query string
- `--json` joins multiple JSON fragments, adds `Content-Type` and `Accept`, and generates `JSON.stringify(...)` output when the JSON parses cleanly
- when a body exists and no explicit `Content-Type` header is provided, the generator adds `application/x-www-form-urlencoded`

## Project structure

- `cmd/curlconv`
  Small CLI for converting a curl command to JavaScript.
- `pkg/parser`
  Minimal curl parser for the supported subset.
- `pkg/request`
  Request model used between parsing and code generation.
- `pkg/generator/javascript`
  JavaScript `fetch()` generator and tests.
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

## Current limitations

This is still a minimal implementation. It does not yet support much of the upstream project's behavior, including:

- complex shell parsing and escaping edge cases
- multipart form handling
- file uploads
- proxy and auth options beyond raw cookie handling
- advanced JavaScript output shaping such as `URLSearchParams`
- the full upstream language/generator matrix

## Module path

This Go module currently uses:

```txt
github.com/mdtosif/go-curlconverter
```
