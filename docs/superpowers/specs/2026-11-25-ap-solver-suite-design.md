# `ap` solver suite: design

Date: 2026-11-25
Status: approved 2026-11-25
Scope: the overall roadmap, plus a detailed design for the first milestone (v0.5).

## Intent

`ap` becomes a repository of high-performance solvers for different types of
assignment problem. Each problem type can be used in two ways:

- **Library:** a Go package with the same shape and usability as `lsap`.
  Build it from cost data, solve it, query the result, and optionally modify
  it and re-solve.
- **Binary:** a single, dependency-free executable that reads JSON on stdin
  and writes JSON on stdout, like `cmd/lsap` today.

Burkard, Dell'Amico & Martello, *Assignment Problems* (SIAM 2012), is the map
of the territory: problem types, standard algorithms, and benchmark instance
classes. It is not a spec. Implementations don't have to follow the book's
algorithms or chapter structure.

## Goals

- Solve large problems routinely: thousands of items on each side, or more.
- Benchmarks decide what counts as high-performance. A solver or algorithm
  earns its place by benchmark results on some class of instances.
- Dense and sparse input from the start, for every problem type where sparse
  input is meaningful.
- The same usability across packages. Someone who knows `lsap` can pick up
  any other solver without reading its docs first.
- Exact results. Every exact solver is checked against brute force on small
  instances.
- Every implementation cites its source (see Citations).

## Non-goals

- Reproducing the book's taxonomy or implementing every algorithm in it.
- Floating-point costs in the libraries (see Constraints).
- Parallel or GPU solvers for now.
- Third-party dependencies, in the libraries or the binaries.

## Anti-goals (failure even if it technically works)

- A solver that silently returns a wrong or infeasible answer. This includes
  the old `lsap` behavior of returning an assignment built from big-M edges
  when no real assignment exists.
- Packages that each feel different to use.
- Generic abstractions that slow down the inner loops.
- A breaking change to the JSON formats or APIs because sparse input was
  added later. That's the reason sparse input is in v0.5.

## Constraints

- Go standard library only. The module requires Go 1.27, as of v0.4.0.
- **Integer costs only in the libraries.** Solvers are generic over
  `ap.Integer` (signed integer types). Exact integer arithmetic avoids the
  tolerance bugs common in floating-point assignment codes, such as cycling
  and wrong optimality checks.
- **The CLIs accept floats through `-scale`.** With `-scale k`, each input cost
  is multiplied by `k` and rounded to the nearest `int64`. Without `-scale`,
  any non-integer cost is an error whose message suggests `-scale`.
- **`lsap` stays incremental.** Its value is incremental re-solving: remove
  an edge and re-optimize with a single augmentation (O(n²) dense), plus a
  cheap `Copy`, which is what tree search and reduced-cost filtering need.
  The **dense kernel keeps the current Carpaneto–Martello–Toth code**. The
  **sparse kernel** is a new implementation in the same primal-dual shortest
  augmenting path family with the same incremental behavior (see v0.5
  section 2). A second dense LSAP algorithm (likely Jonker–Volgenant) is added
  only if benchmarks show the current code losing on one-shot solves of large
  problems.
- Pre-1.0 semantic versioning. Breaking changes are allowed in minor
  releases and listed in the CHANGELOG.

## Roadmap

Each milestone is a GitHub milestone and a release.

