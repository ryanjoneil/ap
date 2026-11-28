# k-best assignments (`klsap`) and `lsap.Fix`: design

Issues #24 (k-best) and #38 (`Fix`). Milestone v0.6.

## Why

The roadmap puts k-best assignments in v0.6: rank the k cheapest
assignments of an LSAP in order of cost, using Murty's method on top of
incremental `lsap`. Ranked assignments are what callers need when the
cheapest assignment alone isn't enough, for example to offer alternatives,
or to keep several hypotheses alive as in multi-target tracking.

Murty's method splits the assignments not yet ranked into subproblems.
Each one requires some edges of an earlier answer and forbids one other.
`lsap` already has the pieces that make this cheap: `Copy` shares cost
rows, `Remove` forbids an edge, and `Assign` re-optimizes from the
current dual prices, so solving a subproblem is a single augmentation
(O(n²) dense) rather than a full solve. What it lacks is a way to require
an edge. Requiring (u, v) by removing every other edge in row u and
column v would touch every row, and since `Copy` copies a row on its
first change, every waiting subproblem would own a full cost matrix:
8 MB at n = 1000 with `int64`, and tens of GB for small k.

## Goals

- Rank the k best assignments of a dense or sparse LSAP, exactly, in
  nondecreasing order of cost, with no assignment repeated.
- Memory that grows with k·n, not with k·n², so that n in the thousands
  is routine.
- A solver with the same shape as `lsap`, and a JSON CLI with the shared
  conventions.
- Never return a wrong ranking. Overflow is reported as `ap.ErrOverflow`,
  as in `lsap`.

## Non-goals

- k-best for other problem types (bottleneck, QAP).
- Parallel subproblem solving.
- Re-solving waiting subproblems from scratch when they're picked. The
  owner considered that approach and found it too risky.

## Anti-goals

- Slowing down `lsap`'s inner loops to support requirements.
- A ranking that skips or repeats an assignment when costs tie.

## Approaches (for requiring an edge)

1. **A public way to require an edge in `lsap` (chosen).** It requires
   edge (u, v) in every later assignment. The kernel keeps u assigned to v and excludes
   row u and column v from later searches. The dense search loop already
   skips excluded columns, so the inner loop gains no work. Requiring an
   edge is O(n) and copies no cost rows. The QAP branch-and-bound in v0.7
   needs the same operation.
2. **The same mechanism, unexported,** with k-best in `lsap` or reaching
   it through an internal hook. This keeps `lsap`'s API smaller, but v0.7
   would expose it anyway, and it breaks the pattern of one problem per
   package.
3. **Require an edge by giving it a very negative cost.** No new search
   logic, but it shrinks the range of costs that solve without overflow,
   which works against #34. Rejected.

The owner chose approach 1. They rejected the name `Force`, which could
be read as forcing an edge to 0 or to 1, and chose
`Fix(u, v int, value bool)`, which fixes the edge's variable x_uv to 1
(`true`, required) or 0 (`false`, forbidden). `Fix` replaces `Remove`, so
there is one way to forbid an edge. `Fix` takes indices rather than an
`ap.Edge` or `ap.Pair`, to match `ReducedCost(u, v)` and to avoid an
ignored `Cost` field.

## Design

### `lsap.Fix(u, v, value)`

- `Fix(u, v, false)` forbids edge (u, v) in every later call to `Assign`,
  exactly as `Remove(u, v)` does today. `Remove` is deleted (a breaking
  change, allowed before 1.0 and listed in the CHANGELOG). Its callers,
  the `lsap` tests and the `ReducedCost` doc example, move to `Fix`.
- `Fix(u, v, true)` requires edge (u, v) in every later call to
  `Assign`. It costs O(n) and copies no cost rows.
- `Fix` is a no-op if the edge is out of range, or if it is already fixed
  to the same value.
- If the requirements can't all be met, `Assign` returns
  `ap.ErrInfeasible`. That covers requiring a forbidden or missing edge,
  fixing one edge both ways (in either order), two required edges that
  share a row or a column, and requirements that leave the rest of the
  problem infeasible. There is no way to unfix an edge; callers that need
  to back out take a `Copy` first.
- Every other edge in a required edge's row or column is unavailable,
  as if forbidden: `ReducedCost` returns false for it, and it never appears
  in an assignment.
- If the current assignment doesn't contain a required (u, v), `Fix`
  unassigns row u and column v, and `Assign` re-augments them, as
  forbidding an assigned edge does.
- `Copy` carries requirements. `Cost` includes required edges, and the
  dual prices satisfy complementary slackness on them.
- Both kernels support it. The dense kernel marks required columns as
  already scanned when a search starts, and skips them in its O(n) dual
  update loop. The sparse kernel skips required columns in its search.
- The overflow margin and checks from #34 are unchanged. Required rows
  and columns take part in them like any others.

### `lbap.Fix` and `mcm.Fix` (issue #38)

The owner chose to replace `Remove` with `Fix` in all three solvers, so
that the packages stay alike (`internal/conformance` checks their shared
shape).

- Same signature and rules as `lsap.Fix`. `Fix(u, v, false)` is today's
  `Remove`. Conflicting fixes make `Assign` return `ap.ErrInfeasible`.
- `Fix(u, v, true)` forbids every other live edge in row u and column v
  and records the requirement, so that later conflicts are detected.
  Neither solver feeds k-best, so this may take time linear in the number
  of edges. Any maximum matching of the reduced graph contains (u, v),
  since u and v have no other edges.
