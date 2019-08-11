# ap: Change Log

## v0.2.0

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
