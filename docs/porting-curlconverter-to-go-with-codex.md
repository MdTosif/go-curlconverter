# Porting curlconverter to Go with Codex and TDD

The most useful part of this Go port was not just writing code faster. It was turning an already mature project into a test-guided workflow where each small piece of Go behavior could be verified against something real.

This repository already had a big advantage: the original `curlconverter` project came with a large fixture corpus and lots of expected outputs across languages. Instead of treating the Go port as a blank-slate rewrite, I used that existing test surface as the specification. That changed the job from "invent the behavior" to "make the Go implementation prove it matches known behavior."

## What I was building

Inside [`go/`](https://github.com/MdTosif/go-curlconverter/go) I built a Go-native version of `curlconverter` with:

- a parser for a growing subset of `curl`
- a shared Go request model
- generators for multiple output targets
- warning-returning APIs and a small CLI

The interesting engineering challenge was not only parsing flags like `-d`, `-G`, `--json`, `-F`, `-u`, or `-T`. It was preserving behavior closely enough that the generated output could be compared with checked-in fixtures from the upstream project.

## How Codex fit into the workflow

I used Codex as a pair programmer, not as a replacement for the feedback loop.

Codex helped most in four areas:

- reading the existing TypeScript project structure and mapping it to Go packages
- sketching parser and generator changes quickly once a failing test made the target behavior clear
- tracing gaps between fixture expectations and the current Go output
- helping keep the public API, parser behavior, and generator behavior aligned as the project grew

That mattered because this port spans a lot of surface area. There are parser entrypoints, warning APIs, fixture-backed generators, copied test fixtures, and a nested Go module structure. Codex was useful for navigating that breadth quickly, but the guardrails still came from tests.

## Why the existing tests were so valuable

The best part of working in this repo was that I did not have to guess what "correct" meant.

The project already had a strong testing shape:

- targeted parser tests for edge cases in [`go/pkg/parser/parser_test.go`](https://github.com/MdTosif/go-curlconverter/go/pkg/parser/parser_test.go)
- parser fixture parity tests in [`go/pkg/parser/parser_fixtures_test.go`](https://github.com/MdTosif/go-curlconverter/go/pkg/parser/parser_fixtures_test.go)
- generator fixture parity tests such as [`go/pkg/generator/golang/golang_test.go`](https://github.com/MdTosif/go-curlconverter/go/pkg/generator/golang/golang_test.go)
- public API coverage in [`go/curlconverter_test.go`](https://github.com/MdTosif/go-curlconverter/go/curlconverter_test.go)

That structure naturally supported TDD.

For parser work, I could start with a focused failing test. For example, tests around empty `--data-binary`, cookie handling, repeated `-d` with `-G`, JSON body merging, and warning behavior made it easy to encode one rule at a time. The test described the contract first, and the implementation only needed to satisfy that contract.

For generator work, the fixture corpus was even more powerful. Instead of asserting a couple of strings manually, the test suite already compared generated Go output against checked-in expected files. In [`go/pkg/generator/golang/golang_test.go`](https://github.com/MdTosif/go-curlconverter/go/pkg/generator/golang/golang_test.go), each fixture test reads a curl command, parses it, generates Go code, and compares the result to the expected `.go` fixture. That is a very practical TDD loop:

1. pick a failing fixture
2. understand the delta between actual and expected output
3. change the generator or parser
4. rerun tests
5. stop when the fixture passes without breaking others

## How the fixture corpus helped me write Go in a TDD style

The existing project test cases helped in a few concrete ways.

First, they reduced ambiguity. A curl converter has many tiny behavioral rules: header ordering, method inference, multipart handling, JSON normalization, body precedence, warnings, and shell quirks. The fixtures gave me examples of what the project already considered correct.

Second, they encouraged small increments. I did not need to "finish the parser" before getting feedback. I could add support for one flag, one warning, or one command-shape at a time and let the tests tell me whether I had broken parity.

Third, they separated parser bugs from generator bugs. If parser JSON fixtures failed, I knew the request model was wrong. If parser tests passed but generator fixtures failed, I knew the bug was in rendering logic. That kept debugging disciplined.

Fourth, they made Codex more effective. AI is much better when it has a crisp target. A vague prompt like "port curl parsing to Go" is broad. A concrete task like "make this failing fixture pass without regressing the parser warning tests" is much sharper. The test suite supplied that sharpness.

## A practical example of the loop

One pattern that came up repeatedly looked like this:

1. I found a missing behavior, such as a shell edge case or a curl flag that the Go parser did not yet support.
2. I wrote or used an existing failing test to capture the expected behavior.
3. I used Codex to inspect nearby code, compare it with the upstream implementation, and suggest the smallest reasonable Go change.
4. I updated the Go implementation.
5. I reran the test suite until the new case passed and the broader fixture-backed suite stayed green.

That is still TDD, even with AI in the loop. Codex accelerated the code-writing and repo navigation, but the tests remained the source of truth.

## What made this feel reliable

Two things made the workflow feel dependable.

One was the combination of narrow unit tests and broad fixture tests. Unit tests caught local logic mistakes early. Fixture tests protected overall parity.

The other was keeping the fixtures local to the Go project under [`go/test/fixtures`](https://github.com/MdTosif/go-curlconverter/go/test/fixtures). That meant the Go module could run its own validation path with `go test ./...` and keep the acceptance criteria close to the implementation.

## Final takeaway

The biggest lesson from this port is that Codex worked best when paired with an existing test culture.

Codex helped me move faster through a large and repetitive porting task, but the repository's existing test cases are what made the work disciplined. They gave me a specification, a red-green-refactor loop, and confidence that the Go implementation was moving toward parity instead of just accumulating code.

If I had to summarize the approach in one line, it would be this: Codex helped me write the Go code, but the project's existing tests taught me how to write it in a TDD way.
