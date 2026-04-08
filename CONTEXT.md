# Go curlconverter context

## What was checked

- The upstream JavaScript/TypeScript project is present in the repository root.
- The Go implementation in [`go/`](/Users/tofiquem/tosif-practice/go-curl/curlconverter/go) is currently an MVP, not a full port of the upstream parser/generators.
- The local environment does not have `node` or `npm`, so the upstream JS test runner could not be executed here.
- The Go tests were run successfully with a local Go build cache:

```sh
cd go
GOCACHE=/tmp/go-build-cache go test ./...
```

## Changes made

- Fixed the Go tokenizer so empty quoted arguments like `--data-binary ""` are preserved.
- Added support for `--data-ascii` in the Go parser.
- Preserved header insertion order in the Go request model so fixture comparisons are stable.
- Updated the Go JavaScript generator to better match upstream fixture style for the supported subset:
  - omit `method` for plain `GET`
  - emit `fetch(...);` without the extra promise chain
  - add the default `Content-Type: application/x-www-form-urlencoded` header when curl data is present and no explicit content type was provided
- Added Go tests that compare generated JavaScript against selected checked-in upstream fixture outputs.
- Added parser tests for empty body handling and `--data-ascii`.

## Current scope

The Go code now matches a small supported subset of the upstream JavaScript fixture behavior. It still does **not** implement the full curlconverter feature set from the JS project.

Examples still out of scope include:

- complex shell parsing
- `-G` query merging
- multipart/form handling
- auth/proxy/cookie options
- advanced JavaScript output shaping like `URLSearchParams` and JSON body inference
