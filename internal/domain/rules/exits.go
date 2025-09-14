package rules

import (
	"math"
	"rogue-go/internal/domain/entity"
)

// CheckExitProximity — сообщение, если игрок рядом с выходом
func CheckExitProximity(state *entity.GameState) {
	playerPos := state.Player.Position
	exitRoom := state.Level.ExitRoom
	if exitRoom == nil {
		return
	}
	exitX := exitRoom.X + exitRoom.Width/2
	exitY := exitRoom.Y + exitRoom.Height/2

	if state.Level.Visible[exitX][exitY] {
		distance := int(math.Abs(float64(playerPos.X-exitX)) + math.Abs(float64(playerPos.Y-exitY)))
		if distance <= 3 {
			state.Messages = append(state.Messages, "You see the exit '>' nearby!")
		}
	}
}

// CheckAutomaticExit — переход на следующий уровень, если игрок стоит на выходе
func CheckAutomaticExit(state *entity.GameState, nextLevel func()) {
	if state.Level.ExitRoom == nil {
		return
	}
	player := state.Player.Position
	exit := state.Level.ExitRoom
	exitX := exit.X + exit.Width/2
	exitY := exit.Y + exit.Height/2
	if player.X == exitX && player.Y == exitY {
		state.Messages = append(state.Messages, "You go down the stairs...")
		nextLevel()
	}
}
