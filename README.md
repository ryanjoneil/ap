# `ap`: assignment problem solvers for go

This package provides data structures common to formulating and solving
[assignment problems][ap], as well as codes for solving three variants: the
linear sum assignment problem (LSAP), maximum cardinality bipartite matching
(MCM), and the linear bottleneck assignment problem (LBAP). More details about
these can be found in:

```text
Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
"Assignment Problems - Revised Reprint."
Society for Industrial and Applied Mathematics (2012).
```

LSAPs take the following form:

```text
min   ∑_i,j c_ij * x_ij
s.t.  ∑_i   x_ij = 1      ∀ j
      ∑_j   x_ij = 1      ∀ i
            x_ij ∈ {0,1}  ∀ i,j
```

LBAPs have the same constraints, but minimize the largest assigned cost,
`min_φ max_i c_iφ(i)`, over permutations φ. MCM finds as many edges as
possible in a bipartite graph, no two of which share a vertex.

## Quick Start: CLI

There is one binary per solver: `lsap`, `mcm` and `lbap`. Install them using
Go.

```bash
go install github.com/ryanjoneil/ap/cmd/lsap@latest
go install github.com/ryanjoneil/ap/cmd/mcm@latest
go install github.com/ryanjoneil/ap/cmd/lbap@latest
```

All three read one JSON value from standard input and write one JSON object to
standard output. A top-level array is dense input, and a top-level object is
sparse input. The output always has a `"status"` of `"optimal"` or
`"infeasible"`, and both exit 0. Invalid input or a bad flag writes
`<binary>: <message>` to standard error, writes nothing to standard output and
exits 1.

### `lsap`

`lsap` reads a square cost matrix and writes an optimal permutation and cost.

```bash
cat <<EOF | lsap | jq
[
    [  90,  76,  75,  70 ],
    [  35,  85,  55,  65 ],
    [ 125,  95,  90, 105 ],
    [  45, 110,  95, 115 ]
]
EOF
```

```json
{
  "status": "optimal",
  "cost": 265,
  "permutation": [
    3,
    2,
    1,
    0
  ]
}
```

Sparse input is an object with `n` and a list of `[u, v, cost]` edges. An edge
that is not listed is forbidden.

```bash
echo '{"n":3,"edges":[[0,0,4],[0,1,1],[1,0,2],[1,2,6],[2,1,3],[2,2,5]]}' | lsap
```

```json
{"status":"optimal","cost":8,"permutation":[1,0,2]}
```

If there is no perfect assignment, the status says so.

```bash
echo '{"n":2,"edges":[[0,0,1],[1,0,2]]}' | lsap
```

```json
{"status":"infeasible"}
```

Costs must be integers. Use `-scale k` to multiply each cost by `k` and round
it to an integer; reported costs, dual prices and reduced costs are divided by
`k`. Without it, a non-integer cost is an error.

```bash
echo '[[1.5,2],[2,1.5]]' | lsap -scale 10
```

```json
{"status":"optimal","cost":3,"permutation":[0,1]}
```

```bash
lsap -h
```

```text
lsap solves linear sum assignment problems, given a square cost matrix.

Input is a dense cost matrix, a JSON array of arrays, or a sparse problem,
a JSON object {"n": N, "edges": [[u, v, cost], ...]}. Costs must be integers
unless -scale k is given, which multiplies each cost by k and rounds it to an
integer. Reported costs, dual prices and reduced costs are divided by k.
The output has a "status" of "optimal" or "infeasible".

Usage:

lsap < input.json -dual -rc > output.json
echo '[[1.5,2],[2,1.5]]' | lsap -scale 10

cat <<EOF | lsap | jq
[
	[  90,  76,  75,  70 ],
	[  35,  85,  55,  65 ],
	[ 125,  95,  90, 105 ],
	[  45, 110,  95, 115 ]
]
EOF

Flags:
  -cycles
    	output cyclic assignment form
  -dual
    	output dual prices
  -matrix
    	output matrix assignment form
  -rc
    	output reduced cost matrix
  -scale k
    	multiply costs by k and round to integers (accepts non-integer costs)
```

### `mcm`

`mcm` reads a boolean adjacency matrix, with one row per vertex in U and one
column per vertex in V, or a sparse graph `{"nu": NU, "nv": NV, "edges":
[[u, v], ...]}`. The two sides may differ in size. It writes the size of a
maximum matching, whether it is perfect, and for each u the matched v, or -1.

