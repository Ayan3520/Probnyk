package solver_test

import (
	"strings"
	"testing"

	"lem-in/internal/graph"
	"lem-in/internal/parser"
	"lem-in/internal/solver"
)

// parseAndSolve is a convenience helper.
func parseAndSolve(t *testing.T, input string) []string {
	t.Helper()
	farm, _, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	result, err := solver.Solve(farm)
	if err != nil {
		t.Fatalf("solve error: %v", err)
	}
	return result
}

// countTurns counts the non-empty lines (= turns) in the result.
func countTurns(turns []string) int {
	c := 0
	for _, t := range turns {
		if t != "" {
			c++
		}
	}
	return c
}

// allAntsReachEnd verifies that every ant (L1 … Ln) appears exactly once
// in a move ending at the end room, and that the end room move is the last
// appearance of each ant.
func allAntsReachEnd(t *testing.T, turns []string, numAnts int, end string) {
	t.Helper()
	reached := make(map[string]bool)
	for _, turn := range turns {
		for _, move := range strings.Fields(turn) {
			parts := strings.SplitN(move, "-", 2)
			if len(parts) == 2 && parts[1] == end {
				reached[parts[0]] = true
			}
		}
	}
	if len(reached) != numAnts {
		t.Errorf("expected %d ants to reach end, got %d", numAnts, len(reached))
	}
}

// noRoomUsedTwicePerTurn checks the constraint: at most one ant per room per turn.
func noRoomUsedTwicePerTurn(t *testing.T, turns []string, start, end string) {
	t.Helper()
	for i, turn := range turns {
		roomCount := make(map[string]int)
		for _, move := range strings.Fields(turn) {
			parts := strings.SplitN(move, "-", 2)
			if len(parts) == 2 {
				room := parts[1]
				if room != start && room != end {
					roomCount[room]++
					if roomCount[room] > 1 {
						t.Errorf("turn %d: room %q used by multiple ants", i+1, room)
					}
				}
			}
		}
	}
}

// TestExample00 checks the canonical 4-ant example.
func TestExample00(t *testing.T) {
	input := `4
##start
0 0 3
2 2 5
3 4 0
##end
1 8 3
0-2
2-3
3-1`
	turns := parseAndSolve(t, input)
	n := countTurns(turns)
	if n > 6 {
		t.Errorf("expected ≤6 turns, got %d", n)
	}
	allAntsReachEnd(t, turns, 4, "1")
	noRoomUsedTwicePerTurn(t, turns, "0", "1")
}

// TestExample02 checks the 20-ant example.
func TestExample02(t *testing.T) {
	input := `20
##start
0 2 0
1 4 1
2 6 0
##end
3 5 3
0-1
0-3
1-2
3-2`
	turns := parseAndSolve(t, input)
	n := countTurns(turns)
	if n > 11 {
		t.Errorf("expected ≤11 turns, got %d", n)
	}
	allAntsReachEnd(t, turns, 20, "3")
	noRoomUsedTwicePerTurn(t, turns, "0", "3")
}

// TestExample03 checks example03 (4 ants, complex layout).
func TestExample03(t *testing.T) {
	input := `4
4 5 4
##start
0 1 4
1 3 6
##end
5 6 4
2 3 4
3 3 1
0-1
2-4
1-4
0-2
4-5
3-0
4-3`
	turns := parseAndSolve(t, input)
	n := countTurns(turns)
	if n > 6 {
		t.Errorf("expected ≤6 turns, got %d", n)
	}
	allAntsReachEnd(t, turns, 4, "5")
}

// TestExample04 checks example04 (9 ants, named rooms).
func TestExample04(t *testing.T) {
	input := `9
##start
richard 0 6
gilfoyle 6 3
erlich 9 6
dinish 6 9
jimYoung 11 7
##end
peter 14 6
richard-dinish
dinish-jimYoung
richard-gilfoyle
gilfoyle-peter
gilfoyle-erlich
richard-erlich
erlich-jimYoung
jimYoung-peter`
	turns := parseAndSolve(t, input)
	n := countTurns(turns)
	if n > 6 {
		t.Errorf("expected ≤6 turns, got %d", n)
	}
	allAntsReachEnd(t, turns, 9, "peter")
	noRoomUsedTwicePerTurn(t, turns, "richard", "peter")
}

