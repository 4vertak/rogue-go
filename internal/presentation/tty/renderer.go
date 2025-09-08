package tty

import (
	"fmt"
	"strings"

	"github.com/4vertak/rogue-go/internal/domain"
	"github.com/4vertak/rogue-go/internal/domain/entity"
)

type Renderer struct {
	log []string
}

func NewRenderer() *Renderer    { return &Renderer{} }
func (r *Renderer) Init() error { return nil }
func (r *Renderer) Shutdown()   {}

func (r *Renderer) Draw(rs *domain.RenderState) {
	fmt.Print("\033[H\033[2J") // clear
	lvl := rs.Level
	//отрисовка уровня и монстров
	for y := 0; y < lvl.H; y++ {
		var b strings.Builder
		for x := 0; x < lvl.W; x++ {
			ch := tileChar(lvl.Tiles[y][x])

			// Проверка монстров
			for _, mob := range lvl.Mobs {
				if mob.Pos.X == x && mob.Pos.Y == y {
					ch = mob.Symbol
					break
				}
			}

			if rs.Player.Pos.X == x && rs.Player.Pos.Y == y {
				ch = '@'
			}
			b.WriteRune(ch)
		}
		fmt.Println(b.String())
	}
	fmt.Printf("HP %d/%d  STR %d  DEX %d  GOLD %d  LVL %d\n",
		rs.Player.Stats.HP, rs.Player.Stats.MaxHP, rs.Player.Stats.STR, rs.Player.Stats.DEX, rs.Player.Gold, rs.Level.Index)
	for _, m := range rs.Log {
		fmt.Println(m)
	}

}

func (r *Renderer) PromptChoice(title string, n int) int {
	fmt.Println(title, "(0..", n, "):")
	return 0
}

func (r *Renderer) Message(text string) { r.log = append(r.log, text) }

func tileChar(t entity.Tile) rune {
	switch t {
	case entity.Wall:
		return '#'
	case entity.Floor:
		return '.'
	case entity.Exit:
		return '>'
	default:
		return ' '
	}
}
