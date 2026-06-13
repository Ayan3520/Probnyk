package solver

import (
	"fmt"
	"sort"
	"strings"

	"lem-in/internal/graph"
)

// Solve finds the optimal set of vertex-disjoint paths and simulates
// the movement of all ants. Returns one formatted string per turn.
func Solve(farm *graph.Farm) ([]string, error) {
	if farm.Start == farm.End {
		return []string{}, nil
	}

	paths, err := findBestPaths(farm)
	if err != nil {
		return nil, err
	}

	return simulate(farm.NumAnts, paths, farm.Start, farm.End), nil
}

// ─── Residual graph (node-split, Edmonds-Karp) ────────────────────────────

// edge represents a directed edge in the residual graph.
type edge struct {
	to, rev int // target node index, reverse edge index
	cap     int
}

type flowGraph struct {
	g    [][]edge
	idx  map[string]int // room-name variant → node index
	n    int
}

// nodeIn / nodeOut map a room name to its two split-node indices.
func (fg *flowGraph) nodeIn(name string) int  { return fg.idx[name+"_in"] }
func (fg *flowGraph) nodeOut(name string) int { return fg.idx[name+"_out"] }

func newFlowGraph(farm *graph.Farm) *flowGraph {
	fg := &flowGraph{idx: make(map[string]int)}
	// Assign indices
	for name := range farm.Rooms {
		fg.idx[name+"_in"] = fg.n
		fg.n++
		fg.idx[name+"_out"] = fg.n
		fg.n++
	}
	fg.g = make([][]edge, fg.n)
	return fg
}

func (fg *flowGraph) addEdge(u, v, cap int) {
	fg.g[u] = append(fg.g[u], edge{v, len(fg.g[v]), cap})
	fg.g[v] = append(fg.g[v], edge{u, len(fg.g[u]) - 1, 0})
}

func buildFlowGraph(farm *graph.Farm) (*flowGraph, int, int) {
	fg := newFlowGraph(farm)

	inf := farm.NumAnts + 1
	for name := range farm.Rooms {
		in, out := fg.nodeIn(name), fg.nodeOut(name)
		if name == farm.Start || name == farm.End {
			fg.addEdge(in, out, inf)
		} else {
			fg.addEdge(in, out, 1)
		}
	}
	for name, room := range farm.Rooms {
		for _, nb := range room.Links {
			fg.addEdge(fg.nodeOut(name), fg.nodeIn(nb), 1)
		}
	}

	return fg, fg.nodeOut(farm.Start), fg.nodeIn(farm.End)
}

// bfs returns a level array for the BFS level graph (Dinic-style), or nil if sink unreachable.
func bfs(fg *flowGraph, s, t int) []int {
	level := make([]int, fg.n)
	for i := range level {
		level[i] = -1
	}
	level[s] = 0
	q := []int{s}
	for len(q) > 0 {
		v := q[0]
		q = q[1:]
		for _, e := range fg.g[v] {
			if e.cap > 0 && level[e.to] < 0 {
				level[e.to] = level[v] + 1
				q = append(q, e.to)
			}
		}
	}
	if level[t] < 0 {
		return nil
	}
	return level
}

