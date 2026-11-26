#!/usr/bin/env python3
"""Compare ap-lsap with SciPy and OR-Tools on LSAP instance files.

Usage: bench/.venv/bin/python bench/compare.py instance.json [...]

Each external solver runs 3 times per instance, each run in a forked child so
a run can be killed after TIMEOUT seconds (default 600). Only the solve call is
timed (time.perf_counter); arrays/arcs are built before the timer starts.
Costs of all solvers must agree, otherwise the script exits non-zero.
"""
import json
import multiprocessing as mp
import os
import statistics
import subprocess
import sys
import time

import numpy as np
from ortools.graph.python import linear_sum_assignment as ort
from scipy.optimize import linear_sum_assignment
from scipy.sparse import csr_matrix
from scipy.sparse.csgraph import min_weight_full_bipartite_matching

RUNS = 3
TIMEOUT = float(os.environ.get("COMPARE_TIMEOUT", "600"))
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TIMEOUT_CELL = "timeout (>600 s)"


def load(path):
    """Return (n, rows, cols, costs, dense) with numpy arrays."""
    with open(path) as f:
        data = json.load(f)
    if isinstance(data, list):
        m = np.array(data, dtype=np.int64)
        n = m.shape[0]
        return n, None, None, None, m
    edges = np.array(data["edges"], dtype=np.int64)
    return data["n"], edges[:, 0], edges[:, 1], edges[:, 2], None


# Each prepare function builds solver inputs (untimed) and returns a thunk
# that does the timed solve and returns the total cost.
def prep_scipy_dense(inst):
    n, _, _, _, m = inst

    def solve():
        r, c = linear_sum_assignment(m)
        return int(m[r, c].sum())

    return solve


def prep_scipy_sparse(inst):
    n, u, v, c, m = inst
    if m is not None:
        u, v = np.divmod(np.arange(n * n), n)
        c = m.ravel()
    # Shift by +1: explicit zeros would be dropped as missing edges.
    g = csr_matrix((c + 1, (u, v)), shape=(n, n))

    def solve():
        r, cc = min_weight_full_bipartite_matching(g)
        return int(g[r, cc].sum()) - n

    return solve


def prep_ortools(inst):
    n, u, v, c, m = inst
    if m is not None:
        u, v = np.divmod(np.arange(n * n, dtype=np.int64), n)
        c = m.ravel()

    def solve():
        a = ort.SimpleLinearSumAssignment()
        a.add_arcs_with_cost(u, v, c)
        st = a.solve()
        if st != a.OPTIMAL:
            raise RuntimeError(f"ortools status {st}")
        return int(a.optimal_cost())

    return solve


def child(solve, conn):
    try:
        t0 = time.perf_counter()
        cost = solve()
        conn.send(("ok", time.perf_counter() - t0, cost))
    except Exception as e:  # recorded in the table, not fatal
        conn.send(("err", 0.0, f"{type(e).__name__}: {e}"))
    conn.close()


def run_once(solve):
    ctx = mp.get_context("fork")
    parent, kid = ctx.Pipe(duplex=False)
    p = ctx.Process(target=child, args=(solve, kid))
    p.start()
    kid.close()
    if parent.poll(TIMEOUT):
        res = parent.recv()
    else:
        res = ("timeout", 0.0, None)
        p.kill()
    p.join()
    return res


def measure(solve):
    """Return (cell text, cost or None, failure or None)."""
    times, cost = [], None
    for _ in range(RUNS):
        st, t, x = run_once(solve)
        if st == "timeout":
            return TIMEOUT_CELL, None, None
        if st == "err":
            return "error: " + x, None, None
        times.append(t)
        if cost is not None and cost != x:
            return "INCONSISTENT", None, f"cost varied across runs: {cost} vs {x}"
        cost = x
    return f"{statistics.median(times):.3f}", cost, None


def ap_lsap(path):
    times, cost = [], None
    for _ in range(RUNS):
        try:
            out = subprocess.run(
                ["go", "run", "./internal/cmd/timesolve", path], cwd=ROOT,
                capture_output=True, text=True, timeout=TIMEOUT, check=True,
            ).stdout
        except subprocess.TimeoutExpired:
            return TIMEOUT_CELL, None
        except subprocess.CalledProcessError as e:
            return "error: " + e.stderr.strip(), None
        r = json.loads(out)
        times.append(r["seconds"])
        cost = r["cost"]
    return f"{statistics.median(times):.3f}", cost


def main(paths):
    print("| instance | n | edges | ap-lsap (s) | scipy dense (s) | scipy sparse (s) | ortools (s) |")
    print("|---|---|---|---|---|---|---|")
    failures = []
    for path in paths:
        inst = load(path)
        n, u, _, _, m = inst
        dense = m is not None
        edges = n * n if dense else len(u)
        name = os.path.basename(path)
        cells, costs = {}, {}
        cells["ap"], costs["ap"] = ap_lsap(path)
        solvers = [("scipy dense", prep_scipy_dense if dense else None),
                   ("scipy sparse", prep_scipy_sparse),
                   ("ortools", prep_ortools)]
        for label, prep in solvers:
            if prep is None:
                cells[label] = "-"
                continue
            cells[label], costs[label], fail = measure(prep(inst))
            if fail:
                failures.append(f"{name}: {label}: {fail}")
        ref = costs.get("ap")
        for label, c in costs.items():
            if c is not None and ref is not None and c != ref:
                failures.append(f"{name}: COST MISMATCH {label}={c} ap-lsap={ref}")
        print(f"| {name} | {n} | {edges} | {cells['ap']} | {cells['scipy dense']} "
              f"| {cells['scipy sparse']} | {cells['ortools']} |", flush=True)
        print(f"{name}: costs {costs}", file=sys.stderr, flush=True)
    for f in failures:
        print("FAILURE:", f, file=sys.stderr)
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
