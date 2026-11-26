# External benchmark results

Date: 2026-11-25

## Environment

- Machine: AMD Ryzen Threadripper PRO 5955WX 16-Cores (16 cores, 32 threads), 247 GB RAM, Linux 7.0.0-44-generic
- Go: go1.27.0 linux/amd64
- Python: 3.14.4
- SciPy 1.18.1, NumPy 2.5.5, OR-Tools 9.15.6755 (from `pip freeze` of `bench/.venv`)

The machine was not otherwise idle: unrelated single-threaded jobs from another
project were running during part of the session, so absolute times may carry
some noise. All solvers ran sequentially, one at a time.

## Commands

    # instances (seed 1, K = 1000000), written to bench/instances/
    go run ./internal/cmd/geninstance -class uniform|geometric|twocost -n 1000|5000 > bench/instances/dense-<class>-<n>.json
    go run ./internal/cmd/geninstance -class macholwien -n 1000|2000 > bench/instances/dense-macholwien-<n>.json
    go run ./internal/cmd/geninstance -class uniform|geometric -n 100000 -degree 10 > bench/instances/sparse-<class>-100000.json

    # comparison
    bench/.venv/bin/python bench/compare.py bench/instances/*.json

## What is timed

Each solver runs 3 times per instance; the table shows the median in seconds.
Only the solve is timed, not JSON parsing or array construction.

- ap-lsap: `internal/cmd/timesolve` times `lsap.New`/`lsap.NewSparse` plus
  `Assign`, excluding JSON parsing. This includes the solver's internal copy of
  the input (`New` copies), which is part of its API cost.
- SciPy: `linear_sum_assignment` (dense) and
  `min_weight_full_bipartite_matching` on a CSR matrix with costs +1 (sparse;
  for dense instances the full matrix is passed as CSR, as an extra data point).
  The numpy arrays / CSR matrix are built before the timer starts.
- OR-Tools: `SimpleLinearSumAssignment`; the timer covers
  `add_arcs_with_cost` (bulk, from prebuilt numpy arrays), `solve` and reading
  the optimal cost. The Python API matched the brief's names in 9.15.
- Each external run is in a forked child process (copy-on-write instance), so
  the 600 s timeout can be enforced. No run timed out.
- All solvers' optimal costs agree on every instance (checked by compare.py).

## Results (median of 3 runs, seconds)

| instance | n | edges | ap-lsap (s) | scipy dense (s) | scipy sparse (s) | ortools (s) |
|---|---|---|---|---|---|---|
| dense-geometric-1000.json | 1000 | 1000000 | 0.171 | 0.104 | 0.270 | 0.135 |
| dense-geometric-5000.json | 5000 | 25000000 | 13.289 | 6.597 | 8.271 | 5.447 |
| dense-macholwien-1000.json | 1000 | 1000000 | 3.225 | 0.635 | 1.766 | 0.203 |
| dense-macholwien-2000.json | 2000 | 4000000 | 23.753 | 2.881 | 8.180 | 0.813 |
| dense-twocost-1000.json | 1000 | 1000000 | 0.011 | 0.015 | 0.025 | 0.156 |
| dense-twocost-5000.json | 5000 | 25000000 | 0.258 | 0.255 | 0.451 | 4.228 |
| dense-uniform-1000.json | 1000 | 1000000 | 0.042 | 0.037 | 0.031 | 0.096 |
| dense-uniform-5000.json | 5000 | 25000000 | 1.751 | 1.959 | 0.794 | 2.655 |
| sparse-geometric-100000.json | 100000 | 1000000 | 5.454 | - | 24.400 | 0.486 |
| sparse-uniform-100000.json | 100000 | 1000000 | 6.146 | - | 19.892 | 0.508 |

`-` means not applicable (SciPy's dense solver is only run on dense instances).

## Ratio of ap-lsap to the fastest external solver

| instance | fastest external | ap-lsap / fastest |
|---|---|---|
| dense-uniform-1000 | scipy sparse 0.031 | 1.35x |
| dense-uniform-5000 | scipy sparse 0.794 | 2.21x |
| dense-geometric-1000 | scipy dense 0.104 | 1.64x |
| dense-geometric-5000 | ortools 5.447 | 2.44x |
| dense-twocost-1000 | scipy dense 0.015 | 0.73x |
| dense-twocost-5000 | scipy dense 0.255 | 1.01x |
| dense-macholwien-1000 | ortools 0.203 | 15.9x |
| dense-macholwien-2000 | ortools 0.813 | 29.2x |
| sparse-uniform-100000 | ortools 0.508 | 12.1x |
| sparse-geometric-100000 | ortools 0.486 | 11.2x |

## Findings

ap-lsap is competitive on the dense random classes: within 1.4x-2.5x of the
fastest external solver on uniform and geometric, level with SciPy on two-cost
(slightly faster at n = 1000, tied at n = 5000), and faster than OR-Tools on
uniform and two-cost at n = 5000. It is clearly slower on the Machol-Wien class
(16x-29x slower than OR-Tools, 5x-8x slower than SciPy's dense solver) and on
sparse instances, where it is 11x-12x slower than OR-Tools and about 4x faster
than SciPy's sparse solver. Only Machol-Wien and the sparse classes exceed the
3x threshold. The gap grows with n on Machol-Wien. Results are for one seed on
one machine, with three runs each.