// dfs finds an augmenting path and returns flow pushed (1 or 0).
func dfs(fg *flowGraph, v, t, f int, level []int, iter []int) int {
	if v == t {
		return f
	}
	for ; iter[v] < len(fg.g[v]); iter[v]++ {
		e := &fg.g[v][iter[v]]
		if e.cap > 0 && level[v] < level[e.to] {
			d := dfs(fg, e.to, t, min(f, e.cap), level, iter)
			if d > 0 {
				e.cap -= d
				fg.g[e.to][e.rev].cap += d
				return d
			}
		}
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxFlow runs Dinic's algorithm and returns total flow.
func maxFlow(fg *flowGraph, s, t int) int {
	flow := 0
	for {
		level := bfs(fg, s, t)
		if level == nil {
			break
		}
		iter := make([]int, fg.n)
		for {
			f := dfs(fg, s, t, 1<<30, level, iter)
			if f == 0 {
				break
			}
			flow += f
		}
	}
	return flow
}

// ─── Path extraction from flow ────────────────────────────────────────────

// extractPaths reads which edges are used (capacity reduced from original)
// and reconstructs actual room-name paths.
func extractPaths(fg *flowGraph, farm *graph.Farm) [][]string {
	// Build adjacency from flow: room_out → room_in edges that carry flow.
	// An edge u→v carries flow if original cap was >0 and current cap < original.
	// We stored original capacities implicitly: for tunnel edges (cap was 1),
	// current cap 0 means flow=1.
	// Easier: look at the room_out → room_in cross-room edges.
	// The reverse of fg.g[u][i] is fg.g[v][rev] — if reverse cap > 0, flow went u→v.

	// Build a "used" adjacency for rooms (not split nodes).
	used := make(map[string][]string) // room → list of next rooms

	// Reverse-lookup: node index → room name + suffix
	revIdx := make(map[int]string)
	for k, v := range fg.idx {
		revIdx[v] = k
	}

	for u := 0; u < fg.n; u++ {
		uName := revIdx[u]
		if !strings.HasSuffix(uName, "_out") {
			continue
		}
		uRoom := strings.TrimSuffix(uName, "_out")

		for _, e := range fg.g[u] {
			vName := revIdx[e.to]
			if !strings.HasSuffix(vName, "_in") {
				continue
			}
			vRoom := strings.TrimSuffix(vName, "_in")
			if uRoom == vRoom {
				continue // skip internal split edge
			}
			// Check the reverse edge capacity — if >0, flow went u→v.
			revCap := fg.g[e.to][e.rev].cap
			if revCap > 0 {
				used[uRoom] = append(used[uRoom], vRoom)
			}
		}
	}

	var paths [][]string
	start := farm.Start
	end := farm.End

	// Each entry in used[start] is the beginning of a path.
	nexts := append([]string{}, used[start]...)
	sort.Strings(nexts)

	for _, first := range nexts {
		path := []string{start, first}
		visited := map[string]bool{start: true, first: true}
		for path[len(path)-1] != end {
			cur := path[len(path)-1]
			nbrs := append([]string{}, used[cur]...)
			sort.Strings(nbrs)
			moved := false
			for _, nb := range nbrs {
				if !visited[nb] {
					visited[nb] = true
					path = append(path, nb)
					moved = true
					break
				}
			}
			if !moved {
				break
			}
		}
		if path[len(path)-1] == end {
			paths = append(paths, path)
		}
	}

	sort.Slice(paths, func(i, j int) bool {
		if len(paths[i]) != len(paths[j]) {
			return len(paths[i]) < len(paths[j])
		}
		return strings.Join(paths[i], ",") < strings.Join(paths[j], ",")
	})

	return paths
}

// ─── Best path set selection ──────────────────────────────────────────────

// findBestPaths incrementally adds augmenting paths and picks the count
// that minimises total turns.
func findBestPaths(farm *graph.Farm) ([][]string, error) {
	fg, source, sink := buildFlowGraph(farm)

	// Run max-flow to get all possible paths at once.
	totalFlow := maxFlow(fg, source, sink)
	if totalFlow == 0 {
		return nil, fmt.Errorf("ERROR: invalid data format, no path between start and end")
	}

	allPaths := extractPaths(fg, farm)
	if len(allPaths) == 0 {
		return nil, fmt.Errorf("ERROR: invalid data format, no path between start and end")
	}

	// Try using 1 path, 2 paths, ... all paths. Pick the subset of the
	// shortest k paths (by length) that gives fewest turns.
	// The paths are already sorted shortest-first by extractPaths.
	bestTurns := -1
	bestK := 1

	for k := 1; k <= len(allPaths); k++ {
		t := calcTurns(farm.NumAnts, allPaths[:k])
		if bestTurns < 0 || t < bestTurns {
			bestTurns = t
			bestK = k
		}
	}

	return allPaths[:bestK], nil
}

// ─── Turn count calculator ────────────────────────────────────────────────

// calcTurns greedily assigns ants to minimise the number of turns.
func calcTurns(numAnts int, paths [][]string) int {
	antCount := make([]int, len(paths))
	pathLen := make([]int, len(paths))
	for i, p := range paths {
		pathLen[i] = len(p) - 1 // number of moves to reach end
	}

	for ant := 0; ant < numAnts; ant++ {
		best, bestFinish := 0, 1<<30
		for i := range paths {
			// Finish turn if we send next ant down this path:
			// it will be (antCount[i]+1)-th ant, finishes at (antCount[i]+1) + pathLen[i] - 1
			finish := antCount[i] + pathLen[i]
			if finish < bestFinish {
				bestFinish = finish
				best = i
			}
		}
		antCount[best]++
	}

	worst := 0
	for i := range paths {
		f := antCount[i] + pathLen[i] - 1
		if f > worst {
			worst = f
		}
	}
	return worst
}

// ─── Simulation ───────────────────────────────────────────────────────────

// simulate produces the turn-by-turn movement strings.
func simulate(numAnts int, paths [][]string, start, end string) []string {
	// Assign ants to paths using same greedy logic.
	antPath := make([]int, numAnts+1) // 1-indexed ant → path index
	antCount := make([]int, len(paths))
	pathLen := make([]int, len(paths))
	for i, p := range paths {
		pathLen[i] = len(p) - 1
	}

	for ant := 1; ant <= numAnts; ant++ {
		best, bestFinish := 0, 1<<30
		for i := range paths {
			finish := antCount[i] + pathLen[i]
			if finish < bestFinish {
				bestFinish = finish
				best = i
			}
		}
		antPath[ant] = best
		antCount[best]++
	}

	// Build per-path ant queues (ordered: ant 1 is first, etc.)
	queues := make([][]int, len(paths))
	for ant := 1; ant <= numAnts; ant++ {
		p := antPath[ant]
		queues[p] = append(queues[p], ant)
	}

	// antStep[ant] = current step index in its path (-1 = not started)
	antStep := make([]int, numAnts+1)
	for i := range antStep {
		antStep[i] = -1
	}
	antDone := make([]bool, numAnts+1)
	released := make([]int, len(paths)) // how many ants from each queue have been released

	var turns []string
	done := 0

	for done < numAnts {
		var moves []string

		// Build the set of rooms occupied by active ants (not start, not end).
		occupied := make(map[string]bool)
		for ant := 1; ant <= numAnts; ant++ {
			if antStep[ant] < 0 || antDone[ant] {
				continue
			}
			p := antPath[ant]
			room := paths[p][antStep[ant]]
			if room != start && room != end {
				occupied[room] = true
			}
		}

		// For each path, advance ants from furthest ahead to least advanced.
		for pi := range paths {
			path := paths[pi]

			// Collect active ants on this path, sorted furthest-first.
			var onPath []int
			for ant := 1; ant <= numAnts; ant++ {
				if antPath[ant] == pi && antStep[ant] >= 0 && !antDone[ant] {
					onPath = append(onPath, ant)
				}
			}
			sort.Slice(onPath, func(i, j int) bool {
				return antStep[onPath[i]] > antStep[onPath[j]]
			})

			// Try to advance each ant one step.
			for _, ant := range onPath {
				nextStep := antStep[ant] + 1
				if nextStep >= len(path) {
					continue
				}
				nextRoom := path[nextStep]
				// Can move if destination is end or is free.
				if nextRoom == end || !occupied[nextRoom] {
					curRoom := path[antStep[ant]]
					if curRoom != start && curRoom != end {
						delete(occupied, curRoom)
					}
					antStep[ant] = nextStep
					if nextRoom != start && nextRoom != end {
						occupied[nextRoom] = true
					}
					moves = append(moves, fmt.Sprintf("L%d-%s", ant, nextRoom))
					if nextRoom == end {
						antDone[ant] = true
						done++
					}
				}
			}

			// Release next queued ant onto path if room 1 is free.
			if released[pi] < len(queues[pi]) {
				nextAnt := queues[pi][released[pi]]
				if len(path) < 2 {
					continue
				}
				nextRoom := path[1]
				if nextRoom == end || !occupied[nextRoom] {
					occupied[nextRoom] = true
					antStep[nextAnt] = 1
					released[pi]++
					moves = append(moves, fmt.Sprintf("L%d-%s", nextAnt, nextRoom))
					if nextRoom == end {
						antDone[nextAnt] = true
						done++
					}
				}
			}
		}

		if len(moves) > 0 {
			// Sort by ant number.
			sort.Slice(moves, func(i, j int) bool {
				var ai, aj int
				fmt.Sscanf(moves[i][1:], "%d-", &ai)
				fmt.Sscanf(moves[j][1:], "%d-", &aj)
				return ai < aj
			})
			turns = append(turns, strings.Join(moves, " "))
		}
	}

	return turns
}
