Go curlconverter (MVP)

This folder contains a minimal Go implementation of a curl -> JavaScript (fetch) converter.

Structure:
- cmd/curlconv: small CLI
- pkg/parser: minimal curl parser for -X, -H, -d and URL
- pkg/request: request types
- pkg/generator/javascript: JS fetch code generator + tests

How to run tests:

cd go && go test ./...
