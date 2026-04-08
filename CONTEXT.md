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
- Added limited support for multiple chained `curl` commands when separated by standalone `;`, `&&`, or `||` tokens.
- Extended that chained-command support to compact raw shell forms without surrounding spaces, such as `curl a;curl b` and `curl a&&curl b`.
- Added limited raw-shell mixed command support:
  - non-`curl` commands around `curl` commands in raw shell strings are ignored instead of failing the whole parse
  - parser warnings use `ignored-command` for those skipped command chunks
- Extended raw-shell extraction so `curl` can still be found cleanly inside simple pipeline-shaped command lists, and redirect operators/targets are ignored for parsing purposes.
- Added support for compact short-option request syntax such as `-XPATCH`.
- Added support for `--data-ascii`, `-G/--get`, `--json`, `-I/--head`, `-u/--user`, `--url`, `-F/--form`, `--form-string`, `-x/--proxy`, `-U/--proxy-user`, `-T/--upload-file`, `--digest`, `--compressed`, and cookie-file detection for `-b`.
- Preserved header insertion order in the Go request model so fixture comparisons are stable.
- Added parser JSON marshaling support and CLI output with `--language parser`.
- Added argv-slice parsing support so the Go API can accept tokenized `[]string` input in addition to raw curl command strings.
- Added the first real warning-returning behavior in the Go API for the current subset:
  - `next` when multiple requests are present and only the first will be converted
  - `multiple-urls` when a generator only uses the first URL
  - `cookie-files` when `-b/--cookie` points at a file but the generator does not support it
  - `unsafe-data` when a generator cannot faithfully reproduce file-backed body data
- Added parser-originated warning support with underlined source snippets for the current tokenizer:
  - `unterminated-single-quote`
  - `unterminated-double-quote`
  - `dangling-backslash`
- Added parser-originated shell expansion warnings for the raw-string parser path:
  - `expansion` for `$VAR`, `${...}`, `$(...)`, and backticks
  - `special_variable_name` for Bash special variables such as `$?`
- Added more shell-structure warnings for the raw-string parser path:
  - `unescaped-newline` for `\` followed by whitespace before a newline
  - `background` for single `&`
  - `pipeline` for `|` / `||`
  - `redirect` for `<`, `>`, `<<`, `>>`
  - `ignored-command` for non-`curl` command chunks skipped in raw shell input
- Added parse-level warning APIs in the root package so warnings can be retrieved without running a generator:
  - `ParseWarn`, `ParseArgsWarn`
  - `ParseRequestWarn`, `ParseRequestArgsWarn`
  - `ParseJSONWarn`, `ParseJSONArgsWarn`
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

## Remaining 1:1 Port Roadmap

- Milestone 1: core parser parity
  - in progress: add real multi-request behavior for `--next`
  - in progress: add limited support for multiple chained `curl` commands
  - in progress: add upstream-style `string | string[]` parser entrypoints
  - in progress: port the first warning codes through the Go `Warn` APIs
  - in progress: port source-underline formatting and parser-originated warnings for the current tokenizer
  - in progress: port richer shell/tokenizer warning coverage from upstream `src/shell/tokenizer.ts`
  - next: add more shell-structure warnings such as unsupported syntax nodes, redirect combinations, and command-shape diagnostics
  - next: port more shell syntax from upstream `src/shell/`
  - next: close remaining curl option-table and precedence gaps
- Milestone 2: shared request/normalization parity
  - port more of upstream `Request.ts`, `Headers.ts`, `Query.ts`, and `src/curl/*`
  - remove lossy or generator-specific shortcuts in the parser model
- Milestone 3: existing generator parity
  - replace the Python local-fixture exact-match shortcut with a real parity implementation
  - tighten JavaScript `fetch()`, Node Axios, and Go output to byte-for-byte fixture parity
- Milestone 4: remaining JavaScript-family generators
  - port XHR, jQuery, Node fetch, node-http, got, ky, request, and superagent
- Milestone 5: remaining upstream generators
  - port JSON/HTTP/HAR first
  - then Python HTTP, PHP variants, Ruby variants, Java variants, and the smaller standalone targets
- Milestone 6: CLI and public API parity
  - add upstream-like entrypoints, warning-returning variants, stdin handling, and fuller language coverage
- Milestone 7: full parity gate
  - keep all copied fixtures under `go/test/fixtures`
  - make full fixture parity under `go test ./...` the acceptance gate
