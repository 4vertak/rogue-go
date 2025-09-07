package domain

type ActionType int

const (
	MoveUp ActionType = iota
	MoveDown
	MoveLeft
	MoveRight
	UseWeapon
	UseFood
	UseElixir
	UseScroll
	Quit
)

type Action struct {
	Type   ActionType
	Choice int
}
