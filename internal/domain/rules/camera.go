package rules

import "rogue-go/internal/domain/entity"

func UpdateCamera(state *entity.GameState) {
	playerPos := state.Player.Position
	cam := state.Camera

	cam.X = playerPos.X - cam.Width/2
	cam.Y = playerPos.Y - cam.Height/2

	if cam.X < 0 {
		cam.X = 0
	}
	if cam.Y < 0 {
		cam.Y = 0
	}
	if cam.X+cam.Width >= len(state.Level.Grid) {
		cam.X = len(state.Level.Grid) - cam.Width
	}
	if cam.Y+cam.Height >= len(state.Level.Grid[0]) {
		cam.Y = len(state.Level.Grid[0]) - cam.Height
	}
}
