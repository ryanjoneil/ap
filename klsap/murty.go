package klsap

import (
	"container/heap"
	"errors"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

// node is a solved subproblem of Murty's method: its solver a, which carries
// the subproblem's required and forbidden edges, its optimal assignment p and
// cost, and the rows whose edges it doesn't require, in increasing order.
type node[T ap.Integer] struct {
	a    *lsap.LSAP[T]
	p    ap.Permutation
	cost T
	free []int
	seq  uint64 // creation order, which breaks ties in cost

	minIdx, maxIdx int // positions in the queue's two heaps
}

// before orders nodes by cost, then by creation order, so that the ranking is
// deterministic when costs tie.
func (x *node[T]) before(y *node[T]) bool {
	return x.cost < y.cost || (x.cost == y.cost && x.seq < y.seq)
}

// split is the branching phase of Murty's method (book §5.4.1, Algorithm
// 5.1). Let i_1 < i_2 < ... < i_m be the free rows of nd. For h = 1, ..., m−1,
// child h requires the edges (i_1, p(i_1)), ..., (i_{h−1}, p(i_{h−1})) and
// forbids (i_h, p(i_h)), on top of nd's own fixes. The children partition the
// assignments of nd other than nd.p, so the ranking never repeats an
// assignment. split solves each child and offers it to the queue, dropping
// infeasible ones. If solving a child fails otherwise, as with
// [ap.ErrOverflow], split stops and returns the error.
//
// split creates no child for the last free row i_m: with every other row
// required, row i_m has only column p(i_m) left, and that is forbidden. This
// follows Algorithm 5.1, which creates n − |I| − 1 children, and the paper,
// whose abstract counts at most n − 1 subproblems per assignment.
//
// Departures from Murty's method as Algorithm 5.1 and the paper give it:
//
//   - They solve each child from scratch, as an assignment problem on the
//     rows and columns that aren't required (of sizes 2 through n, in the
//     paper's abstract). Here each child is a Copy of the parent's lsap
//     solver, which shares the parent's costs and starts from its optimal
//     assignment and dual prices. Forbidding the assigned edge
//     (i_h, p(i_h)) leaves only row i_h unassigned, so lsap solves the child
//     incrementally with a single augmentation, O(n²) dense, rather than a
//     full solve. Book §5.4.1 attributes this reoptimization, one
//     shortest-path augmentation per child from the parent's dual prices,
//     to Pedersen, Nielsen and Andersen (book reference [565]; see the
//     package doc).
//   - The popped node's solver is reused as the working solver from which
//     the children are copied, rather than copying it once more: requiring
//     (i_h, p(i_h)) in it after each child leaves nd.p optimal, so it needs no
//     re-solve. nd is never offered again, so its solver is free to change.
func (a *KLSAP[T]) split(nd *node[T]) error {
	w := nd.a
	for h, i := range nd.free[:len(nd.free)-1] {
		child := w.Copy()
		child.Fix(i, nd.p[i], false)
		p, err := child.Assign()
		switch {
		case errors.Is(err, ap.ErrInfeasible):
		case err != nil:
			return err
		default:
			a.offer(&node[T]{a: child, p: p, cost: child.Cost(), free: nd.free[h:]})
		}
		w.Fix(i, nd.p[i], true)
	}
	return nil
}

// offer adds nd to the queue, keeping at most k − r nodes, where r is the
// number of assignments returned so far.
//
// Departure from Murty's method: Algorithm 5.1 keeps every waiting
// subproblem in its set Q. After r assignments have been returned, only the
// k − r best waiting nodes can contribute to the ranking, since every
// assignment of a worse node costs at least as much as each of the k − r
// nodes' own assignments. So offer keeps the k − r best by (cost, seq) and
// drops the rest, which holds memory to O(k) subproblems. Each pop lowers
// both the queue's size and its capacity by one, so only offers need the
// check.
func (a *KLSAP[T]) offer(nd *node[T]) {
	nd.seq = a.seq
	a.seq++
	if a.queue.len() < a.k-a.r {
		a.queue.push(nd)
		return
	}
	if nd.before(a.queue.worst()) {
		a.queue.popMax()
		a.queue.push(nd)
	}
}

// queue is a double-ended priority queue of nodes ordered by (cost, seq),
// built from a min-heap and a max-heap that hold the same nodes. Each node
// records its position in both, so that popping from one heap can remove the
// node from the other in O(log size) time.
type queue[T ap.Integer] struct {
	min minHeap[T]
	max maxHeap[T]
}

func (q *queue[T]) len() int { return len(q.min) }

// worst returns the worst node. The queue must not be empty.
func (q *queue[T]) worst() *node[T] { return q.max[0] }

func (q *queue[T]) push(nd *node[T]) {
	heap.Push(&q.min, nd)
	heap.Push(&q.max, nd)
}

// popMin removes and returns the best node.
func (q *queue[T]) popMin() *node[T] {
	nd := heap.Pop(&q.min).(*node[T])
	heap.Remove(&q.max, nd.maxIdx)
	return nd
}

// popMax removes the worst node.
func (q *queue[T]) popMax() {
	nd := heap.Pop(&q.max).(*node[T])
	heap.Remove(&q.min, nd.minIdx)
}

// minHeap is a heap.Interface with the best node at the top.
type minHeap[T ap.Integer] []*node[T]

func (h minHeap[T]) Len() int           { return len(h) }
func (h minHeap[T]) Less(i, j int) bool { return h[i].before(h[j]) }

func (h minHeap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].minIdx, h[j].minIdx = i, j
}

func (h *minHeap[T]) Push(x any) {
	nd := x.(*node[T])
	nd.minIdx = len(*h)
	*h = append(*h, nd)
}

func (h *minHeap[T]) Pop() any {
	s := *h
	last := len(s) - 1
	nd := s[last]
	s[last] = nil
	*h = s[:last]
	return nd
}

// maxHeap is a heap.Interface with the worst node at the top.
type maxHeap[T ap.Integer] []*node[T]

func (h maxHeap[T]) Len() int           { return len(h) }
func (h maxHeap[T]) Less(i, j int) bool { return h[j].before(h[i]) }

func (h maxHeap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].maxIdx, h[j].maxIdx = i, j
}

func (h *maxHeap[T]) Push(x any) {
	nd := x.(*node[T])
	nd.maxIdx = len(*h)
	*h = append(*h, nd)
}

func (h *maxHeap[T]) Pop() any {
	s := *h
	last := len(s) - 1
	nd := s[last]
	s[last] = nil
	*h = s[:last]
	return nd
}
