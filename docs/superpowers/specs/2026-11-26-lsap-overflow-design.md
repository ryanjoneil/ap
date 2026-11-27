# Overflow detection in lsap: design

Issue #34. Milestone v0.6.

## Why

`lsap.New` rejects any cost range for which overflow is possible in the
worst case. It requires 4(n+1)·max(|minC|, |M|) to fit in `T`, where
M = n·(maxC − minC) + maxC + 1 is the cost given to removed edges. That is
roughly 4n²·range. With `int32` and n = 1000, only costs in [0, 535] are
accepted, which rules out every benchmark class except narrow uniform
instances. `NewSparse` has a looser bound of the same kind,
4(n+1)·max |c|.

In practice the duals stay far below these bounds. At n = 1000 (3 seeds per
class, with and without removals), the largest final |u| or |v| was 0.01×
the cost range for uniform costs, 0.53× for geometric, under 0.005× for
two-cost and 0.50× for Machol–Wien. The owner's point: just because costs
can overflow doesn't mean they will.

## Goals

- Accept every input whose costs can be represented, and fail only when
  overflow would actually occur.
- Never return a wrong answer because of overflow. Go integers wrap
  silently, so every operation that could overflow must be checked or
  proven safe.
- Keep the dense and sparse kernels within 5% of v0.5 on the benchmark
  suite.

## Non-goals

- Automatic fallback to a wider type, or to big integers.
- Changes to `lbap` and `mcm`. Neither does arithmetic on costs except
  `lbap`'s comparisons and maximum.

## Approaches

1. **Detect overflow where it can happen (chosen).** This is how OR-Tools'
   min-cost flow handles it: a cheap up-front check, then a check whenever
   a node potential crosses a threshold, returning `BAD_COST_RANGE` only if
   that happens.
2. **Prove a tighter bound and keep checking up front.** This rejects fewer
   inputs, but it still rejects some that would solve. OR-Tools'
   `SimpleLinearSumAssignment` does this, and its own comments call the
   bound pessimistic.
3. **No checks.** This is fast, but it can return wrong answers without any
   error. Rejected.

The owner chose approach 1.

## Design

- **New sentinel.** `ap.ErrOverflow = errors.New("ap: arithmetic overflow")`
  goes in the root package next to `ErrInfeasible`. `Assign` returns it
  when the solve would overflow `T`. After `Assign` returns it, results are
  undefined, as after any failed `Assign`, and callers should switch to a
  wider `T`. The docs suggest calling `Copy` first if the caller wants to
  keep a usable solver.
- **Up-front checks.**
  - `New` requires only that M can be computed without overflow. With
    `int32` and n = 1000 that allows ranges up to 2,145,338.
  - `NewSparse` drops its cost bound. Any edge costs are accepted.
- **Dense kernel.** The inner loop computes `a[r][j] − u[r] − v[j]` n²
  times per augmentation. It stays unchecked: the kernel keeps every |u[i]|
  and |v[j]| within a margin B, chosen so that max |a| + 2B fits in `T`.
  The duals change only in the O(n) update loops of `path` and
  `initialize`, so the check that every new value stays within B goes
  there. A value outside B makes `Assign` return `ErrOverflow`.
- **Sparse kernel.** Path lengths grow along each search, so the label
  arithmetic (`d + c[e] − v[k]` and `it.d − u[i]`) is checked per scanned
  edge with `internal/intmath`, or with a margin like the dense one,
  whichever benchmarks better. The dual updates are checked as in the dense
  kernel.
- **Cost.** `Assign` sums the assignment's cost with checked adds and
  stores the result, so `Cost` never overflows.
- **`ReducedCost`.** It is safe under the same margin. If the margin
  doesn't cover it, `ReducedCost` uses checked arithmetic, and on overflow
  it returns `ok == false`.
- **Big M.** It stays for now. Its soundness argument is unchanged: the
  optimum uses an M edge only if no assignment exists without one.
  Replacing M with forbidden-edge marks would remove the last up-front
  limit, but it adds a check to the dense inner loop. That is a separate
  follow-up, and benchmarks decide it.
- **CLI.** `cmd/lsap` already sends every non-infeasibility error from
  `Assign` to stderr and exits 1, so it needs no new status. The message
  should say what to do, for example to use a smaller `-scale`.

## Tests

- **Never wrong.** A property test runs random small instances in `int8`
  and `int16`, including removals. Each `Assign` either returns
  `ErrOverflow` or matches `internal/bruteforce` computed in `int64`. This
  is the key guarantee.
- **Overflow is detected.** At least one instance passes `New` or
  `NewSparse`, then returns `ErrOverflow` from `Assign`.
- **No false alarms.** The benchmark classes from the earlier measurement
  solve in `int32` at n = 1000, matching their `int64` results.
- **Boundaries.** `New` is tested at the largest and smallest ranges for
  which M fits, and the old boundary tests are removed.
- **Benchmarks.** Results are compared with v0.5 on the existing suite.

## Docs

- Update `New`'s and `NewSparse`'s comments, the CHANGELOG, and the design
  doc's "Input validation" rule.
- The new rule: constructors reject costs that can't be represented, and
  `Assign` returns `ap.ErrOverflow` if overflow would actually occur.

## Left to the builder

- The exact margin B, and where the checks sit.
- Checked arithmetic or a margin in the sparse kernel.
- Whether `initialize` needs its own checks, or is covered by the bound on
  the costs.
