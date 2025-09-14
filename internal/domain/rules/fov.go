package rules

import (
	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/utils"
)

func UpdateVisibility(state *entity.GameState) {
	if state.Level == nil || state.Player == nil {
		return
	}
	playerPos := state.Player.Position

	// Сбрасываем видимость
	for i := range state.Level.Visible {
		for j := range state.Level.Visible[i] {
			state.Level.Visible[i][j] = false
		}
	}
	// Устанавливаем видимость в радиусе 6 клеток
	radius := 6
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			x := playerPos.X + dx
			y := playerPos.Y + dy
			if x >= 0 && x < len(state.Level.Grid) && y >= 0 && y < len(state.Level.Grid[0]) {
				// Проверяем лучами видимость
				if hasLineOfSight(state, playerPos, entity.Position{X: x, Y: y}) {
					state.Level.Visible[x][y] = true
					state.Level.Discovered[x][y] = state.Level.Grid[x][y]
				}
			}
		}
	}

	// Обновляем видимость врагов
	for _, enemy := range state.Enemies {
		if enemy.InvisibilityCooldown == 0 {
			enemy.IsVisible = state.Level.Visible[enemy.Position.X][enemy.Position.Y]
		}
	}
}

func hasLineOfSight(state *entity.GameState, start, end entity.Position) bool {
	dx := utils.Abs(end.X - start.X)
	dy := utils.Abs(end.Y - start.Y)

	var sx, sy int
	if start.X < end.X {
		sx = 1
	} else {
		sx = -1
	}
	if start.Y < end.Y {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy
	x, y := start.X, start.Y

	for {
		if x == end.X && y == end.Y {
			break
		}
		// добваиляем проверку стены
		if x < 0 || x >= len(state.Level.Grid) || y < 0 || y >= len(state.Level.Grid[0]) {
			return false
		}
		if state.Level.Grid[x][y].IsWall {
			return false
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
	return true
}

func UpdateEffects(state *entity.GameState) {
	// Уменьшаем длительность эффектов
	if state.Player == nil {
		return
	}
	for i := len(state.Player.ActiveEffects) - 1; i >= 0; i-- {
		e := state.Player.ActiveEffects[i]
		e.Duration--
		if e.Duration <= 0 {
			state.Player.ActiveEffects = append(state.Player.ActiveEffects[:i], state.Player.ActiveEffects[i+1:]...)
		}
	}
}
