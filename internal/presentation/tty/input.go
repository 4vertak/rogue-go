package tty

import (
	"bufio"
	"os"

	"github.com/4vertak/rogue-go/internal/domain"
)

type Input struct {
	in *bufio.Reader
}

func NewInput() *Input { return &Input{in: bufio.NewReader(os.Stdin)} }

func (i *Input) NextAction() domain.Action {
	b, _, err := i.in.ReadRune()
	if err != nil {
		return domain.Action{Type: domain.Quit}
	}

	switch b {
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
	case 27: // ESC key
		return domain.Action{Type: domain.Quit}
	default:
		return domain.Action{Type: -1} // Unknown action
	}
}
