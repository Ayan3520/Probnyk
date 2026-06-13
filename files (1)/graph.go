package graph

// Farm holds all data about the ant colony.
type Farm struct {
	NumAnts int
	Rooms   map[string]*Room
	Start   string
	End     string
}

// Room represents a single node in the colony graph.
type Room struct {
	Name  string
	X, Y  int
	Links []string
}

// NewFarm creates an initialised Farm.
func NewFarm() *Farm {
	return &Farm{
		Rooms: make(map[string]*Room),
	}
}
