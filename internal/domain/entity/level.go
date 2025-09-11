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
	IsGone     bool
	Doors      map[int]Pos // двери верх, низ , левая правая с 0 по 3 индексы будут
}

type Level struct {
	Index    int
	W, H     int
	Tiles    [][]Tile
	Explored [][]bool
	Rooms    []Room
	Exit     Pos
	Mobs     []Monster
	Items    []Item
}
