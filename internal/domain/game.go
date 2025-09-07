package domain

import (
	"github.com/4vertak/rogue-go/internal/domain/rules"
)

type Game struct {
	r     Renderer
	i     Input
	s     Storage
	state *GameSession
}

func NewGame(s Storage, r Renderer, i Input) *Game { return &Game{r: r, i: i, s: s} }

func (g *Game) Run() {
	if sess, err := g.s.LoadProgress(); err == nil && sess != nil {
		g.state = sess
	} else {
		g.state = NewSession(1)
	}
	g.r.Message("Добро пожаловать в Rogue-Go (TTY)")
	for !g.state.Over {
		g.tick()
	}
	_ = g.s.AppendScore(g.state.ToScore())
}

func (g *Game) tick() {
	g.r.Draw(BuildRenderState(g.state))
	a := g.i.NextAction()

	switch a.Type {
	case MoveUp:
		rules.TryMove(&g.state.Player, 0, -1, &g.state.Level)
	case MoveDown:
		rules.TryMove(&g.state.Player, 0, 1, &g.state.Level)
	case MoveLeft:
		rules.TryMove(&g.state.Player, -1, 0, &g.state.Level)
	case MoveRight:
		rules.TryMove(&g.state.Player, 1, 0, &g.state.Level)
	case UseWeapon, UseFood, UseElixir, UseScroll:
		g.r.Message("Использование предметов пока не реализовано")
	case Quit:
		g.state.Over = true
		g.r.Message("Выход")
		return
	default:
		g.r.Message("Неизвестное действие")
	}
	// TODO: AI, бой, сбор предметов, переход на выход
}
