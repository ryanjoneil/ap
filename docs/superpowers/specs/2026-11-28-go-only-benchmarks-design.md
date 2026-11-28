# Go-only benchmarks: design

Issue #39. Milestone v0.6. Short design, approved by the owner on
2026-11-28.

## Why

The owner wants the repo to be Go only: "I saw python in the repo. Go only.
Use go's benchmarking functionality." v0.5's `bench/compare.py` timed `lsap`
against SciPy and OR-Tools from a Python virtual environment, fed by two Go
helper commands. All of it goes, and the large sizes it timed become Go
benchmarks.

## Approaches (for keeping large benchmarks out of `go test -bench .`)

1. **An environment variable (chosen, Claude's call).** Benchmarks named
   `BenchmarkLarge...` skip themselves unless `AP_BENCH_LARGE` is set. They
   are still compiled, vetted and linted on every run, so they can't rot.
2. **A build tag.** Keeps them out of normal builds, but also out of `go vet`
   and golangci-lint unless every run passes the tag.
3. **Name filtering alone.** `-bench .` matches every name, so this isn't
   opt-in.

## Design

- Remove `bench/compare.py`, `internal/cmd/timesolve`,
  `internal/cmd/geninstance`, the Python setup in `bench/README.md`, and the
  `bench/.venv/` and `bench/instances/` lines in `.gitignore`.
- `instances.SkipUnlessLarge(tb testing.TB)` skips unless `AP_BENCH_LARGE`
  is non-empty, so that every package uses the same switch.
- `lsap`: `BenchmarkLargeDense` runs uniform, geometric and two-cost at
  n = 1,000 and 5,000, and Machol–Wien at n = 1,000 and 2,000.
  `BenchmarkLargeSparse` runs uniform and geometric at n = 100,000, degree 10.
  Instances use seed 1 and K = 10⁶, as `geninstance` did, so the times are
  comparable with v0.5's. `BenchmarkDenseUniform5K` and
  `BenchmarkSparseUniform100K` move into this set.
- `klsap` (Claude's call): `BenchmarkLargeDense` ranks n = 1,000 with k = 100.
  The dense benchmarks report `held-B/subproblem`: the live heap that the
  first split adds, after a garbage collection, divided by the subproblems it
  keeps.
- `bench/RESULTS.md` records Go benchmark results only: machine, Go version,
  command and benchstat output. `bench/README.md` says how to run them. The
  README's Benchmarks section, the suite design doc's External comparison
  bullet and the CHANGELOG follow.
- The branch is stacked on #42 (klsap), since the klsap benchmarks need it.
