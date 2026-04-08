# Go curlconverter context

## What was checked

- The upstream JavaScript/TypeScript project is present in the repository root.
- The Go implementation in [`go/`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go) is a focused Go port for JavaScript `fetch()` generation, not a full cross-language port of the upstream parser/generators.
- The local environment does not have `node` or `npm`, so the upstream JS test runner could not be executed here.
- The Go tests were run successfully with a local Go build cache:

```sh
cd go
GOCACHE=/tmp/go-build-cache go test ./...
```

## Changes made

- Fixed the Go tokenizer so empty quoted arguments like `--data-binary ""` are preserved.
- Added support for `--data-ascii` in the Go parser.
- Added support for `-G/--get` query merging with repeated data flags.
- Added support for `--json`, including default JSON headers.
- Added support for `-I/--head`, `-u/--user`, and `--url`.
- Added support for `-F/--form` and `--form-string`, including placeholder file parts.
- Added support for parsing proxy flags `-x/--proxy` and `-U/--proxy-user`.
- Added support for `-T/--upload-file` and file-backed request bodies such as `--data-binary @file`.
- Added support for `--digest`, generating `digest-fetch` client code for browser-style JavaScript parity.
- Added support for common multiline shell ergonomics including line continuations and inline comments.
- Preserved header insertion order in the Go request model so fixture comparisons are stable.
- Updated the Go JavaScript generator to better match upstream fixture style for the supported subset:
  - omit `method` for plain `GET`
  - emit `fetch(...);` without the extra promise chain
  - add the default `Content-Type: application/x-www-form-urlencoded` header when curl data is present and no explicit content type was provided
  - generate parsed JSON bodies with `JSON.stringify(...)` for supported `--json` input
  - generate `Authorization: 'Basic ' + btoa(...)` for basic auth
  - generate `digest-fetch` client code for digest auth
  - generate `URLSearchParams` for simple explicitly form-encoded request bodies
  - generate `FormData()` and placeholder `File(...)` values for supported multipart forms
  - intentionally ignore proxy settings in browser `fetch()` output to match upstream JS behavior
  - generate real `readFile(...)` calls for upload-file, multipart file parts, and file-backed body cases
- Added Go tests that compare generated JavaScript against selected checked-in upstream fixture outputs.
- Added parser tests for empty body handling, `--data-ascii`, `-G`, `--json`, `-I`, `-u`, `--url`, and multipart forms.

## Current scope

The Go code now covers a broader, tested `curl` to JavaScript `fetch()` surface. It still does **not** implement the full curlconverter feature set from the JS project.

Examples still out of scope include:

- full shell language parsing beyond common quoted, escaped, continued, and commented curl commands
- additional upstream output languages beyond the Go project's JavaScript `fetch()` target
