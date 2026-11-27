# ap: Change Log

## v0.6.0 (unreleased)

New:

* #34: `ap.ErrOverflow` is the error `lsap`'s `Assign` returns when solving
  would overflow `T`.

Changed:

* #38: Breaking: `Remove` is replaced by `Fix(u, v, value)` in `lsap`, `lbap`
  and `mcm`, which can also require an edge. `mcm.Match` returns an error.
* #34: `lsap` detects overflow while solving instead of ruling it out up
  front. `lsap.New` now rejects only a cost equal to `T`'s minimum, or an
  M = n·(maxC − minC) + maxC + 1 that overflows or equals `T`'s maximum. For
  example, with `int32` and n = 1000, costs in [0, 2,145,338] are accepted
  (v0.5 accepted [0, 535]). `lsap.NewSparse` rejects only costs equal to `T`'s
  minimum or maximum. `Assign` returns `ap.ErrOverflow` when the solve would
  overflow `T`. That happens most often for costs above about a third (dense)
  or a fifth (sparse) of `T`'s largest value, or once removed edges bring M
  into the dense costs, but the dual prices can exceed the costs, so it can
  happen for smaller costs too. The `lsap` CLI reports it as "use a smaller
  -scale or smaller costs".

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
* `Remove` and `Copy` on every solver, including the new ones; the sparse and
  dense solvers re-solve incrementally after a `Remove`.
* `lsap.LSAP.Copy` may be called concurrently on the same solver, so a solved
  parent can seed parallel work. No other method may run concurrently.
* `ap.Edge` and `ap.Pair` describe sparse input (weighted and unweighted
  edges), and `ap.ErrInfeasible` is the error every solver's `Assign` returns
  when no feasible assignment exists.
* Without `-scale`, the CLIs accept a cost such as `5.0` or `1e3` whose value is
  an integer of magnitude at most 2^53.

Breaking changes (#19: `ap`/`lsap` API cleanup):

* `lsap.New` returns `(*LSAP, error)`. Handle the error; bad input no longer
  panics.
* `lsap.New` rejects cost ranges that could overflow `T` while solving; see
  `New` for the bound. For example, with `int32` and n = 1000, costs in
  [0, 535] are accepted but costs in [0, 536] are not. Use `int64`, or
  `lsap.NewSparse`, for wider ranges.
* `lsap.New` copies its input. v0.4 kept the caller's matrix and modified it
  when edges were removed.
* `Copy` returns `*LSAP[T]` instead of the removed `Assigner[T]` interface.
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
