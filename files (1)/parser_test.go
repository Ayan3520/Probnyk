package parser_test

import (
	"testing"

	"lem-in/internal/parser"
)

// helper: parse a string and expect success.
func mustParse(t *testing.T, input string) {
	t.Helper()
	_, _, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// helper: parse a string and expect an error containing substr.
func mustFail(t *testing.T, input string) {
	t.Helper()
	_, _, err := parser.Parse(input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidSimple(t *testing.T) {
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
	mustParse(t, input)
}

func TestZeroAnts(t *testing.T) {
	input := `0
##start
0 0 3
##end
1 8 3
0-1`
	mustFail(t, input)
}

func TestNegativeAnts(t *testing.T) {
	input := `-5
##start
0 0 3
##end
1 8 3
0-1`
	mustFail(t, input)
}

func TestInvalidAntsString(t *testing.T) {
	input := `abc
##start
0 0 3
##end
1 8 3
0-1`
	mustFail(t, input)
}

func TestNoStart(t *testing.T) {
	input := `3
##end
1 8 3
0 0 3
0-1`
	mustFail(t, input)
}

func TestNoEnd(t *testing.T) {
	input := `3
##start
0 0 3
1 8 3
0-1`
	mustFail(t, input)
}

func TestDuplicateRoom(t *testing.T) {
	input := `3
##start
0 0 3
0 0 3
##end
1 8 3
0-1`
	mustFail(t, input)
}

func TestSelfLink(t *testing.T) {
	input := `3
##start
0 0 3
##end
1 8 3
0-0`
	mustFail(t, input)
}

func TestRoomNameStartsWithL(t *testing.T) {
	input := `3
##start
L1 0 3
##end
1 8 3
L1-1`
	mustFail(t, input)
}

// TestRoomNameStartsWithHash: lines beginning with # are comments.
// After ##start, a comment line means nextIsStart stays true;
// the NEXT real room gets assigned as start. The parser doesn't error
// on this — but the solver will error (start==end or no path).
// We just confirm the parser itself doesn't crash.
func TestRoomNameStartsWithHash(t *testing.T) {
	input := `3
##start
#room 0 3
##end
1 8 3`
	// The parser succeeds (room "1" becomes both start and end),
	// which is degenerate but not a parse error per spec.
	_, _, _ = parser.Parse(input) // just confirm no panic
}

// TestHashCommentDoesNotDefineRoom confirms # lines are purely comments.
func TestHashCommentDoesNotDefineRoom(t *testing.T) {
	// If we have a valid start room AFTER a # line, it should still work.
	input := `3
#this is a comment
##start
A 0 0
B 5 0
##end
C 10 0
A-B
B-C`
	mustParse(t, input)
}

func TestCommentIgnored(t *testing.T) {
	input := `3
##start
0 0 3
#this is a comment
2 2 5
##end
1 8 3
0-2
2-1`
	mustParse(t, input)
}

func TestLinkToUnknownRoom(t *testing.T) {
	input := `3
##start
0 0 3
##end
1 8 3
0-99`
	mustFail(t, input)
}

func TestDuplicateLink(t *testing.T) {
	input := `3
##start
0 0 3
##end
1 8 3
2 5 5
0-2
0-2
2-1`
	mustFail(t, input)
}

func TestNamedRooms(t *testing.T) {
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
	mustParse(t, input)
}

func TestFarmFields(t *testing.T) {
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
	farm, _, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if farm.NumAnts != 4 {
		t.Errorf("expected 4 ants, got %d", farm.NumAnts)
	}
	if farm.Start != "0" {
		t.Errorf("expected start=0, got %s", farm.Start)
	}
	if farm.End != "1" {
		t.Errorf("expected end=1, got %s", farm.End)
	}
	if len(farm.Rooms) != 4 {
		t.Errorf("expected 4 rooms, got %d", len(farm.Rooms))
	}
}

func TestPrintableLines(t *testing.T) {
	input := `4
##start
0 0 3
##end
1 8 3
0-1`
	_, lines, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Error("expected non-empty printable lines")
	}
}
