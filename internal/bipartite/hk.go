// Package bipartite implements the Hopcroft–Karp maximum-cardinality bipartite
// matching kernel shared by the matching and lbap solvers.
//
// References:
//
//   - J. Hopcroft and R.M. Karp, "An n^5/2 algorithm for maximum matchings in
//     bipartite graphs", SIAM J. Comput. 2:225–231 (1973) (book bibliography
//     entry [389]).
//   - R. Burkard, M. Dell'Amico and S. Martello, Assignment Problems (SIAM,
//     2012), §3.3, Algorithm 3.2 (Hopcroft_Karp).
package bipartite

const inf = int(^uint(0) >> 1)

// State is a matching between U (rows) and V (columns). Row u's allowed
// columns are adj[start[u]:end[u]] in every call.
type State struct {
	MateU []int // len nU; −1 = free
	MateV []int // len nV; −1 = free

	dist   []int // BFS layer of each row
	queue  []int
	cursor []int // per-row next edge index in the DFS
	stack  []int
}

// New returns an empty matching on nU rows and nV columns.
func New(nU, nV int) *State {
	s := &State{
		MateU:  make([]int, nU),
		MateV:  make([]int, nV),
		dist:   make([]int, nU),
		queue:  make([]int, 0, nU),
		cursor: make([]int, nU),
		stack:  make([]int, 0, nU),
	}
	for i := range s.MateU {
		s.MateU[i] = -1
	}
	for i := range s.MateV {
		s.MateV[i] = -1
	}
	return s
}

// Size returns the number of matched pairs.
func (s *State) Size() int {
	n := 0
	for _, v := range s.MateU {
		if v >= 0 {
			n++
		}
	}
	return n
}

// Unmatch frees row u and its mate. It is a no-op if u is free.
func (s *State) Unmatch(u int) {
	if v := s.MateU[u]; v >= 0 {
		s.MateV[v] = -1
		s.MateU[u] = -1
	}
}

// Clone returns an independent copy of the matching. Scratch space is not
// shared.
func (s *State) Clone() *State {
	c := New(len(s.MateU), len(s.MateV))
	copy(c.MateU, s.MateU)
	copy(c.MateV, s.MateV)
	return c
}

// Greedy matches each free row u, in increasing order, to the first free
// column in its window. Existing pairs are never changed.
func (s *State) Greedy(start, end, adj []int) {
	for u := range s.MateU {
		if s.MateU[u] >= 0 {
			continue
		}
		for e := start[u]; e < end[u]; e++ {
			if v := adj[e]; s.MateV[v] < 0 {
				s.MateU[u], s.MateV[v] = v, u
				break
			}
		}
	}
}

// Maximize grows the current matching to maximum cardinality with the
// Hopcroft–Karp algorithm and returns its size: J. Hopcroft and R.M. Karp,
// "An n^5/2 algorithm for maximum matchings in bipartite graphs", SIAM J.
// Comput. 2:225–231 (1973) [book ref. 389]; Burkard, Dell'Amico & Martello,
// Assignment Problems (SIAM 2012), §3.3, Algorithm 3.2 (Hopcroft_Karp).
//
// The starting matching may be any valid one: MateU and MateV consistent and
// every matched edge present in the current windows. Maximize does not repair
// an invalid matching; the caller must Unmatch any row whose mate is no longer
// in its window first.
//
// Departures from Algorithm 3.2: the matching is warm-started from the current
// state rather than the empty one; layers are stored as row distances (columns
// are reached through MateV); and the layered DFS (Find_set_M) is iterative,
// with a per-row edge cursor and an explicit stack, so deep augmenting paths
// cannot overflow the goroutine stack. Edges are scanned in window order and
// free rows in increasing index, so the result is deterministic.
func (s *State) Maximize(start, end, adj []int) int {
	for s.phase(start, end, adj) {
	}
	return s.Size()
}

// phase runs one BFS plus a maximal set of vertex-disjoint shortest augmenting
// paths. It reports whether any path was found.
func (s *State) phase(start, end, adj []int) bool {
	// Construct_layered_graph: BFS over rows, stopping after the first layer
	// that reaches a free column.
	q := s.queue[:0]
	for u, m := range s.MateU {
		if m < 0 {
			s.dist[u] = 0
			q = append(q, u)
		} else {
			s.dist[u] = inf
		}
	}
	limit := inf
	for h := 0; h < len(q); h++ {
		u := q[h]
		if s.dist[u] >= limit {
			break
		}
		for e := start[u]; e < end[u]; e++ {
			w := s.MateV[adj[e]]
			if w < 0 {
				limit = s.dist[u]
			} else if s.dist[w] == inf {
				s.dist[w] = s.dist[u] + 1
				q = append(q, w)
			}
		}
	}
	s.queue = q
	if limit == inf {
		return false
	}

	// Find_set_M: iterative layered DFS from each free row.
	found := false
	for u0, m := range s.MateU {
		if m >= 0 || s.dist[u0] != 0 {
			continue
		}
		stack := append(s.stack[:0], u0)
		s.cursor[u0] = start[u0]
		for len(stack) > 0 {
			x := stack[len(stack)-1]
			if s.cursor[x] >= end[x] {
				s.dist[x] = inf // dead end
				stack = stack[:len(stack)-1]
				continue
			}
			v := adj[s.cursor[x]]
			s.cursor[x]++
			w := s.MateV[v]
			switch {
			case w < 0:
				if s.dist[x] != limit {
					continue
				}
				// Augment along the stack: each row's last scanned edge is
				// the column leading to the next row.
				for _, r := range stack {
					c := adj[s.cursor[r]-1]
					s.MateU[r], s.MateV[c] = c, r
					s.dist[r] = inf // rows are used by one path per phase
				}
				found = true
				stack = stack[:0]
			case s.dist[w] == s.dist[x]+1 && s.dist[w] <= limit:
				s.cursor[w] = start[w]
				stack = append(stack, w)
			}
		}
		s.stack = stack
	}
	return found
}
