package domain

import "github.com/4vertak/rogue-go/internal/domain/entity"

type RenderState struct {
	Level  *entity.Level
	Player *entity.Player
	Log    []string
}

func BuildRenderState(s *GameSession) *RenderState {
	return &RenderState{
		Level:  &s.Level,
		Player: &s.Player,
		Log:    s.LogTail(5),
	}
}
