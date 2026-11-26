# AGENTS.md

Standing rules for anyone, human or agent, working in this repo.

## What this is

`ap` is a standard-library-only Go module of assignment-problem solvers,
generic over integer costs: `lsap` (linear sum assignment, dense and
sparse), `mcm` (maximum-cardinality bipartite matching) and `lbap`
(linear bottleneck assignment), each with a JSON-in/JSON-out CLI under
`cmd/`. The design and roadmap are in
[the design doc](docs/superpowers/specs/2026-11-25-ap-solver-suite-design.md).

## Rules

- Standard library only. Don't add `require` lines to `go.mod`.
- Library costs are integers (generic over `ap.Integer`). Only the CLIs
  accept non-integer input, through `-scale`.
- Name each package and binary by a lowercase acronym for its problem (the
  book's where one exists, else one defined in the package doc: `lsap`,
  `lbap`, `mcm`); the solver type is the uppercase acronym.
- Test first. Check every exact solver against `internal/bruteforce`.
- Cite every implementation. In short: each package's `doc.go` has the full
  reference plus the book section and algorithm number; each kernel's doc
  comment names its algorithm and source; every knowing departure from a
  source gets a comment; instance generators cite their classes; the README
  References section lists every source. Check citations against the
  original paper or the book's bibliography, never memory. Full rules:
  [Citations](docs/superpowers/specs/2026-11-25-ap-solver-suite-design.md#citations).
- Benchmarks decide which algorithms stay. Don't keep an algorithm that
  doesn't earn its place.
- Follow the
  [Shared conventions](docs/superpowers/specs/2026-11-25-ap-solver-suite-design.md#shared-conventions-all-solvers)
  for package and type shape, input validation, internal packages and CLIs.
  They live in the design doc, not here.

## Checks

Run all four before every commit:

```
gofmt -l .        # must print nothing
go vet ./...
go test -race ./...
GOTOOLCHAIN=go1.27.4 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...
```

## Releases

- One milestone per release. Each deliverable is an issue with its own PR.
- The CHANGELOG lists issue numbers.
- Squash-merge PRs.
- Tag the merge commit with a lightweight `vX.Y.Z` tag.
- Never merge a PR or push a tag without the owner's go-ahead for that
  specific PR or tag. See
  [Project management](docs/superpowers/specs/2026-11-25-ap-solver-suite-design.md#project-management-github).
