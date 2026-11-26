package lsap

import "github.com/ryanjoneil/ap"

// heapItem is a column j with tentative distance d.
type heapItem[T ap.Integer] struct {
	d T
	j int32
}

// less orders items by distance, then by column, so that ties are broken
// deterministically in favor of the lowest column.
func (a heapItem[T]) less(b heapItem[T]) bool {
	return a.d < b.d || (a.d == b.d && a.j < b.j)
}

// minHeap is a binary min-heap of heapItems ordered by (d, j). It is typed,
// rather than built on container/heap, to avoid interface calls in the
// shortest path search.
type minHeap[T ap.Integer] []heapItem[T]

func (h minHeap[T]) len() int { return len(h) }

// push adds x to the heap.
func (h *minHeap[T]) push(x heapItem[T]) {
	*h = append(*h, x)
	s := *h
	i := len(s) - 1
	for i > 0 {
		p := (i - 1) / 2
		if !x.less(s[p]) {
			break
		}
		s[i] = s[p]
		i = p
	}
	s[i] = x
}

// pop removes and returns the smallest item. The heap must not be empty.
func (h *minHeap[T]) pop() heapItem[T] {
	s := *h
	top := s[0]
	n := len(s) - 1
	x := s[n]
	s = s[:n]
	*h = s
	if n == 0 {
		return top
	}
	i := 0
	for {
		c := 2*i + 1
		if c >= n {
			break
		}
		if r := c + 1; r < n && s[r].less(s[c]) {
			c = r
		}
		if !s[c].less(x) {
			break
		}
		s[i] = s[c]
		i = c
	}
	s[i] = x
	return top
}