| Milestone | Contents |
|---|---|
| **v0.5** | `ap`/`lsap` API cleanup; sparse `lsap`; shared internal plumbing and CLI conventions; `mcm` (bipartite matching); `lbap` (bottleneck assignment, closes #2); benchmark instance generators. |
| v0.6 | k-best assignments (Murty's ranking), built on incremental `lsap`. |
| v0.7 | Quadratic assignment (QAP): strong heuristics for large instances, plus an exact branch-and-bound with the Gilmore–Lawler bound for small instances (n ≤ about 30). Benchmarked against QAPLIB's best-known solutions. |
| Later, as wanted | Rectangular and k-cardinality LSAP; a one-shot LSAP algorithm if benchmarks call for it; balanced and lexicographic bottleneck assignment; semi-assignment; axial 3-index assignment. |

Only v0.5 is designed in detail below. Each later milestone gets its own
design when it starts.

## Approaches considered (how the solvers share code)

- **A. Fully independent packages.** Each solver owns its dense and sparse
  data structures, and only result types are shared through `ap`. Simplest,
  and each hot loop is tuned for its solver, but graph building, validation
  and JSON handling get written three times or more.
- **B. One shared cost abstraction.** Solvers are written once, generic over a
  cost-lookup interface that dense and sparse inputs both satisfy. Least code,
  but Go compiles that into an indirect call for every cost lookup in the
  innermost loops, which is likely to show up badly in benchmarks.
- **C. Shared plumbing, specialized kernels (chosen).** Public types live in
  `ap`. A private `internal/` tree holds everything outside the hot loops:
  building sparse graphs, validation, overflow checks, JSON reading and
  writing, and `-scale`. Each solver writes its own dense and sparse kernels
  over plain slices. This avoids A's duplication without B's cost in the inner
  loops.

## Shared conventions (all solvers)

### Package and type shape

Each problem type has a package (`lsap`, `mcm`, `lbap`) and a binary
under `cmd/<package>`. Each package provides:

- **Names:** each package and its binary are named by a lowercase acronym
  for the problem: the book's acronym where one exists (`lsap`, `lbap`,
  `qap`, `lexbap`), otherwise one we define in the package doc (`mcm`,
  maximum cardinality matching). The solver type is the uppercase acronym
  (`lsap.LSAP`, `lbap.LBAP`, `mcm.MCM`). Go names can't start with a digit,
  so a 3-index problem needs a name like `ap3`.
- `New(...)` for dense input and `NewSparse(...)` for sparse input. Both
  return `(*Concrete, error)`. They return a concrete pointer type, not an
  interface, so methods can be added later without breaking callers.
- The constructors **copy their input**. Callers' slices are never modified.
- A solve method. For assignment problems this is
  `Assign() (ap.Permutation, error)`, which returns `ap.ErrInfeasible` when no
  feasible assignment exists.
- Result accessors that are valid after a successful solve (`Cost()` and
  others).
- `Remove(u, v int)` forbids an edge for later solves. It's a no-op if the
  edge doesn't exist or is already removed.
- `Copy()` returns an independent solver, including its current solution
  state, so the next re-solve after `Remove` stays incremental.
- Solvers aren't safe for concurrent use. Use `Copy` to work in parallel.
- After a solve returns an error, the solver's results are undefined. The
  solver can still be copied or discarded, but further solves aren't
  guaranteed to work. *(Builder's call whether to make it reusable.)*

### Public types in `ap`

Changes to the root package:

```go
// Edge is an allowed assignment of u ∈ U to v ∈ V with a cost.
type Edge[T Integer] struct {
	U, V int
	Cost T
}

// Pair is an allowed assignment of u ∈ U to v ∈ V without a cost.
type Pair struct {
	U, V int
}

// ErrInfeasible means no feasible assignment exists.
var ErrInfeasible = errors.New("ap: no feasible assignment")
```

**`ap` holds shared types, not interfaces.** The interfaces `Assigner`,
`Coster[T]`, `DualPricer[T]` and `ReducedCoster[T]` are removed. Nothing
consumes them, and `Cost()` means different things in different solvers (a
sum in `lsap`, a maximum in `lbap`). Go interfaces are satisfied implicitly,
so any caller that wants one can declare it at the point of use. Internal
code (CLI plumbing, test helpers) declares its own small unexported
interfaces the same way.

`MinOf` and `MaxOf` move to an internal package (e.g. `internal/intmath`),
shared by the solvers and the overflow checks. Nothing outside the module
needs them once `ReducedCost` uses the comma-ok form (below). The internal
version should compute the bounds without `unsafe` if that's
straightforward.

After v0.5, `ap` contains exactly: `Permutation`, `Cycles`, `Matrix`,
`DualPrices[T]`, `Edge[T]`, `Pair`, `Integer` and `ErrInfeasible`.

**Consistency is enforced by tests, not interfaces.** An internal test checks
that every solver has the shared method set with the agreed signatures
(`Assign() (ap.Permutation, error)`, `Cost() T` where the problem has costs,
`Remove(u, v int)`, and `Copy()` returning its own concrete type), so the
packages can't drift apart.

### Citations

Every implementation cites the source it implements, following the pattern
`lsap/doc.go` already uses:

- **Package doc** (`doc.go`): the full reference for each algorithm the
  package implements (authors, title, venue, volume, pages, year), plus the
  section and algorithm number in Burkard, Dell'Amico & Martello where there
  is one.
- **Kernel doc comments:** each solver kernel or major procedure (e.g. the
  dense and sparse `lsap` path searches) names its algorithm and source.
- **Departures:** wherever the code knowingly departs from its source, a
  comment says how and why. The existing `p[r] = n` note in
  `lsap/initialize.go` is the model.
- **Instance generators and benchmark classes** cite where each instance class
  comes from.
- **README:** a References section lists every source cited in the module.
- Citations are checked against the original paper, or against the book's
  bibliography if the paper isn't available, never written from memory. For
  example, the current `lsap/doc.go` gives pages 191–223 for
  Carpaneto–Martello–Toth (1988), but the book's bibliography entry [169]
  lists pages starting at 193. Check that one as part of v0.5.

Likely sources for v0.5, as listed in the book's bibliography (to be
confirmed against the code actually written):

| Implementation | Source |
|---|---|
| dense `lsap` | G. Carpaneto, S. Martello, P. Toth, "Algorithms and codes for the assignment problem," *Ann. Oper. Res.* 13 (1988) [book ref 169]; book §4.4 |
| sparse `lsap` | R. Jonker, A. Volgenant, "A shortest augmenting path algorithm for dense and sparse linear assignment problems," *Computing* 38:325–340 (1987) [book ref 408]; book §4.4, Algorithms 4.9–4.10. Cite whichever source the search actually follows. |
| `mcm` | J. Hopcroft, R.M. Karp, "An n^5/2 algorithm for maximum matchings in bipartite graphs," *SIAM J. Comput.* 2:225–231 (1973) [book ref 389]; book §3.3, Algorithm 3.2 |
| `lbap` threshold | R. Garfinkel, "An improved algorithm for the bottleneck assignment problem," *Oper. Res.* 19:1747–1751 (1971) [book ref 313]; book §6.2.2, Algorithm 6.1 |
| Machol–Wien instances | R.E. Machol, M. Wien, "A hard assignment problem," *Oper. Res.* 24:190–192 (1976), and Errata, *Oper. Res.* 25:364 (1977) [book refs 488–489]; book §4.10.1 |
| other instance classes | book §4.10.1 and the sources it cites |

### Input validation

Constructors return descriptive errors, never panic, for:

- an empty problem (n = 0);
- a dense matrix that isn't square (where the problem requires it) or has
  ragged rows;
- a sparse edge with an endpoint out of range;
- a duplicate sparse edge (same `U, V`);
- costs that can't be represented in `T`. Each constructor documents its
  exact rule and tests it at the boundary. `Assign` returns `ap.ErrOverflow`
  when the solve would overflow `T`.

### Internal packages (`internal/`)

The exact layout is the builder's call. It needs to cover:

- building a compressed sparse row (CSR) adjacency structure from
  `[]ap.Edge[T]` or `[]ap.Pair`, with validation;
- cost-range and overflow checks;
- CLI plumbing: flag conventions, reading dense or sparse JSON, `-scale`
  handling, writing results, error and exit-code handling;
- benchmark instance generators (see Benchmarks).

Hot-loop code stays in each solver package.

### CLI conventions

Every binary:

- reads one JSON problem from stdin and writes one JSON object to stdout;
- **detects the input format automatically:** a top-level array is a dense
  matrix; a top-level object is a sparse problem;
- writes a `"status"` field of `"optimal"` or `"infeasible"`. Both exit with
  code 0, because infeasible is a valid answer and not an error;
- on invalid input, writes a message to stderr and exits with code 1, with
  no stdout output and no panic;
- supports `-h`; binaries whose problems have costs also support `-scale k`;
  binaries that return a permutation support `-cycles` and `-matrix`. A flag
  a binary doesn't support is a usage error (exit 1), which is Go's `flag`
  default;
- with `-scale`, reports costs, dual prices and reduced costs in the
  original units (divided by `k`, as JSON numbers). The solution is exact for
  the scaled integer problem;
- keeps its logic in a testable `run(args, stdin, stdout, stderr) int`
  function, with `main` as a thin wrapper. CLIs get tests, which they have
  none of today.

**Sparse JSON** uses compact arrays, so million-edge inputs stay small:

```json
{ "n": 3, "edges": [[0, 1, 15], [0, 2, 12], [1, 2, 23], [2, 0, 11]] }
```

Each edge is `[u, v, cost]`. Matching uses `"nu"` and `"nv"` instead of
`"n"`, and `[u, v]` edges with no cost.

## v0.5 detailed design

### 1. `lsap` API cleanup (breaking)

```go
func New[T ap.Integer](costs [][]T) (*LSAP[T], error)
func NewSparse[T ap.Integer](n int, edges []ap.Edge[T]) (*LSAP[T], error)

func (a *LSAP[T]) Assign() (ap.Permutation, error)
func (a *LSAP[T]) Cost() T
func (a *LSAP[T]) DualPrices() ap.DualPrices[T]
func (a *LSAP[T]) ReducedCost(u, v int) (T, bool)
func (a *LSAP[T]) Remove(u, v int)
func (a *LSAP[T]) Copy() *LSAP[T]
```

- `M()` and `SetM()` are removed. Big-M, if it's still used, becomes an
  internal detail of the dense kernel.
- The `lsap.Assigner[T]` interface is removed in favor of the concrete type.
- **Infeasibility.** `Assign` returns `ap.ErrInfeasible` when the remaining
  edges, after removals or because of sparsity, admit no perfect assignment.
  A removed or missing edge never appears in a returned assignment.
- **Big-M soundness (dense).** If the dense kernel keeps big-M for removed
  edges, validation in `New` must guarantee that the optimum uses an M edge
  only when no assignment exists without one. For example, require
  n · (max cost − min cost) < M. Overflow in the solver is detected, not
  ruled out up front: `Assign` returns `ap.ErrOverflow` instead of a wrong
  answer.
  After solving, any M edge in the assignment means `ErrInfeasible`.
- `ReducedCost(u, v)` uses the comma-ok form: it returns `ok == false` for a
  removed or nonexistent edge, the way a map lookup does, so callers never
  compare against a sentinel value.
- **`Copy` performance.** `Copy` is on the hot path in tree search. It
  shouldn't duplicate the full cost data if that can be avoided, for example
  by sharing immutable costs and copying only removals and per-solve state.
  The approach is the builder's call, and it must be measured with the
  existing `Copy` benchmarks.
- The existing tests (`TestAssigner`, `Example`, the brute-force and
  determinism tests) are updated to the new API, and their expected values
  stay the same. The one exception is reduced costs on removed edges in
  `TestAssigner`, which become `ok == false` checks instead of big-M
  values.

### 2. Sparse `lsap`

- **Target instances:** large and genuinely sparse, e.g. n ≈ 100,000 with
  average degree ≈ 10 (about 1 million edges). Porting the dense code by
  scanning adjacency lists isn't enough: `path()` still makes several passes
  over all n columns per labelled row, so each augmentation stays O(n²).
- **Algorithm:** a shortest augmenting path search with a binary heap
  (Dijkstra-style) over reduced costs, touching only the adjacency lists of
  labelled rows, with dual updates deferred to the end of each search
  (standard for sparse primal-dual codes, e.g. the book's section 4.4). Same
  dual feasibility and complementary slackness invariants as the dense code,
  so `DualPrices`, `ReducedCost`, `Remove` and incremental re-solve behave
  the same. A cheap initial heuristic (e.g. column and row reduction
  restricted to existing edges) is the builder's call.
- Ties are broken deterministically (for example by lowest index), so the
  same input always gives the same assignment.
- When the search finds no augmenting path (the heap empties before reaching
  an unassigned column), `Assign` returns `ap.ErrInfeasible`.
- **Tests:** dense and sparse kernels give the same optimal cost on random
  instances (the dense matrix converted to a complete edge list, and random
  sparse subsets of it); brute-force checks on small sparse instances,
  including infeasible ones; `Remove` and `Copy` on sparse instances.

### 3. `mcm`: maximum-cardinality bipartite matching

```go
func New(adjacency [][]bool) (*MCM, error)       // dense, nU × nV
func NewSparse(nU, nV int, edges []ap.Pair) (*MCM, error)

func (m *MCM) Match() []int                   // m[u] = v, or -1 if unmatched
func (m *MCM) Size() int                      // matching cardinality
func (m *MCM) Assign() (ap.Permutation, error) // perfect matching; nU == nV required
func (m *MCM) Remove(u, v int)
func (m *MCM) Copy() *MCM
```

- Works on rectangular graphs (nU ≠ nV). `Assign` returns
  `ap.ErrInfeasible` if the graph isn't square or has no perfect matching.
- Default algorithm: Hopcroft–Karp with a greedy initial matching. Removing
  a matched edge unmatches it, and the next solve repairs the matching
  incrementally rather than starting over.
- The dense kernel's representation (for example bitsets or conversion to
  CSR) is the builder's call, decided by benchmark.
- **CLI `cmd/mcm`:**
  - Dense input is a JSON array of arrays of booleans, `nU` rows of `nV`.
    Sparse input is `{"nu": .., "nv": .., "edges": [[u, v], ...]}`.
  - Output is `{"status": "optimal", "size": k, "perfect": bool,
    "matching": [...]}`. `"matching"` is the `Match()` array: one entry per
    u ∈ U, the matched v or -1. A maximum matching always exists, so status
    is always `"optimal"`; `"perfect"` says whether it's a perfect matching
    (nU == nV and size == nU).
  - `-cycles` and `-matrix` add `"cycles"` and `"matrix"` fields only when
    the matching is perfect. Otherwise those fields are omitted, the output
    is still valid, and the exit code is 0.
  - No `-scale` flag (no costs).
- **Tests:** brute-force maximum matching on small random graphs; square and
  rectangular graphs; `Remove` and `Copy`.

### 4. `lbap`: linear bottleneck assignment (closes #2)

Minimize the largest cost used in a perfect assignment.

```go
func New[T ap.Integer](costs [][]T) (*LBAP[T], error)
func NewSparse[T ap.Integer](n int, edges []ap.Edge[T]) (*LBAP[T], error)

func (a *LBAP[T]) Assign() (ap.Permutation, error)
func (a *LBAP[T]) Cost() T // the bottleneck value
func (a *LBAP[T]) Remove(u, v int)
func (a *LBAP[T]) Copy() *LBAP[T]
```

- First algorithm: a threshold method, i.e. binary search over the sorted
  distinct costs, checking for a perfect matching among edges at or below the
  threshold. It reuses `mcm` (Hopcroft–Karp), warm-starting each check
  from the previous matching. An augmenting-path method (book Algorithm 6.3)
  is added only if benchmarks show the threshold method losing.
- Bottleneck problems have many optimal assignments. Which one is returned
  is deterministic but otherwise unspecified. Minimizing the sum among
  bottleneck-optimal assignments is a later "lexicographic" feature, not
  v0.5.
- There are no dual prices or reduced costs in v0.5.
- **CLI `cmd/lbap`:** output is `{"status", "cost", "permutation"}`, plus
  `-cycles` and `-matrix`.
- **Tests:** brute force on small dense and sparse instances, including
  infeasible ones; `Remove` and `Copy`.

### 5. `cmd/lsap` updates

- Uses the shared CLI plumbing: sparse input, `-scale`, `"status"`, errors to
  stderr, and a testable `run`.
- Keeps `-dual`, `-rc`, `-cycles` and `-matrix`.
- **Breaking change:** output gains `"status"`. With `-rc` on sparse input,
  the output is a list of `[u, v, rc]` for existing edges, not a dense matrix.

### 6. Benchmarks

- **Instance generators** in `internal/`, deterministic from a seed, for the
  book's section 4.10 classes: uniform random, geometric, two-cost, and
  Machol–Wien (c_ij = i·j). Each comes in dense and sparse versions, with
  sparse instances generated so a perfect matching is guaranteed (or not, for
  infeasibility tests).
- **Go benchmarks** for every solver and kernel at sizes that keep
  `go test -bench .` practical. Benchmark setup must not allocate huge inputs
  for each iteration (the current `BenchmarkSolve10K` builds an 800 MB matrix
  per iteration).
- **External comparison**, run by hand and not in CI. `lsap` is compared with
  SciPy's `linear_sum_assignment` and OR-Tools' linear assignment solver on
  dense n = 1,000 and 5,000, and on sparse instances of about 1 million edges
  (n ≈ 100,000, average degree ≈ 10).
  Results are recorded in a markdown file in the repo, with machine, versions
  and commands. If `lsap` loses badly on one-shot dense solves, a "one-shot
  LSAP algorithm" issue is filed for a later milestone. If sparse `lsap`
  loses badly, a follow-up issue is filed too. Neither blocks v0.5, as long as
  the results are recorded honestly. The reference machine is the project
  owner's development machine. The external tools are installed in an
  isolated Python virtual environment at their current releases, with exact
  versions recorded alongside the results. Scripts for the
  external tools may use Python; they live outside the Go module's packages
  and are never needed to build or test.

### 7. Docs and release

- The README covers all three solvers, with library and CLI examples and the
  JSON formats.
- CHANGELOG v0.5.0 lists every breaking change with a short migration note
  (`New` returns an error, `Assign` returns an error, `M`/`SetM` removed,
  `ap` and `lsap` interfaces removed, `MinOf`/`MaxOf` made internal,
  `ReducedCost` returns `(T, bool)`, `"status"` added to CLI output).
- Runnable `Example` functions for each package.
- A References section in the README (see Citations).

## Project management (GitHub)

All GitHub setup below is done by Claude through `gh`, using the owner's
access, once this design is approved.

- Close the `v0.4` milestone as released. Move issue #2 (LBAP) to `v0.5`.
- Create `v0.5` issues, one per deliverable, each with its own PR. CHANGELOG
  entries reference issue numbers.
  1. `ap`/`lsap` API cleanup
  2. internal plumbing and CLI conventions (`-scale`, format detection,
     status, errors, testable `run`)
  3. sparse `lsap`
  4. `mcm` package and CLI
  5. `lbap` package and CLI (#2)
  6. benchmark instance generators and external comparison
- Create `v0.6` and `v0.7` milestones, each with one tracking issue.
- This design doc is committed to the repo with the first v0.5 PR.
- That PR also adds a short `AGENTS.md` with the standing rules for anyone,
  human or agent, working in the repo. Those are the rules that otherwise
  live only in this conversation:
  - standard library only; integer costs in the libraries;
  - test first, and check exact solvers against brute force;
  - cite every implementation (see Citations);
  - benchmarks decide which algorithms stay;
  - the shared package and CLI conventions, pointing to this doc rather than
    repeating it;
  - the check commands: `gofmt`, `go vet`, `go test -race ./...`, and
    golangci-lint v2.15;
  - the release process: milestone, CHANGELOG with issue numbers, squash
    merge, lightweight `vX.Y.Z` tag on the merge commit.

  A one-line `CLAUDE.md` containing `@AGENTS.md` makes Claude Code load the
  same file, so there's one source of truth.

## Decided with the project owner

- Purpose: high-performance solvers, usable as libraries and as standalone
  JSON CLIs. The book is a guide, not a spec.
- Order: matching, bottleneck, k-best, QAP, then niche types.
- Integer costs in the libraries; the CLIs accept floats through `-scale`.
- Benchmarks decide performance. Thousands of items per side should be
  routine.
- The current `lsap` algorithm stays for its incremental re-solve.
- Sparse and dense input from the start.
- The `lsap` API cleanup, as listed in section 1, including making M
  internal.
- Code sharing: approach C.
- GitHub milestones per release, issues per deliverable.
- Every implementation cites its source.
- `ap` keeps shared types only. Its interfaces are removed, `MinOf` and
  `MaxOf` become internal, and `ReducedCost` uses the comma-ok form.
- Package and binary names are acronyms (`mcm` for bipartite matching).

## Calls made by Claude (for the owner to check)

- Citation placement and format (see Citations), modeled on `lsap/doc.go`.
- The sparse `lsap` kernel is a new heap-based shortest augmenting path
  search, not a port of the dense code, so that it can handle n ≈ 100,000
  sparse instances. The dense kernel keeps the current code.
- Sparse benchmark instances: n ≈ 100,000, average degree ≈ 10.
- A benchmark loss against SciPy or OR-Tools becomes a follow-up issue and
  doesn't block v0.5.
- The `cmd/mcm` JSON format and its `"perfect"` field.
- `"status"` on every CLI output; infeasible is exit code 0.
- Sparse JSON edges as compact arrays, e.g. `[u, v, cost]`.
- `ReducedCost` returns `(T, bool)`, with `false` for removed or missing
  edges.
- Package and binary names are acronyms (`mcm` for bipartite matching).
- The LBAP threshold method first, reusing `mcm`.

## Left to the builder

- The exact `internal/` package layout.
- Dense kernel representations (big-M or a forbidden-edge set for `lsap`;
  bitsets or CSR for `mcm`), decided by benchmark.
- How `Copy` avoids duplicating cost data.
- The exact overflow bounds, as long as each is documented and tested.
- Whether a solver stays reusable after returning an error.
- The sparse `lsap` initial heuristic.
- The order of PRs within v0.5, provided each PR leaves `master` releasable.
