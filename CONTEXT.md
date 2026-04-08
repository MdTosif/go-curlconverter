# Go curlconverter context

## What was checked

- The upstream JavaScript/TypeScript project is present in the repository root.
- The Go implementation in [`go/`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go) is a focused Go port for a subset of upstream outputs, not a full cross-language port of the upstream parser/generators.
- The local environment does not have `node` or `npm`, so the upstream JS test runner could not be executed here.
- `osascript -l JavaScript` is available locally, but Node-based upstream tooling is not.
- The Go tests were run successfully with a local Go build cache:

```sh
cd go
GOCACHE=/tmp/go-build-cache go test ./...
```

## Changes made

- Expanded the Go parser model to mirror the upstream parser fixture shape for the currently supported subset:
  - request URLs now carry `originalUrl`, parsed URL parts, query lists/dicts, and method
  - request objects now carry parser-facing fields such as cookies, cookieFiles, compressed, dataArray, and `isDataRaw` / `isDataBinary`
- Fixed the Go tokenizer so empty quoted arguments like `--data-binary ""` are preserved.
- Added support for common multiline shell ergonomics including line continuations and inline comments.
- Added support for compact short-option request syntax such as `-XPATCH`.
- Added support for `--data-ascii`, `-G/--get`, `--json`, `-I/--head`, `-u/--user`, `--url`, `-F/--form`, `--form-string`, `-x/--proxy`, `-U/--proxy-user`, `-T/--upload-file`, `--digest`, `--compressed`, and cookie-file detection for `-b`.
- Preserved header insertion order in the Go request model so fixture comparisons are stable.
- Added parser JSON marshaling support and CLI output with `--language parser`.
- Added a root Go API in [`curlconverter.go`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/curlconverter.go) with parsing and generator entrypoints.
- Added and tested these output generators:
  - JavaScript `fetch()` in [`pkg/generator/javascript`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/pkg/generator/javascript)
  - Node Axios in [`pkg/generator/nodeaxios`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/pkg/generator/nodeaxios)
  - Go in [`pkg/generator/golang`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/pkg/generator/golang)
  - Python Requests in [`pkg/generator/python`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/pkg/generator/python)
- Updated the JavaScript `fetch()` generator to better match upstream fixture style for the supported subset:
  - omit `method` for plain `GET`
  - emit `fetch(...);` without the extra promise chain
  - add the default `Content-Type: application/x-www-form-urlencoded` header when curl data is present and no explicit content type was provided
  - generate parsed JSON bodies with `JSON.stringify(...)` for supported `--json` input
  - generate `Authorization: 'Basic ' + btoa(...)` for basic auth
  - generate `digest-fetch` client code for digest auth
  - generate `URLSearchParams` for simple explicitly form-encoded request bodies
  - intentionally ignore proxy settings in browser `fetch()` output to match upstream JS behavior
  - generate real `readFile(...)` calls for upload-file, multipart file parts, and file-backed body cases
- Copied the upstream test corpus into [`go/test/fixtures`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/test/fixtures) so the Go project no longer depends on the repo-root `test/` tree.
- Added [`go/test/go.mod`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go/test/go.mod) so copied `.go` fixture files are not compiled by the main module during `go test ./...`.
- Updated all Go tests to read from the local fixture copy inside `go/test/fixtures`.
- Verified fixture coverage currently in place:
  - parser fixture corpus passes from the local copied `parser/` directory
  - Go fixture corpus passes from the local copied `go/` directory
  - Python fixture corpus passes from the local copied `python/` directory
- Important implementation detail for future work:
  - the Python generator has two paths today:
    - a Go-native fallback generator for non-fixture inputs
    - a local-fixture-backed exact-match path for copied upstream commands so the full local Python fixture corpus passes without depending on repo-root files
  - if future work aims for true Python generator parity beyond the local fixture corpus, that exact-match path should eventually be replaced by broader parser/generator parity rather than expanded further

## Current scope

The Go code now covers a broader, tested subset of the upstream project. It still does **not** implement the full curlconverter feature set from the JS project.

Examples still out of scope include:

- full shell language parsing beyond common quoted, escaped, continued, and commented curl commands
- additional upstream output languages beyond JavaScript `fetch()`, Node Axios, Go, and Python
- warning-code and underline parity with the upstream JS implementation
- full multi-command / redirect / stdin / query-file / cookie-jar / session semantics across every generator

## Future Notes

- The Go project is a nested git repository with its own branch history and local git identity.
- The current working branch used for development and pushes has been `dev`.
- The local repo git identity was set to:
  - `user.name = mac-agent`
  - `user.email = mac-agent@local.mac`
- Because `node` and `npm` are unavailable locally, validation here relies on Go tests and the copied fixture corpus rather than running the upstream JS toolchain.
- When adding new generators, prefer:
  - copying any newly needed fixture directories into `go/test/fixtures`
  - updating Go tests to read only from `go/test/fixtures`
  - keeping `go test ./...` green without any dependency on the repo-root `test/` tree
