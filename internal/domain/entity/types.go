package entity

type Pos struct{ X, Y int }

type Stats struct {
	MaxHP int
	HP    int
	DEX   int
	STR   int
}

type Weapon struct {
	Name       string
	DamageMin  int
	DamageMax  int
	ToHitBonus int
}

type Character struct {
	Stats  Stats
	Weapon *Weapon
}

type Player struct {
	Character
	Pos  Pos
	Gold int
}

func DefaultPlayer() Player {
	return Player{Character: Character{Stats: Stats{MaxHP: 20, HP: 20, DEX: 5, STR: 5}}, Pos: Pos{X: 2, Y: 2}}
}

type Tile int

const (
	Unknown Tile = iota
	Wall
	Floor
	Exit
)

type Room struct {
	X, Y, W, H int
}

type Level struct {
	Index int
	W, H  int
	Tiles [][]Tile
	Rooms []Room
	Exit  Pos
	Mobs  []Monster
}

type Monster struct {
	Pos       Pos
	Stats     Stats
	Type      string
	Hostility int
	Symbol    rune
}