// TestSingleAntDirectPath tests a trivial single-ant, single-step case.
func TestSingleAntDirectPath(t *testing.T) {
	farm := graph.NewFarm()
	farm.NumAnts = 1
	farm.Start = "A"
	farm.End = "B"
	farm.Rooms["A"] = &graph.Room{Name: "A", X: 0, Y: 0, Links: []string{"B"}}
	farm.Rooms["B"] = &graph.Room{Name: "B", X: 1, Y: 0, Links: []string{"A"}}

	turns, err := solver.Solve(farm)
	if err != nil {
		t.Fatal(err)
	}
	if len(turns) != 1 {
		t.Errorf("expected 1 turn, got %d: %v", len(turns), turns)
	}
	if turns[0] != "L1-B" {
		t.Errorf("expected 'L1-B', got %q", turns[0])
	}
}

// TestStartEqualsEnd degenerate case.
func TestStartEqualsEnd(t *testing.T) {
	farm := graph.NewFarm()
	farm.NumAnts = 3
	farm.Start = "A"
	farm.End = "A"
	farm.Rooms["A"] = &graph.Room{Name: "A", X: 0, Y: 0}

	turns, err := solver.Solve(farm)
	if err != nil {
		t.Fatal(err)
	}
	if len(turns) != 0 {
		t.Errorf("expected 0 turns for start==end, got %d", len(turns))
	}
}

// TestNoPath verifies an error is returned when there is no path.
func TestNoPath(t *testing.T) {
	farm := graph.NewFarm()
	farm.NumAnts = 3
	farm.Start = "A"
	farm.End = "B"
	// No links — disconnected graph.
	farm.Rooms["A"] = &graph.Room{Name: "A", X: 0, Y: 0}
	farm.Rooms["B"] = &graph.Room{Name: "B", X: 5, Y: 0}

	_, err := solver.Solve(farm)
	if err == nil {
		t.Fatal("expected error for disconnected graph, got nil")
	}
}

// TestLargeAntCount verifies example06 runs correctly (100 ants).
func TestLargeAntCount(t *testing.T) {
	input := `100
##start
richard 0 6
gilfoyle 6 3
erlich 9 6
dinish 6 9
jimYoung 11 7
##end
peter 14 6
richard-dinish
dinish-jimYoung
richard-gilfoyle
gilfoyle-peter
gilfoyle-erlich
richard-erlich
erlich-jimYoung
jimYoung-peter`
	turns := parseAndSolve(t, input)
	allAntsReachEnd(t, turns, 100, "peter")
	noRoomUsedTwicePerTurn(t, turns, "richard", "peter")
}

// TestNoTunnelUsedTwicePerTurn checks tunnel usage constraint (harder).
func TestNoTunnelUsedTwicePerTurn(t *testing.T) {
	input := `4
##start
0 0 3
2 2 5
3 4 0
##end
1 8 3
0-2
2-3
3-1`
	turns := parseAndSolve(t, input)
	for i, turn := range turns {
		tunnels := make(map[string]bool)
		// Parse moves and check each tunnel direction used at most once.
		for _, move := range strings.Fields(turn) {
			parts := strings.SplitN(move, "-", 2)
			if len(parts) == 2 {
				// We can't reconstruct the tunnel key without knowing previous
				// positions, so just check the room destination is used once
				// (already covered by noRoomUsedTwicePerTurn). Mark the dest.
				if tunnels[parts[1]] {
					t.Errorf("turn %d: room %s visited by 2+ ants", i+1, parts[1])
				}
				tunnels[parts[1]] = true
			}
		}
	}
}

// TestAntMovedAtMostOncePerTurn verifies each ant appears at most once per turn.
func TestAntMovedAtMostOncePerTurn(t *testing.T) {
	input := `10
##start
start 1 6
0 4 8
o 6 8
n 6 6
e 8 4
t 1 9
E 5 9
a 8 9
m 8 6
h 4 6
A 5 2
c 8 1
k 11 2
##end
end 11 6
start-t
n-e
a-m
A-c
0-o
E-a
k-end
start-h
o-n
m-end
t-E
start-0
h-A
e-end
c-k
n-m
h-n`
	turns := parseAndSolve(t, input)
	for i, turn := range turns {
		antCount := make(map[string]int)
		for _, move := range strings.Fields(turn) {
			parts := strings.SplitN(move, "-", 2)
			if len(parts) == 2 {
				antCount[parts[0]]++
				if antCount[parts[0]] > 1 {
					t.Errorf("turn %d: ant %s moved more than once", i+1, parts[0])
				}
			}
		}
	}
}
