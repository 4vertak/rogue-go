package domain

import (
	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/rules"
)

type RenderState struct {
	Level   *entity.Level
	Player  *entity.Player
	Log     []string
	Visible map[entity.Pos]bool
}

func BuildRenderState(s *GameSession) *RenderState {
	// Вычисляем видимые
	visible := rules.VisibleTiles(&s.Level, s.Player.Pos, 5, func(p entity.Pos) bool {
		// Проверяем, находится ли позиция в пределах уровня
		if p.X < 0 || p.X >= s.Level.W || p.Y < 0 || p.Y >= s.Level.H {
			return true
		}
		return s.Level.Tiles[p.Y][p.X] == entity.Wall
	})

	// Обновляем исследованные
	for pos := range visible {
		if pos.Y >= 0 && pos.Y < s.Level.H && pos.X >= 0 && pos.X < s.Level.W {
			s.Level.Explored[pos.Y][pos.X] = true
		}
	}

	return &RenderState{
		Level:   &s.Level,
		Player:  &s.Player,
		Log:     s.LogTail(5),
		Visible: visible,
	}
}
