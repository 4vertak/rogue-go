package tty

import (
	"github.com/4vertak/rogue-go/internal/domain"
	"github.com/rthornton128/goncurses"
)

type Input struct {
	stdscr *goncurses.Window
}

func NewInput(stdscr *goncurses.Window) *Input {
	return &Input{stdscr: stdscr}
}

func (i *Input) NextAction() domain.Action {
	key := i.stdscr.GetChar()

	switch key {
	case 'w', 'W':
		return domain.Action{Type: domain.MoveUp}
	case 's', 'S':
		return domain.Action{Type: domain.MoveDown}
	case 'a', 'A':
		return domain.Action{Type: domain.MoveLeft}
	case 'd', 'D':
		return domain.Action{Type: domain.MoveRight}
	case 'h':
		return domain.Action{Type: domain.UseWeapon}
	case 'j':
		return domain.Action{Type: domain.UseFood}
	case 'k':
		return domain.Action{Type: domain.UseElixir}
	case 'e':
		return domain.Action{Type: domain.UseScroll}
	case goncurses.KEY_ESC:
		return domain.Action{Type: domain.Quit}
	default:
		return domain.Action{Type: -1}
	}
}
