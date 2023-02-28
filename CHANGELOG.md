# ap: Change Log

## v0.3.0

* #15: Add `ap.Arc` type for convenient matrix creation.
* Misc:
  * Update Go to v1.20.
  * Update linting and GitHub Actions.

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
