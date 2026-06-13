package parser

import (
	"fmt"
	"strconv"
	"strings"

	"lem-in/internal/graph"
)

// Parse reads the full file content and returns a populated Farm and the
// printable lines (original content minus comment-only lines, per spec).
func Parse(content string) (*graph.Farm, []string, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	farm := graph.NewFarm()

	// State machine flags
	var (
		antsParsed  bool
		nextIsStart bool
		nextIsEnd   bool
		startFound  bool
		endFound    bool
	)

	// printable collects lines the spec says to echo (everything except
	// pure single-# comments, but keeping ##start/##end markers).
	printable := []string{}

	// Track seen room names and coordinates to detect duplicates
	seenRooms := map[string]bool{}
	seenCoords := map[string]bool{}

	for i, rawLine := range lines {
		line := strings.TrimRight(rawLine, " \t")

		// Skip empty lines (at end of file especially)
		if line == "" {
			if i == len(lines)-1 {
				continue
			}
			printable = append(printable, line)
			continue
		}

		// ── First non-empty line: number of ants ──────────────────────────
		if !antsParsed {
			n, err := strconv.Atoi(line)
			if err != nil || n <= 0 {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, invalid number of ants")
			}
			farm.NumAnts = n
			antsParsed = true
			printable = append(printable, line)
			continue
		}

		// ── Command lines ─────────────────────────────────────────────────
		if line == "##start" {
			if startFound {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, duplicate ##start")
			}
			nextIsStart = true
			printable = append(printable, line)
			continue
		}
		if line == "##end" {
			if endFound {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, duplicate ##end")
			}
			nextIsEnd = true
			printable = append(printable, line)
			continue
		}

		// ── Comments (#, not ##start/##end) ──────────────────────────────
		if strings.HasPrefix(line, "#") {
			// Single-# comments are kept but not treated as room/link data.
			printable = append(printable, line)
			continue
		}

		// ── Link line "name1-name2" ───────────────────────────────────────
		if isLink(line) {
			parts := strings.SplitN(line, "-", 2)
			a, b := parts[0], parts[1]
			if _, ok := farm.Rooms[a]; !ok {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, unknown room in link: %s", a)
			}
			if _, ok := farm.Rooms[b]; !ok {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, unknown room in link: %s", b)
			}
			if a == b {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, room linked to itself: %s", a)
			}
			// Check for duplicate link
			if isDuplicateLink(farm.Rooms[a].Links, b) {
				return nil, nil, fmt.Errorf("ERROR: invalid data format, duplicate link: %s-%s", a, b)
			}
			farm.Rooms[a].Links = append(farm.Rooms[a].Links, b)
			farm.Rooms[b].Links = append(farm.Rooms[b].Links, a)
			printable = append(printable, line)
			continue
		}

		// ── Room line "name x y" ──────────────────────────────────────────
		room, err := parseRoom(line)
		if err != nil {
			return nil, nil, fmt.Errorf("ERROR: invalid data format, %v", err)
		}

		// Room name must not start with 'L' or '#'
		if strings.HasPrefix(room.Name, "L") || strings.HasPrefix(room.Name, "#") {
			return nil, nil, fmt.Errorf("ERROR: invalid data format, invalid room name: %s", room.Name)
		}

		if seenRooms[room.Name] {
			return nil, nil, fmt.Errorf("ERROR: invalid data format, duplicate room: %s", room.Name)
		}
		coordKey := fmt.Sprintf("%d,%d", room.X, room.Y)
		if seenCoords[coordKey] {
			return nil, nil, fmt.Errorf("ERROR: invalid data format, duplicate coordinates: %d %d", room.X, room.Y)
		}

		seenRooms[room.Name] = true
		seenCoords[coordKey] = true
		farm.Rooms[room.Name] = room

		if nextIsStart {
			farm.Start = room.Name
			startFound = true
			nextIsStart = false
		}
		if nextIsEnd {
			farm.End = room.Name
			endFound = true
			nextIsEnd = false
		}
		printable = append(printable, line)
	}

	// ── Validation ────────────────────────────────────────────────────────
	if !startFound {
		return nil, nil, fmt.Errorf("ERROR: invalid data format, no start room found")
	}
	if !endFound {
		return nil, nil, fmt.Errorf("ERROR: invalid data format, no end room found")
	}
	if len(farm.Rooms) == 0 {
		return nil, nil, fmt.Errorf("ERROR: invalid data format, no rooms found")
	}

	return farm, printable, nil
}

// isLink returns true if the line looks like "a-b" (a link, not a room).
// Room lines have the format "name int int"; link lines have exactly one '-'
// and no spaces (after the split there are no spaces in either part is NOT
// a reliable check because room names can contain digits).
// We use the heuristic: if the line has a '-' and no space, and it doesn't
// parse as a room, treat it as a link.
func isLink(line string) bool {
	if !strings.Contains(line, "-") {
		return false
	}
	// If it has spaces it might be a room definition that contains a hyphen
	// in the name — but the spec says rooms are "name x y" with spaces.
	if strings.Contains(line, " ") {
		return false
	}
	parts := strings.SplitN(line, "-", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return true
}

func parseRoom(line string) (*graph.Room, error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid room definition: %q", line)
	}
	x, err1 := strconv.Atoi(parts[1])
	y, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("invalid room coordinates: %q", line)
	}
	return &graph.Room{Name: parts[0], X: x, Y: y}, nil
}

func isDuplicateLink(links []string, target string) bool {
	for _, l := range links {
		if l == target {
			return true
		}
	}
	return false
}
