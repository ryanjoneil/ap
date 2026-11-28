# Benchmark results

Date: 2026-11-28. Code: `v0.6-klsap-24` at 70300e9, plus the #39 benchmarks.

## Environment

- Machine: AMD Ryzen Threadripper PRO 5955WX 16-Cores (16 cores, 32 threads),
  247 GB RAM, Linux 7.0.0-44-generic
- Go: go1.27.0 linux/amd64

The machine was not idle: unrelated jobs from another project kept the load
average between 7 and 9 throughout. Each benchmark is single-threaded, so
this mostly shows up as occasional slow runs (for example, one geometric
n = 5,000 run took 14.3 s against a median of 7.9 s). Medians resist those.

## Commands

    AP_BENCH_LARGE=1 go test -p 1 -run '^$' -bench Large -benchtime 1x -count 5 -timeout 0 ./lsap ./klsap > large.txt
    go run golang.org/x/perf/cmd/benchstat@v0.0.0-20261117185222-4febdb082b2c large.txt

See [README.md](README.md) for what each benchmark times. Instances use seed
1 and K = 10⁶, the same instances as v0.5's comparison.

## `lsap` (median of 5 runs, seconds)

The last column is v0.5.0's `lsap` on the same instances, from that
release's `bench/RESULTS.md` (median of 3 runs, timed by a helper command
rather than `go test`, on a similarly loaded machine). It's there for scale.
Use benchstat on runs from the same session to compare versions.

| benchmark | n | edges | v0.6 (s) | v0.5.0 (s) |
|---|---|---|---|---|
| dense uniform | 1,000 | 10⁶ | 0.063 | 0.042 |
| dense uniform | 5,000 | 2.5·10⁷ | 1.526 | 1.751 |
| dense geometric | 1,000 | 10⁶ | 0.094 | 0.171 |
| dense geometric | 5,000 | 2.5·10⁷ | 7.882 | 13.289 |
| dense two-cost | 1,000 | 10⁶ | 0.008 | 0.011 |
| dense two-cost | 5,000 | 2.5·10⁷ | 0.256 | 0.258 |
| dense Machol–Wien | 1,000 | 10⁶ | 1.526 | 3.225 |
| dense Machol–Wien | 2,000 | 4·10⁶ | 12.29 | 23.753 |
| sparse uniform | 100,000 | 10⁶ | 3.927 | 6.146 |
| sparse geometric | 100,000 | 10⁶ | 4.865 | 5.454 |

Dense uniform at n = 1,000 is the noisiest row: its runs ranged from 0.038 s
to 0.069 s.

## `klsap` (median of 5 runs)

| benchmark | time (s) | allocated per ranking | allocations | held per waiting subproblem |
|---|---|---|---|---|
| dense uniform, n = 1,000, k = 100 | 67.89 | 10.1 GiB | 2.11 M | 115 KiB |

Each waiting subproblem holds about 115 KiB at n = 1,000, so a ranking with
k = 100 keeps about 11 MiB of subproblems alive, plus the shared 8 MB cost
matrix. Most of the time and allocation goes to the up to n − 1 children solved
at each split, most of which the bounded queue then drops.
