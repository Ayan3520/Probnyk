package main

import (
	"fmt"
	"os"

	"lem-in/internal/parser"
	"lem-in/internal/solver"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: invalid data format, cannot read file: %v\n", err)
		os.Exit(1)
	}

	farm, rawLines, err := parser.Parse(string(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Print the original file content
	for _, line := range rawLines {
		fmt.Println(line)
	}
	fmt.Println()

	// Find and simulate paths
	result, err := solver.Solve(farm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Print each turn's moves
	for _, turn := range result {
		fmt.Println(turn)
	}
}