```bash
echo '[[true,false,true],[true,true,false],[false,true,false]]' | mcm
```

```json
{"status":"optimal","size":3,"perfect":true,"matching":[2,0,1]}
```

```bash
echo '{"nu":2,"nv":3,"edges":[[0,0],[1,0],[1,2]]}' | mcm
```

```json
{"status":"optimal","size":2,"perfect":false,"matching":[0,2]}
```

```bash
mcm -h
```

```text
mcm finds a maximum cardinality matching (MCM) in a bipartite graph.

Input is a dense boolean adjacency matrix, a JSON array of arrays with one row
per vertex in U and one column per vertex in V, or a sparse graph, a JSON
object {"nu": NU, "nv": NV, "edges": [[u, v], ...]}. The two sides may differ
in size. The output has a "status" of "optimal", the "size" of the matching,
whether it is "perfect" (both sides the same size and fully matched), and
"matching", where entry u is the vertex in V matched to u, or -1.
-cycles and -matrix add the permutation forms, only for a perfect matching.

Usage:

mcm < input.json -cycles > output.json

cat <<EOF | mcm | jq
[
	[ true,  false, true  ],
	[ true,  true,  false ],
	[ false, true,  false ]
]
EOF

Flags:
  -cycles
    	output cyclic assignment form (perfect matchings only)
  -matrix
    	output matrix assignment form (perfect matchings only)
```

### `lbap`

`lbap` takes the same input as `lsap`, but minimizes the largest cost used in
the assignment. The reported cost is that bottleneck value.

```bash
echo '[[90,76,75,70],[35,85,55,65],[125,95,90,105],[45,110,95,115]]' | lbap
```

```json
{"status":"optimal","cost":90,"permutation":[3,1,2,0]}
```

```bash
lbap -h
```

```text
lbap solves linear bottleneck assignment problems, given a square cost
matrix. It minimizes the largest cost used in a perfect assignment.

Input is a dense cost matrix, a JSON array of arrays, or a sparse problem,
a JSON object {"n": N, "edges": [[u, v, cost], ...]}. Costs must be integers
unless -scale k is given, which multiplies each cost by k and rounds it to an
integer. The reported cost, the bottleneck value, is divided by k.
The output has a "status" of "optimal" or "infeasible".

Usage:

lbap < input.json -cycles > output.json
echo '[[1.5,2],[2,1.5]]' | lbap -scale 10

cat <<EOF | lbap | jq
[
	[  90,  76,  75,  70 ],
	[  35,  85,  55,  65 ],
	[ 125,  95,  90, 105 ],
	[  45, 110,  95, 115 ]
]
EOF

Flags:
  -cycles
    	output cyclic assignment form
  -matrix
    	output matrix assignment form
  -scale k
    	multiply costs by k and round to integers (accepts non-integer costs)
```

## Quick Start: Packages

Examples are available in the [package docs][docs].

### `ap`: assignment representations

Package `ap` provides the types shared by the solvers: `Permutation`, `Cycles`
and `Matrix` representations of an assignment; `Edge` and `Pair` for sparse
input; `DualPrices`; the `Integer` constraint that costs are generic over; and
`ErrInfeasible`.

```bash
go get github.com/ryanjoneil/ap
```

The default representation of an assignment is a `Permutation`, where `p[u]`
is the column assigned to row `u`. Permutations can be converted to cyclic and
matrix representations of assignments, and vice versa. All representations
provide `Inverse` methods that reverse the direction of assignment.

```go
p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
p.Cycles()  // {{0, 1}, {2}, {3, 6, 4, 5}}
p.Inverse() // {1, 0, 2, 5, 6, 4, 3}
p.Matrix()  // p[u] == v -> m[u][v] == true
```

Every solver's constructors return an error for bad input, and its `Assign`
returns `ap.ErrInfeasible` when there is no feasible assignment. Constructors
copy their input.

### `ap/lsap`: linear sum assignment problem solver

Package `ap/lsap` provides a primal-dual solver, with a dense variant, built
from a complete cost matrix, and a sparse variant, built from a list of edges.

```bash
go get github.com/ryanjoneil/ap/lsap
```

```go
a, err := lsap.New([][]int64{
    {10, 15, 12},
    {51, 75, 23},
    {11, 91, 10},
})
if err != nil {
    return err
}

p, err := a.Assign() // [1 2 0]
cost := a.Cost()     // 49
```

