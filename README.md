# `ap`: assignment problem solvers for go

This package provides interfaces and data structures common to formulating and
solving [assignment problems](https://en.wikipedia.org/wiki/Assignment_problem),
as well as production-ready codes for solving particular variants. More details
about these can be found in:

```text
Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
"Assignment Problems - Revised Reprint."
Society for Industrial and Applied Mathematics (2012).
```

At this time, `ap` only provides an incremental code to solve the Linear Sum
Assignment Problem. Additional forms are planned for future milestones.

LSAPs take the following form:

```text
min   ∑_i,j c_ij * x_ij
s.t.  ∑_i   x_ij = 1      ∀ j
      ∑_j   x_ij = 1      ∀ i
            x_ij ∈ {0,1}  ∀ i,j
```

## Quick Start: CLI

To solve LSAPs from the command line, first install the `lsap` binary using Go.

```bash
go install github.com/ryanjoneil/ap/cmd/lsap
```

`lsap` reads JSON input data in the form of a square cost matrix from standard
input and writes an optimal permutation and cost to standard output.

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
  "assignment": [
    3,
    2,
    1,
    0
  ],
  "cost": 265
}
```

## Quick Start: Packages

Extensive examples are available in the module docs.

```bash
godoc -http=localhost:6060
```

### `ap`: assignment representations & interfaces

Package `ap` provides solution representations and interfaces for working with
assignment problems and solvers.

```bash
go get github.com/ryanjoneil/ap
```

The default representation of an assignment produced by an `Assigner` is a
`Permutation`.

```go
a := SomeAssigner{} // implements ap.Assign
p := a.Assign()     // p is an ap.Permutation
```

Permutations can be converted to cyclic and matrix representations of
assignments, and vice versa. All representations provide `Inverse` methods
reverse the direction of assignment.

```go
p := ap.Permutation{1, 0, 2, 6, 5, 3, 4}
p.Cycles()  // {{0, 1}, {2}, {3, 6, 4, 5}}
p.Inverse() // {1, 0, 2, 5, 6, 4, 3}
p.Matrix()  // p[u] == v -> m[u][v] == true
```

### `ap/lsap`: linear sum assignment problem solver

Package `ap/lsap` provides a efficient, iterative implementation of a
primal-dual linear sum assignment problem solver.

```bash
go get github.com/ryanjoneil/ap/lsap
```

LSAPs are easy to construct and solve from a cost matrix.

```go
a := lsap.New([][]int64{
    {10, 15, 12},
    {51, 75, 23},
    {11, 91, 10},
})

permutation := a.Assign() // [1 2 0]
cost := a.Cost()          // 49
```

`lsap` provides command line flags for outputting dual prices and reduced costs.

```bash
lsap -h
```

```text
lsap solves linear sum assignment problems, given a square cost matrix
Usage:
        lsap < input.json -dual -rc > output.json
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
```
