# Benchmarks

All benchmarks are Go benchmarks in the solver packages' `bench_test.go`
files. Results are in [RESULTS.md](RESULTS.md).

## Running

`go test -bench .` runs the benchmarks sized to stay practical. The large
benchmarks, whose names start with `BenchmarkLarge`, skip themselves unless
`AP_BENCH_LARGE` is set:

    AP_BENCH_LARGE=1 go test -p 1 -run '^$' -bench Large -benchtime 1x -count 5 -timeout 0 ./lsap ./klsap

`-p 1` runs one package at a time, so the packages don't compete for the
machine. `-benchtime 1x` times one run per count, since a single run takes
seconds to minutes. `-timeout 0` lifts the default 10-minute test timeout.

The large benchmarks are:

- `lsap`'s `BenchmarkLargeDense`: the uniform, geometric and two-cost classes
  of book Section 4.10.1 at n = 1,000 and 5,000, and Machol–Wien at n = 1,000
  and 2,000.
- `lsap`'s `BenchmarkLargeSparse`: uniform and geometric at n = 100,000 with 10
  edges per row.
- `klsap`'s `BenchmarkLargeDense`: ranking the k = 100 best assignments of a
  uniform instance with n = 1,000.

Instances come from `internal/instances` with seed 1 and K = 10⁶, generated
outside the timed loop. `lsap` benchmarks time `New` or `NewSparse` plus
`Assign`, which includes the solver's copy of its input. `klsap`'s dense
benchmarks also report `held-B/subproblem`, the live heap each waiting
subproblem holds.

## Comparing

To compare two versions, save each one's output and run
[benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

    go test -run '^$' -bench . -count 10 ./lsap > old.txt
    # change the code
    go test -run '^$' -bench . -count 10 ./lsap > new.txt
    go run golang.org/x/perf/cmd/benchstat@v0.0.0-20261117185222-4febdb082b2c old.txt new.txt
