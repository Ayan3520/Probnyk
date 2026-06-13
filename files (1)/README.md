# lem-in

A digital ant-farm simulator written in Go.

## Overview

`lem-in` reads a colony description from a file and finds the **quickest path** to move `N` ants from `##start` to `##end`. It uses **Dinic's max-flow algorithm** on a node-split graph to discover vertex-disjoint paths and then simulates optimal ant distribution across those paths.

## Build & Run

```bash
go run . <filename>
```

or build the binary first:

```bash
go build -o lem-in .
./lem-in <filename>
```

## Usage Examples

```bash
go run . examples/example00.txt
go run . examples/example04.txt
go run . examples/example05.txt
go run . examples/badexample00.txt   # prints ERROR
```

## Input Format

```
number_of_ants
##start
start_room_name x y
[other rooms...]
##end
end_room_name x y
[links in "room1-room2" form]
```

- Lines beginning with `#` (but not `##start`/`##end`) are comments and are ignored.
- Room names must not begin with `L` or `#`.
- Each room can hold at most **one ant per turn** (except `##start` and `##end`).
- Each tunnel can be used by at most **one ant per turn**.

## Output Format

The program prints the original file contents, then one line per turn:

```
Lx-room Ly-room ...
```

where `x`/`y` are ant numbers (1-based) and `room` is the destination.

## Error Handling

Prints `ERROR: invalid data format[, reason]` and exits with status 1 for:
- Zero or negative ant count
- Missing `##start` or `##end`
- Duplicate rooms or links
- Self-links (`room-room`)
- Links to undefined rooms
- No path between start and end

## Algorithm

1. **Parse** the input into a `Farm` graph.
2. **Node-split** each room into `room_in` → `room_out` with capacity 1 (∞ for start/end).
3. **Dinic's algorithm** finds maximum flow = maximum number of vertex-disjoint paths.
4. **Extract paths** from the flow assignment.
5. **Greedy ant assignment**: pick the path that minimises the turn the last ant finishes.
6. **Simulate** movement turn-by-turn, advancing ants from furthest-ahead first.

## Tests

```bash
go test ./...
```

All standard packages only — no external dependencies.

## Project Structure

```
lem-in/
├── main.go                     # Entry point
├── go.mod
├── examples/                   # All provided example files
│   ├── example00.txt … example07.txt
│   └── badexample00.txt, badexample01.txt
└── internal/
    ├── graph/
    │   └── graph.go            # Farm and Room data structures
    ├── parser/
    │   ├── parser.go           # Input parser
    │   └── parser_test.go      # Parser unit tests
    └── solver/
        ├── solver.go           # Max-flow + simulation
        └── solver_test.go      # Solver unit tests
```
