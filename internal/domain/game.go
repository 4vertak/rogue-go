package domain

import (
	"fmt"

	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/gen"
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
		g.state = NewSession()
	}
	g.r.Message("Добро пожаловать в Rogue-Go (TTY)")
	for !g.state.Over {
		g.tick()
	}
	_ = g.s.AppendScore(g.state.ToScore())
}

func (g *Game) tick() {
	//отрисовка
	g.r.Draw(BuildRenderState(g.state))

	// действия игрока
	act := g.i.NextAction()
	applyPlayerAction(g, act)

	// AI монстров
	runEnemiesAI(g)

	// Бой
	resolveCombats(g)

	// Сбор предметов
	collectingItem(g)

	// Переход на новый level
	if onExit(g) {
		g.r.Message(fmt.Sprintf("Вы спустились на уровень %d", g.state.Level.Index+1))
		g.NewLevel(g.state.Level.Index + 1) // генерируем новый ehjdty
		return
	}

}

func (g *Game) NewLevel(depth int) {
	rng := gen.NowRNG()
	level := gen.BuildLevel(rng, depth, gen.DefaultConfig())

	// Ставим игрока в случайную комнату
	if len(level.Rooms) > 0 {
		for {
			startIdxRoom := rng.Intn(len(level.Rooms))
			if !level.Rooms[startIdxRoom].IsGone {
				start := level.Rooms[rng.Intn(len(level.Rooms))]
				g.state.Player.Pos = entity.Pos{
					X: start.X + start.W/2,
					Y: start.Y + start.H/2,
				}
				break
			}
		}
	}

	g.state.Level = level
}

func applyPlayerAction(g *Game, a Action) {

	switch a.Type {
	case MoveUp:
		rules.MovePlayer(&g.state.Player, 0, -1, &g.state.Level)
	case MoveDown:
		rules.MovePlayer(&g.state.Player, 0, 1, &g.state.Level)
	case MoveLeft:
		rules.MovePlayer(&g.state.Player, -1, 0, &g.state.Level)
	case MoveRight:
		rules.MovePlayer(&g.state.Player, 1, 0, &g.state.Level)
	case UseWeapon, UseFood, UseElixir, UseScroll:
		g.r.Message("Использование предметов пока не реализовано")
	case Quit:
		g.state.Over = true
		g.r.Message("Выход")
		return
	default:
		g.r.Message("Неизвестное действие")
	}
}

func runEnemiesAI(g *Game) {
	for mi := range g.state.Level.Mobs {
		rules.MoveMonster(&g.state.Level.Mobs[mi], g.state.Player, &g.state.Level)
	}
}

func resolveCombats(g *Game) {
	for mi := range g.state.Level.Mobs {
		mob := &g.state.Level.Mobs[mi]
		if mob.Pos == g.state.Player.Pos {
			hit, dmg := rules.Attack(&g.state.Player.Stats, &mob.Stats, g.state.Player.Weapon)
			if hit {
				g.r.Message("Вы ударили " + mob.Type + " на " + fmt.Sprint(dmg))
			}
			if mob.Stats.HP <= 0 {
				g.r.Message(mob.Type + " повержен!")
				// TODO: дроп сокровищ
				g.state.Level.Mobs = append(g.state.Level.Mobs[:mi], g.state.Level.Mobs[mi+1:]...)
				break
			}
		}
	}

	// Бой (монстры vs игрока)
	for _, mob := range g.state.Level.Mobs {
		if rules.Abs(mob.Pos.X-g.state.Player.Pos.X)+rules.Abs(mob.Pos.Y-g.state.Player.Pos.Y) == 1 {
			hit, dmg := rules.Attack(&mob.Stats, &g.state.Player.Stats, nil)
			if hit {
				g.r.Message(mob.Type + " ударил вас на " + fmt.Sprint(dmg))
			}
			if g.state.Player.Stats.HP <= 0 {
				g.r.Message("Вы погибли!")
				g.state.Over = true
				return
			}
		}
	}
}

func collectingItem(g *Game) {
	for i := 0; i < len(g.state.Level.Items); i++ {
		item := g.state.Level.Items[i]
		if item.Pos == g.state.Player.Pos {
			if g.state.Player.PickItem(item) {
				g.r.Message("Вы подобрали " + item.Type)
				g.state.Level.Items = append(g.state.Level.Items[:i], g.state.Level.Items[i+1:]...)
				i--
			}
		}
	}
}

func onExit(g *Game) bool {
	return g.state.Level.Tiles[g.state.Player.Pos.Y][g.state.Player.Pos.X] == entity.Exit
}
