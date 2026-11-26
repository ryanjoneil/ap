package lsap

import (
	"math/rand"
	"sort"
	"testing"
)

// The heap pops items in (d, j) order, breaking ties in d by lowest column.
func TestHeapOrder(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 100; trial++ {
		var h minHeap[int64]
		var want []heapItem[int64]
		for k := rng.Intn(50); k > 0; k-- {
			x := heapItem[int64]{d: rng.Int63n(5), j: int32(rng.Intn(20))}
			h.push(x)
			want = append(want, x)
		}
		sort.Slice(want, func(a, b int) bool { return want[a].less(want[b]) })
		for i, w := range want {
			if got := h.pop(); got != w {
				t.Fatalf("pop %v: got %v; want %v", i, got, w)
			}
		}
		if h.len() != 0 {
			t.Fatalf("len = %v after popping everything", h.len())
		}
	}
}
