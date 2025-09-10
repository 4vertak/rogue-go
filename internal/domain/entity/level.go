package entity

type Tile int

const (
	Unknown Tile = iota
	Wall
	Floor
	Exit
)

type Room struct {
	X, Y, W, H int
	IsGone bool
}

type Level struct {
	Index int
	W, H  int
	Tiles [][]Tile
	Rooms []Room
	Exit  Pos
	Mobs  []Monster
	Items []Item
}
