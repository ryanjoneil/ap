# ap: Incremental Primal-Dual Assignment Problem Solver

This code provides a Go implementation of an incremental primal-dual assignment problem solver. It is a port of the original Fortran for the dense AP discussed in the following paper.

```
Giorgio Carpaneto, Silvano Martello, and Paolo Toth.
"Algorithms and codes for the assignment problem."
Annals of Operations Research 13, no. 1 (1988): 191-223.
```

This code can efficiently optimize any model in the form of an Assignment Problem.

```
min   sum {i,j} c_ij * x_ij
s.t.  sum {i} x_ij = 1       for all j
      sum {j} x_ij = 1       for all i
      x_ij in {0,1}          for all i, j
```

## What is this for?

While there are a number of codes available for solving APs directly, and indeed a Linear Optimization solver can also make quick work of them, most implementations are not intended for _incremental_ solution. Incremental solving of assignment problems is important when an AP forms a relaxation of some other model, such as a Traveling Salesman Problem, and variable filtering is achieved through reduced cost-based domain filtering.

In such scenarios, an initial relaxation of a more complex model into an AP is solved in `O(n^3)` time, arcs are typically removed from the AP solution in the search tree, and successive solutions that do not include those arcs are computed in `O(n^2)` time. The efficiency gained by incremental solution makes the AP relaxation effective in furthering domain propagation. See the following paper for more details.

```
Filippo Focacci, Andrea Lodi, and Michela Milano.
"A hybrid exact algorithm for the TSPTW."
INFORMS Journal on Computing 14, no. 4 (2002): 403-417.
```

## How do I use it?

An `AP` instance has the following methods and attributes.

* `Z`: current objective value
* `A`: arc objective values
* `U`: dual prices for rows
* `V`: dual prices for columns
* `Solve()`: optimizes the current AP
* `Remove(row, col, obj}`: removes an arc from the current solution, sets its objective value
* `X(row, col)`: true if the cell is in the current solution, false otherwise
* `Col(row)`: the column assigned to a row, -1 if unassigned
* `Row(col)`: the row assigned to a column, -1 if unassigned
* `RC(row, col})`: the current reduced cost for a cell
* `Size`: size of the assignment problem

### Basic Example

The following example shows how to optimize a simple 5x5 AP. First initialize the solver with the correct size (5 in this case). Set the objective value of each cell in `A`. Then simply call `Solve`. This algorithm runs in `O(n^3)` time.

```go
obj := [][]int64{
    {25, 40, 9, 15},
    {12, 33, 22, 31},
    {18, 37, 18, 30},
    {32, 11, 12, 26},
}

ap := ap.Create(len(obj))
for i, row := range obj {
    for j, val := range row {
        ap.A[i][j] = val
    }
}
ap.Solve()
```

You can now examine the solution by looking at `Z`, `U`, `V`, and by calling `RC` and `X`. See the example and test code for more details.

```
$ go run example/main.go
```
```
Z =	56
U =	[0 3 9 0]
V =	[9 11 9 15]

RC = 16 29  0  0
      0 19 10 13
      0 17  0  6
     23  0  3 11

X =   -  -  -  1
      1  -  -  -
      -  -  1  -
      -  1  -  -
```
### Incremental Example

A more complex use case removes arcs from the AP's feasible set. In this case we don't want those arcs being added back in subsequent solutions, so we set their objective values to something large (`M`) to discourage that. The first call to `solve` runs in `O(n^3)` time, while subsequent calls after an arc removal run in `O(n^2)`.

```go
var M int64 = 1000
ap.Remove(2, 2, M)
ap.Solve()
```

This gives a new solution, which can again be examined by looking at `Z`, `U`, `V`, and by calling `RC` and `X`.
