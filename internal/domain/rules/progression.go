package rules

import (
	"rogue-go/internal/domain/entity"
)

func AdjustDifficulty(state *entity.GameState) {
	if state.PlayerProgress == nil {
		return
	}
	progress := state.PlayerProgress
	player := state.Player
	survivalRate := float64(player.Health) / float64(player.MaxHealth)

	if state.GameSession.CellsWalked-progress.LastAdjustmentTime > 50 {
		if survivalRate < 0.3 {
			if progress.DifficultyLevel > 1 {
				progress.DifficultyLevel--
			}
			state.Messages = append(state.Messages, "The dungeon feels less dangerous...")
		} else if survivalRate > 0.8 {
			if progress.DifficultyLevel < 5 {
				progress.DifficultyLevel++
			}
			state.Messages = append(state.Messages, "The dungeon grows more treacherous...")
		}
		progress.LastAdjustmentTime = state.GameSession.CellsWalked
	}
}
