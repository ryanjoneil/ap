# ap: Change Log

## v0.5.0

New:

* #20: Sparse LSAP: `lsap.NewSparse` solves problems given as a list of edges,
  where a missing edge is forbidden.
* #22: Bipartite matching solver: new `mcm` package and `cmd/mcm` find a
  maximum cardinality matching (Hopcroft-Karp).
* #2: Linear bottleneck assignment: new `lbap` package and `cmd/lbap`
  (threshold algorithm).
* #21: Shared CLI conventions: all binaries read one JSON value (array for
  dense, object for sparse input), write one JSON object, and accept `-scale k`
  for non-integer costs.
* #23: Benchmark suite and external comparison with SciPy and OR-Tools, in
  `bench/`.
* `Remove` and `Copy` on every solver; the sparse and dense solvers re-solve
  incrementally after a `Remove`.

Breaking changes (#19: `ap`/`lsap` API cleanup):

* `lsap.New` returns `(*LSAP, error)`. Handle the error; bad input no longer
  panics.
* `Assign` returns `(ap.Permutation, error)`. Check the error, which is
  `ap.ErrInfeasible` when no feasible assignment exists.
* `M` and `SetM` are removed. The solver handles infeasibility itself, so
  delete calls to them.
* The `ap` interfaces (`Assigner`, `Coster`, `ReducedCoster`, `DualPricer`)
  and the `lsap` interfaces are removed. Depend on the concrete solver types, or
  declare the small interface you need.
* `ap.MinOf` and `ap.MaxOf` are now internal. Use the `math` package constants
  for the integer type you need, such as `math.MinInt64`.
* `ReducedCost` returns `(T, bool)`. The bool is false for a missing or removed
  edge.
* CLI output gains `"status"`; infeasible problems print
  `{"status":"infeasible"}` and exit 0.

Follow-ups filed for later: #32 one-shot dense LSAP algorithm, #33 faster
sparse LSAP.

## v0.4.0

* Fix LSAP solver returning suboptimal assignments for cost matrices with
  negative values. Initialization compared a raw cost against a reduced cost
  when breaking ties.
* LSAP solver breaks ties by lowest column index, so it returns the same
  assignment for the same input. Previously, map iteration order made results
  vary between runs when there were multiple optimal assignments.
* Add brute-force and regression tests for the LSAP solver.
* Fix LSAP benchmarks, which only generated costs on the diagonal.
* Update Go to v1.27.
* Update GitHub Actions and golangci-lint to v2.
* Fix README CLI output and usage examples.

## v0.3.0

* LSAP solver auto-computes big-M. 
* LSAP solver provides `M` and `SetM` methods.
* Update Go to v1.21.
* Update linting and GitHub Actions.
* Assigners operate on generic signed `int` types. 
* Remove `Int64*` interface prefixes.
* `ap.Matrix` is a `[][]bool` again.
* Improve docs and examples.

## v0.2.0

* #12: Add cycle and matrix output to lsap command
  * `lsap` provides `-cycles` and `-matrix` flags for assignment forms.
  * `ap.Matrix` is now a `[][]int8` instead of `[][]bool`.
* #10: Add reduced cost and dual price output to lsap command
  * `lsap` uses `-dual` and `-rc` flags for dual prices and reduced costs.
* #6: Provide Cycle struct and method
  * Adds `ap.Cycle` and `ap.Cycles`.
  * `Permutation`, `Cycles,` and `Matrix` structs all convert to each other.
  * All assignment problem representations provide `Inverse` methods.
* #5: Rename "Assignment" to "Permutation"
  * `ap.Assignment` renamed to `ap.Permutation`.
  * `ap.Inverse(ap.Assignment)` is now `Permutation.Inverse()`.
  * `ap.ToMatrix(ap.Assignment)` is now `Permutation.Matrix()`.
  * `ap.ToPermutation(ap.Matrix)` is now `Matrix.Permutation()`.

## v0.1.0

* #1: AP code and module refactor
  * New API supports multiple forms of assignment problems.
  * Splits code into `ap` and `lsap` packages.
  * Adds benchmarks and examples.
  * Adopts Go modules and versioning.
  * Adds `lsap` binary under `cmd` subfolder.
  * Licenses library under APL v2.
