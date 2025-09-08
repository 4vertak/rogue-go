package entity

type Item struct {
	Type      string // Food, Elixir, Scroll, Weapon, Treasure
	Subtype   string
	Health    int
	MaxHP     int
	Dexterity int
	Strength  int
	Value     int
	Weapon    *Weapon
	Pos       Pos
}