- `mcm.Match` gains an error: `Match() ([]int, error)`. It returns
  `ap.ErrInfeasible` when the required edges can't all be in one
  matching, which is exactly when they conflict, and otherwise a maximum
  matching that contains every required edge. `Size` is 0 after a
  conflict.

### `klsap`: k-best linear sum assignment

- **Name.** `klsap` (k-best LSAP), solver type `KLSAP`, binary
  `cmd/klsap`, following the acronym rule. The package doc defines the
  acronym.
- **API.** `klsap.New(costs [][]T, k int)` and
  `klsap.NewSparse(n int, edges []ap.Edge[T], k int)` validate as `lsap`
  does and reject k < 1. Then `Next() (ap.Permutation, error)` returns
  the next assignment in the ranking and `Cost() T` its cost. `Next`
  returns `ap.ErrInfeasible` once no further assignment exists, which
  includes the first call for an infeasible problem, and after k
  assignments. `ap.ErrOverflow` from any subproblem is sticky, since the
  ranking can't continue without that subproblem's cost.
- **k up front.** Taking k in the constructor bounds memory. After r
  assignments have been returned, only the k − r cheapest waiting
  subproblems can contribute, since the others' best assignments already
  cost at least as much as k − r other assignments. So the solver keeps
  at most k − r waiting subproblems and drops the rest. That holds memory
  to O(k) subproblems of O(n) each.
- **Algorithm.** Murty's method. Each waiting subproblem is an
  `lsap.LSAP` that has been solved, ordered by cost in a heap. To return
  the next assignment, pop the cheapest. To split it, for each edge
  (i, π(i)) of its assignment that isn't already required, take a `Copy`,
  `Fix` the edges before it to `true` and it to `false`, and `Assign`. Infeasible
  children are dropped.
- **Ties.** The order among assignments of equal cost is unspecified but
  deterministic for a given input. Murty's partition guarantees that no
  assignment is returned twice.
- **Subproblem size.** A waiting subproblem holds its row pointers, its
  assignment and dual prices, and the one cost row that forbidding its edge copied.
  `lsap.Copy` today also copies scratch slices that are only used while
  solving (`p`, `c`, `pi`). The builder should drop those from waiting
  copies if the memory benchmark shows they matter.
- **Refinements.** The literature has refinements to Murty's method: the
  order in which children are created, solving a child only when it might
  be needed, and reusing the parent's solution. Benchmarks decide which,
  if any, are worth adding.

### `cmd/klsap`

- Input: the same dense or sparse JSON as `cmd/lsap`.
- Flags: `-k` (required, at least 1), plus `-scale`, `-cycles` and
  `-matrix` as in `cmd/lsap`, applied to each assignment.
- Output:

  ```json
  {"status":"optimal","assignments":[{"cost":265,"permutation":[3,2,1,0]},
    {"cost":270,"permutation":[3,2,0,1]}]}
  ```

  Fewer than k assignments are listed if the problem has fewer. If it
  has none, the status is `"infeasible"` and `"assignments"` is empty.
  No dual prices.

## Tests

- **`Fix` against brute force.** Random small dense and sparse
  instances with random edges fixed each way. `Assign` matches brute
  force restricted to the assignments that contain every required edge
  and no forbidden one, or both report infeasible. This includes
  conflicting requirements and fixing one edge both ways. The existing
  `Remove` tests carry over as `Fix(u, v, false)`.
- **Ranking against brute force.** For small n, ask for every assignment
  (k = n!, or more). The costs must match brute force's sorted list of
  all feasible assignment costs, every permutation must be feasible and
  appear exactly once, and the order must be nondecreasing. Include
  instances with many ties (for example two-cost) and sparse instances
  with few feasible assignments.
- **Pruning.** For small k on instances with many assignments, the
  ranking matches the first k of the full ranking's costs.
- **Overflow.** An `int8` instance where a subproblem overflows makes
  `Next` return `ap.ErrOverflow`, and keeps returning it.
- **CLI.** The shared CLI tests, plus `-k` validation and the output
  shape above.

## Benchmarks

Go benchmarks only (`go test -bench`), on the `internal/instances`
classes: `Fix` must not slow `lsap.Assign` on the existing suite
(within noise), and `klsap` gets benchmarks for time and allocated bytes
at n = 100 and 1,000 with k = 10 and 100. The refinements above stay
only if these show a gain.

## Docs

Package docs and citations (Murty 1968, plus any refinement used, each
checked against the original), the README, the CHANGELOG, and the suite
design doc's roadmap row.

## Related: Go-only benchmarks

The owner wants the repo to be Go only. `bench/compare.py` and its
SciPy/OR-Tools comparison go, in their own v0.6 issue and PR: the large
benchmark sizes move into opt-in `go test -bench` benchmarks, and
`bench/RESULTS.md` records Go benchmark results only.

## Calls made by Claude (for the owner to check)

- The package name `klsap`.
- k is given up front, which bounds memory, rather than an open-ended
  stream.
- `Next` signals the end of the ranking with `ap.ErrInfeasible`, rather
  than a new sentinel.
- Edges sharing a required edge's row or column read as forbidden in
  `ReducedCost`.
- `cmd/klsap` requires `-k` and prints no dual prices.
- `mcm.Match` returns an error, so that conflicting fixes have a way to
  be reported.

## Left to the builder

- How the kernels record requirements, and how `Fix` interacts with
  `initialize`.
- The heap and pruning data structures.
- Whether waiting copies drop scratch slices.
- Brute-force helpers for enumerating and filtering assignments.
- Book section and algorithm numbers for the citations, checked against
  the book.
