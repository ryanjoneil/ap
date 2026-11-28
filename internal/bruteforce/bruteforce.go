// Package bruteforce provides exhaustive reference solvers for testing the
// exact solvers in this module on small instances.
//
// A mask ok restricts the allowed edges: ok == nil allows every edge, and
// ok[i][j] == false forbids the edge from row i to column j.
package bruteforce

import "slices"

// MinSum returns the minimum total cost of any assignment of the square cost
// matrix c that uses only allowed edges. feasible is false if no such
// assignment exists.
func MinSum(c [][]int64, ok [][]bool) (best int64, feasible bool) {
	return search(c, ok, func(acc, cost int64) int64 { return acc + cost })
}

// MinMax returns the minimum, over all assignments of the square cost matrix
// c that use only allowed edges, of the largest edge cost in the assignment.
// feasible is false if no such assignment exists.
func MinMax(c [][]int64, ok [][]bool) (best int64, feasible bool) {
	return search(c, ok, func(acc, cost int64) int64 { return max(acc, cost) })
}

// MaxMatching returns the size of a maximum matching in the bipartite graph
// with len(ok) rows and len(ok[0]) columns, where row i and column j are
// adjacent if ok[i][j] is true.
func MaxMatching(ok [][]bool) int {
	if len(ok) == 0 {
		return 0
	}
	used := make([]bool, len(ok[0]))

	var rec func(i int) int
	rec = func(i int) int {
		if i == len(ok) {
			return 0
		}
		best := rec(i + 1) // Leave row i unmatched.
		for j, allowed := range ok[i] {
			if allowed && !used[j] {
				used[j] = true
				best = max(best, 1+rec(i+1))
				used[j] = false
			}
		}
		return best
	}

	return rec(0)
}

// AllSums returns the total cost of every assignment of the square cost
// matrix c that uses only allowed edges, sorted in ascending order. It
// returns nil if no such assignment exists.
func AllSums(c [][]int64, ok [][]bool) []int64 {
	var sums []int64
	enumerate(c, ok, func(acc, cost int64) int64 { return acc + cost }, func(total int64) {
		sums = append(sums, total)
	})
	slices.Sort(sums)
	return sums
}

// search enumerates every assignment of c that uses only allowed edges,
// combining edge costs with combine, and returns the smallest result.
func search(
	c [][]int64,
	ok [][]bool,
	combine func(acc, cost int64) int64,
) (best int64, feasible bool) {
	enumerate(c, ok, combine, func(total int64) {
		if !feasible || total < best {
			best, feasible = total, true
		}
	})
	return best, feasible
}

// enumerate calls visit with the combined cost of every assignment of c that
// uses only allowed edges, where combine folds each edge cost into the total.
func enumerate(
	c [][]int64,
	ok [][]bool,
	combine func(acc, cost int64) int64,
	visit func(total int64),
) {
	n := len(c)
	used := make([]bool, n)

	var rec func(i int, acc int64)
	rec = func(i int, acc int64) {
		if i == n {
			visit(acc)
			return
		}
		for j := range n {
			if used[j] || (ok != nil && !ok[i][j]) {
				continue
			}
			used[j] = true
			if i == 0 {
				rec(i+1, c[i][j])
			} else {
				rec(i+1, combine(acc, c[i][j]))
			}
			used[j] = false
		}
	}

	rec(0, 0)
}
