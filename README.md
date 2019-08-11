# `ap`: assignment problems

[![Build Status](https://semaphoreci.com/api/v1/ryanjoneil/ap/branches/master/badge.svg)](https://semaphoreci.com/ryanjoneil/ap)

This package provides interfaces and data structures common to formulating and solving [assignment problems](https://en.wikipedia.org/wiki/Assignment_problem), as well as production-ready codes for solving particular variants. More details about these can be found in:

```text
Rainer Burkard, Mauro Dell'Amico, and Silvano Martello.
"Assignment Problems - Revised Reprint."
Society for Industrial and Applied Mathematics (2012).
```

At this time, `ap` only provides an incremental code to solve the Linear Sum Assignment Problem. Additional forms are planned for future milestones.

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

`lsap` reads JSON input data in the form of a square cost matrix from standard input and writes an optimal permutation and cost to standard output.

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

## `lsap` Package

In order to embed an LSAP solver in other Go code, `go get` the library.

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

assignment := a.Assign() // [1 2 0]
cost := a.Cost()         // 49
```

Extensive examples are available in the module docs.

```bash
godoc -http=localhost:6060
```
