# External benchmark comparison

Compares `lsap` with SciPy and OR-Tools on the instance classes of book
Section 4.10.1. Results are in [RESULTS.md](RESULTS.md).

## Setup

    python3 -m venv bench/.venv && bench/.venv/bin/pip install scipy ortools

If `python3 -m venv` lacks pip (Debian/Ubuntu without `python3-venv`), use
uv instead: `uv venv bench/.venv && uv pip install -p bench/.venv/bin/python scipy ortools`.

## Instances

Generated into `bench/instances/` (git-ignored) with `internal/cmd/geninstance`
(seed 1, K = 10^6):

- dense uniform and geometric, n = 1000 and 5000
- dense two-cost, n = 1000 and 5000
- dense Machol-Wien, n = 1000 and 2000
- sparse uniform and geometric, n = 100000, degree 10

For example:

    go run ./internal/cmd/geninstance -class uniform -n 1000 > bench/instances/dense-uniform-1000.json
    go run ./internal/cmd/geninstance -class geometric -n 100000 -degree 10 > bench/instances/sparse-geometric-100000.json

## Running

    bench/.venv/bin/python bench/compare.py bench/instances/*.json

Each solver runs 3 times per instance and the median is reported. A run that
exceeds 600 s is killed and reported as a timeout (`COMPARE_TIMEOUT` overrides
the limit). Differing costs make the script exit non-zero.
