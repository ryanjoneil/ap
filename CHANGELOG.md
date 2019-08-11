# ap: Change Log

## v0.2.0

* `ap.Assignment` is renamed to `ap.Permutation`.
* `ap.Inverse(ap.Assignment)` is now `Permutation.Inverse()`.
* `ap.ToMatrix(ap.Assignment)` is now `Permutation.Matrix()`.
* `ap.ToPermutation(ap.Matrix)` is now `Matrix.Permutation()`.

## v0.1.0

* New API supports multiple forms of assignment problems.
* Split code into `ap` and `lsap` packages.
* Added benchmarks and examples.
* Adopted Go modules and versioning.
* Added `lsap` binary under `cmd` subfolder.
* Licensed library under APL v2.
