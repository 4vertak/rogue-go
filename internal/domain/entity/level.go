package entity

type Tile int

const (
	Unknown Tile = iota
	Wall
	Floor
	Door
	Exit
)

type Room struct {
	X, Y, W, H int
	IsGone     bool
	Doors      map[int]Pos //для ключей на будущее
}

type Level struct {
	Index          int
	W, H           int
	Tiles          [][]Tile
	Explored       [][]bool
	Rooms          []Room
	Exit           Pos
	PlayerStartPos Pos
	Mobs           []Monster
	Items          []Item
}
