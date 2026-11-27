# lsap Overflow Detection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `lsap.New` and `lsap.NewSparse` stop rejecting costs that merely
could overflow. `Assign` returns `ap.ErrOverflow` only when the solve would
leave the range where its arithmetic is safe.

**Architecture:** Each kernel keeps its duals inside a margin B, chosen so
that every reduced cost `c − u − v` fits in `T`. B is checked only where
the duals change (the O(n) update loops and `initialize`), so the hot loops
stay unchecked. The sparse kernel also caps path labels, checked once per
popped column. The total cost is summed with checked adds. A sticky
`overflow` flag makes `Assign` return `ap.ErrOverflow`.

**Tech Stack:** Go 1.27, standard library only.

**Spec:** `docs/superpowers/specs/2026-11-26-lsap-overflow-design.md`
(issue #34). Repo rules are in `AGENTS.md`.

## Global Constraints

- Standard library only. No `require` lines in `go.mod`.
- New sentinel, exactly:
  `var ErrOverflow = errors.New("ap: arithmetic overflow")`, in `edge.go`
  next to `ErrInfeasible`.
- Never return a wrong answer because of overflow. Every arithmetic
  operation in both kernels is either checked or proven safe by the margin
  invariant. The proof goes in a comment at the margin's definition.
- `ErrInfeasible` takes precedence over `ErrOverflow` when infeasibility is
  structural (sparse: a row or column with no edges).
- After `Assign` returns `ErrOverflow`, the solver's results are undefined,
  as after any failed `Assign`. Every later `Assign` on it, and on any
  `Copy` of it, returns `ErrOverflow` again.
- Performance budget: each kernel must stay within 5% of v0.5.0 on the
  `lsap` Go benchmarks and on the Machol–Wien timing in Task 3.
- Run all four AGENTS.md checks before every commit.
- Commit messages end with:
  ```
  Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_014seuZ3PDN1H6ii3nMKxSFY
  ```

## Review Focus

1. **Removals after a solve shrink the dense margin.** The first `Remove`
   brings M into the costs, so B can drop below duals that were valid
   before. `Assign` must re-check every |u|, |v| against the new B before
   solving. Task 1's property test removes edges between solves.
2. **Costs equal to the largest value of `T`** were impossible before,
   because the old bound excluded them. The sparse `initialize` used
   `v[j] == maxT` to mean "column has no edges". Task 2 replaces that
   sentinel and tests a cost of 126 in `int8`.
3. **Costs equal to the smallest value of `T`.** |MinOf| can't be
   represented, so both constructors reject such costs. Tasks 1 and 2 test
   this.
4. **Infeasible re-solves near the margin (dense).** Proving infeasibility
   means solving with M edges, which can overflow. Both `ErrInfeasible` and
   `ErrOverflow` are then acceptable, but a permutation is not. The
   property test allows exactly that.
5. **A label equal to `maxT` in the sparse search** would be read as
   "untouched" and the edge silently skipped. The label cap L keeps every
   computed label at or below `maxT − 1`.

---

### Task 1: `ap.ErrOverflow` and overflow detection in the dense kernel

**Files:**
- Modify: `edge.go` (add `ErrOverflow`)
- Modify: `lsap/dense.go`, `lsap/solve.go`, `lsap/initialize.go`
- Modify: `lsap/lsap.go` (doc comments for `New`, `Assign` and `LSAP`)
- Test: `lsap/solve_test.go`, new `lsap/overflow_test.go`

**Interfaces:**
- Produces: `ap.ErrOverflow`. Task 2 reuses the test helpers in
  `lsap/overflow_test.go`: `checkNeverWrong`, run from
  `TestDenseNeverWrongInt8`/`Int16`, and `TestNoFalseAlarmsInt32`.

**Rules for `New`:**
- Reject any cost equal to `intmath.MinOf[T]()`.
- Reject when M = n·(maxC − minC) + maxC + 1 overflows, or equals
  `intmath.MaxOf[T]()`.
- Remove the old `4·(n+1)·…` check.
- Use the error message
  `"lsap: cost range [%d, %d] is too wide for %T with n = %d"`.
- With `int32` and n = 1000, range [0, 2,145,338] is accepted and
  [0, 2,145,339] is rejected.

**Margin (dense):**
- A = max(|minC|, |maxC|) until the first `Remove`, then max(A, |M|).
- B = (maxT − A − 1) / 2.
- Invariant: |u[i]| ≤ B and |v[j]| ≤ B. Then every `a[i][j] − u[i] − v[j]`,
  and every intermediate of it, lies in [−(maxT−1), maxT−1].
- `pi[j] −= d` is safe because d is the minimum of the `pi` values over the
  unlabelled columns.
- After `initialize`, u only increases and v only decreases. So `path`
  checks `d > B − u[h]` before `u[h] += d`, and `d > v[j] + B` before
  `v[j] −= d`. Neither check can overflow.
- `initialize` checks each v[j] as phase 1 sets it, and each u[i] as phase
  2 sets it, before u[i] is used.
- `assign` recomputes B, then checks every u and v. On any violation it
  sets the sticky `overflow` flag and returns `ap.ErrOverflow`.
- `z` is computed in `assign` with `intmath.Add`, after the M-edge
  infeasibility check, not in `solve`.
- `copy` carries `overflow` and the removed state.
- `reducedCost` needs no change, because the invariant covers it.

- [ ] **Step 1: Write the failing tests**

  In `lsap/overflow_test.go`:
  - `TestDenseMarginInt8`: `New([][]int8{{42}})` solves with cost 42.
    `New([][]int8{{43}})` succeeds, then `Assign` returns an error that is
    `ap.ErrOverflow`. Why: n = 1 gives A = c and v = c, and c ≤ (126 − c)/2
    holds for c = 42 but not for c = 43.
  - `TestDenseNeverWrongInt8` and `TestDenseNeverWrongInt16` call the
    generic `checkNeverWrong[T]` with 2000 trials each. Its signature is
    `checkNeverWrong[T ap.Integer](t *testing.T, build func(c [][]int64, mask [][]bool) (*lsap.LSAP[T], error), wantErr func(c [][]int64) bool)`,
    where `mask` marks the edges that exist (all true for dense):
    - n is uniform in 1..4. K is drawn from {1, 2, 5, 10, 20, 40, MaxT/4,
      MaxT/2, MaxT−1}, and costs are uniform in [−K, K].
    - Assert that `New` errors exactly when the rule above says it should.
      Skip the trial when it does.
    - Then loop as `TestSparseOverflowBoundInt8` does: alternate removing
      an assigned edge and a random edge, until brute force says the
      problem is infeasible.
    - At each step, an `ErrOverflow` from `Assign` ends the trial. If brute
      force is infeasible, the error must be `ErrInfeasible` or
      `ErrOverflow`. Otherwise `Cost()` must equal `bruteforce.MinSum` (in
      `int64`), the permutation's int64 cost must equal it too, and no
      removed edge may be used.
    - After all trials, assert that at least one step solved and at least
      one returned `ErrOverflow`, so neither outcome goes untested.
  - `TestNoFalseAlarmsInt32`, skipped under `testing.Short()`: `int32`
    instances from `internal/instances` at n = 1000 (uniform K = 10⁶,
    geometric k = 10⁶, two-cost), and Machol–Wien at n = 300. Each solves
    with `New` and gives the same cost as the `int64` solve.

  In `lsap/solve_test.go`:
  - Rewrite `TestNewBoundExampleInt32` for 2,145,338 accepted and
    2,145,339 rejected, using the same matrix shape.
  - Rewrite the int8 cases in `TestNewErrors`: `{{0, 41}, {0, 0}}` is
    accepted (M = 124). `{{0, 42}, {0, 0}}` is rejected (M = 127).
    `{{-128, 0}, {0, 0}}` is rejected.
  - Rename `TestOverflowBoundInt8` to `TestExhaustive2x2Int8`.

- [ ] **Step 2: Run the tests and confirm they fail**

  Run: `go test ./lsap/ -run 'Margin|NeverWrong|NoFalseAlarms|NewBound|NewErrors'`
  Expected: they fail to compile (`ap.ErrOverflow` is undefined) or fail
  their assertions under the old bound.

- [ ] **Step 3: Implement**

  - Add `ErrOverflow` to `edge.go`, with the doc comment "ErrOverflow is
    returned by solvers when solving would overflow the cost type T. Use a
    wider T."
  - Implement the rules and margin above.
  - Update `New`'s doc. State the rule and the `int32` example. Say that
    `Assign` may return `ap.ErrOverflow` for costs above about a third of
    T's largest value, or once removals bring in M. Suggest calling `Copy`
    first if a caller wants a solver to fall back on.
  - Add `ErrOverflow` to `Assign`'s doc.

- [ ] **Step 4: Run the tests and confirm they pass**

  Run: `go test -race ./...`
  Expected: PASS.

- [ ] **Step 5: Commit**

  `lsap: detect dense overflow while solving instead of up front (#34)`

### Task 2: Overflow detection in the sparse kernel

**Files:**
- Modify: `lsap/sparse.go`, `lsap/lsap.go` (`NewSparse` doc)
- Test: `lsap/overflow_test.go`, `lsap/sparse_test.go`

**Interfaces:**
- Consumes: `ap.ErrOverflow` and `checkNeverWrong[T]` from Task 1.

**Rules for `NewSparse`:**
- Reject any cost equal to `intmath.MinOf[T]()` or `intmath.MaxOf[T]()`.
- Remove the old `4·(n+1)·…` check.
- Use the message `"lsap: cost %d is out of range for %T"`.

**Margin and label cap (sparse):**
- A = max |c| over all edges. It never changes, since `Remove` only marks
  edges.
- B = (maxT − A − 1) / 4 and L = maxT − A − 2B − 1. Compute both in
  `newSparse`.
- Invariant: |u|, |v| ≤ B.
- Every popped label `it.d` must be ≤ L, checked right after the stale
  entry check and before the sink test. Then every intermediate of
  `(it.d − u[i]) + c[e] − v[k]` lies in [−(maxT−1), maxT−1], so every
  pushed label is ≤ maxT − 1.
- Dual updates check before each add: `delta > B − u[s]`, then for each
  scanned j, with r = delta − dist[j] ≥ 0, `r > B − u[colMate[j]]` and
  `r > v[j] + B`.
- `augment` reports overflow distinctly from "no path". For example, it
  returns a small enum, or sets the flag and returns false, with `assign`
  checking the flag first.
- On overflow, reset the touched labels before returning, so the scratch
  invariant holds.
- `initialize` detects empty rows and columns with a per-column "seen"
  slice instead of the `v[j] == maxT` sentinel. It does this before any
  arithmetic, so `ErrInfeasible` keeps precedence. It checks each v[j] and
  u[i] against B as they are set.
- `assign` computes the cost with `intmath.Add` over the assigned edges and
  stores it in a new field `z`. `cost()` returns `z`. Overflow in the sum
  gives `ErrOverflow`.
- `copy` carries B, L, `overflow` and `z`.

- [ ] **Step 1: Write the failing tests**

  In `lsap/overflow_test.go`:
  - `TestSparseMarginInt8`: n = 1 with one edge of cost 25 solves with cost
    25. Cost 26 is accepted by `NewSparse`, then `Assign` returns
    `ErrOverflow`. Why: B = (126 − c)/4 must be ≥ c.
  - `TestSparseMaxCostInt8`: one edge of cost 126 (the largest accepted) is
    accepted. `Assign` returns `ErrOverflow`, not `ErrInfeasible`, which
    shows that the empty-column sentinel is gone.
  - `TestSparseNeverWrongInt8` and `TestSparseNeverWrongInt16` run
    `checkNeverWrong` through `NewSparse`. The full edge list is built from
    the matrix, and a random quarter of the edges is dropped up front as
    missing edges.
  - Extend `TestNoFalseAlarmsInt32` to sparse instances:
    `instances.Sparse` with `instances.BookDegree(n)`, the same classes and
    sizes.

  In `lsap/sparse_test.go`:
  - In `TestNewSparseErrors`, replace the int8 bound cases: a cost of 127
    and a cost of −128 are rejected; 11 and −11 are accepted.
  - Delete `TestSparseOverflowBoundInt8`, which the never-wrong tests
    replace.

- [ ] **Step 2: Run the tests and confirm they fail**

  Run: `go test ./lsap/ -run 'Sparse'`
  Expected: the new tests fail under the old bound.

- [ ] **Step 3: Implement** the rules above, and update `NewSparse`'s doc
  the same way as `New`'s, with about a fifth of T's largest value in place
  of a third.

- [ ] **Step 4: Run the tests and confirm they pass**

  Run: `go test -race ./...`
  Expected: PASS.

- [ ] **Step 5: Commit**

  `lsap: detect sparse overflow while solving instead of up front (#34)`

### Task 3: CLI message, docs and benchmark check

**Files:**
- Modify: `cmd/lsap/run.go`, `cmd/lsap/run_test.go`
- Modify: `CHANGELOG.md`,
  `docs/superpowers/specs/2026-11-26-lsap-overflow-design.md`,
  `docs/superpowers/specs/2026-11-25-ap-solver-suite-design.md`

- [ ] **Step 1: Write the failing test**

  `TestOverflow` in `cmd/lsap/run_test.go`: the input
  `[[4000000000000000000]]` exits 1 with empty stdout. Stderr is exactly
  `"lsap: ap: arithmetic overflow; use a smaller -scale or smaller costs\n"`.

- [ ] **Step 2: Run the test and confirm it fails**

  Run: `go test ./cmd/lsap/ -run TestOverflow`

- [ ] **Step 3: Implement**

  - In `run.go`, when `errors.Is(err, ap.ErrOverflow)`, fail with
    `fmt.Errorf("%w; use a smaller -scale or smaller costs", err)`.
  - CHANGELOG: add a `## v0.6.0 (unreleased)` section with a "Changed"
    entry for #34. It should state the new rules, the `int32` example
    (2,145,338), and `ap.ErrOverflow`.
  - In the suite design doc, change the "Input validation" rule. Replace
    "costs whose range could overflow …" with: constructors reject costs
    that can't be represented, and `Assign` returns `ap.ErrOverflow` when
    the solve would overflow. Also update the "Big-M soundness" line to
    say that overflow is detected rather than ruled out.

- [ ] **Step 4: Benchmark against v0.5.0**

  Set `S` to the scratchpad directory the controller gives you. Run:

  ```
  git worktree add "$S/v050" v0.5.0
  go test -run '^$' -bench . -count 6 ./lsap/ > "$S/new.txt"
  (cd "$S/v050" && go test -run '^$' -bench . -count 6 ./lsap/) > "$S/old.txt"
  go run golang.org/x/perf/cmd/benchstat@latest "$S/old.txt" "$S/new.txt"
  go run ./internal/cmd/geninstance -class macholwien -n 1000 > "$S/mw.json"
  for i in 1 2 3 4 5; do go run ./internal/cmd/timesolve "$S/mw.json"; done
  (cd "$S/v050" && for i in 1 2 3 4 5; do go run ./internal/cmd/timesolve "$S/mw.json"; done)
  git worktree remove "$S/v050"
  ```

  Expected: every benchmark and the median Machol–Wien time are within 5%
  of v0.5.0. Put the benchstat table and the timings in the report. If a
  kernel misses the budget, report DONE_WITH_CONCERNS with the numbers;
  don't tune the code beyond the plan.

- [ ] **Step 5: Run the checks and commit**

  `cmd/lsap, docs: report overflow, document the new rules (#34)`