Both variants report dual prices and reduced costs, and both are incremental:
`Remove` forbids an edge, and the next `Assign` re-solves from the previous
solution rather than starting over. `ReducedCost` returns `(T, bool)`, where
the bool is false for an edge that is missing or removed. `Copy` returns an
independent solver, which makes it easy to explore what-if removals.

```go
a, err := lsap.NewSparse(3, []ap.Edge[int64]{
    {U: 0, V: 0, Cost: 4}, {U: 0, V: 1, Cost: 1},
    {U: 1, V: 0, Cost: 2}, {U: 1, V: 2, Cost: 6},
    {U: 2, V: 1, Cost: 3}, {U: 2, V: 2, Cost: 5},
})
if err != nil {
    return err
}

p, err := a.Assign()      // [1 0 2], cost 8
rc, ok := a.ReducedCost(0, 0) // 2, true

b := a.Copy()
b.Remove(0, 1)
p, err = b.Assign()       // [0 2 1], cost 13; a is unchanged
b.Remove(2, 1)
_, err = b.Assign()       // ap.ErrInfeasible
```

### `ap/mcm`: maximum cardinality matching solver

Package `ap/mcm` finds a maximum cardinality matching in a bipartite graph,
given as a boolean adjacency matrix with `New` or as edges with `NewSparse`.
It also supports `Remove`, incremental re-solve and `Copy`.

```bash
go get github.com/ryanjoneil/ap/mcm
```

```go
m, err := mcm.NewSparse(2, 3, []ap.Pair{{U: 0, V: 0}, {U: 1, V: 0}, {U: 1, V: 2}})
if err != nil {
    return err
}

m.Match() // [0 2]: entry u is the column matched to u, or -1
m.Size()  // 2
```

`Assign` returns the matching as a `Permutation` when it is perfect, and
`ap.ErrInfeasible` otherwise.

### `ap/lbap`: linear bottleneck assignment problem solver

Package `ap/lbap` solves the bottleneck variant with a threshold algorithm. It
offers `New`, `NewSparse`, `Assign`, `Cost`, `Remove` and `Copy`, like `lsap`.
Many assignments are often optimal for a bottleneck problem. Which one is
returned is deterministic but otherwise unspecified.

```bash
go get github.com/ryanjoneil/ap/lbap
```

```go
a, err := lbap.New([][]int{
    {10, 15, 12},
    {51, 75, 23},
    {11, 91, 10},
})
if err != nil {
    return err
}

p, err := a.Assign() // [1 2 0]
cost := a.Cost()     // 23
```

## Benchmarks

`bench/` compares `lsap` with SciPy and OR-Tools on the instance classes of
the book's Section 4.10.1. See [bench/RESULTS.md](bench/RESULTS.md) for the
results and [bench/README.md](bench/README.md) for how to reproduce them.

## References

Each package's documentation describes how it uses these.

```text
Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
"Assignment Problems - Revised Reprint."
Society for Industrial and Applied Mathematics (2012).

Giorgio Carpaneto, Silvano Martello, and Paolo Toth.
"Algorithms and codes for the assignment problem."
In B. Simeone, P. Toth, G. Gallo, F. Maffioli, and S. Pallottino,
editors, "Fortran Codes for Network Optimization."
Annals of Operations Research 13 (1988): 193-223. Baltzer, Basel.

R. Garfinkel.
"An improved algorithm for the bottleneck assignment problem."
Operations Research 19 (1971): 1747–1751.

A.V. Goldberg and R. Kennedy. An efficient cost scaling algorithm for
the assignment problem. Math. Program., 71:153-177, 1995.

John E. Hopcroft and Richard M. Karp.
"An n^5/2 algorithm for maximum matchings in bipartite graphs."
SIAM Journal on Computing 2 (1973): 225–231.

Roy Jonker and Anton Volgenant.
"A shortest augmenting path algorithm for dense and sparse linear
assignment problems."
Computing 38 (1987): 325-340.

R.E. Machol and M. Wien. A hard assignment problem. Oper. Res.,
24:190-192, 1976.

R.E. Machol and M. Wien. Errata. Oper. Res., 25:364, 1977.
```

[ap]:   https://en.wikipedia.org/wiki/Assignment_problem
[docs]: https://pkg.go.dev/github.com/ryanjoneil/ap
