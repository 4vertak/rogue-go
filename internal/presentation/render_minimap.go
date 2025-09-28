package presentation

import "math"

func (gi *GameInterface) renderMinimap() {
	state := gi.game.State
	_, maxX := gi.gameWindow.MaxYX()
	mapSize := 8

	startX := maxX - mapSize - 2
	startY := 1

	gi.gameWindow.ColorOn(ColorDefault)
	// Верхняя граница
	gi.gameWindow.MovePrint(startY, startX, "+")
	for x := 1; x < mapSize+1; x++ {
		gi.gameWindow.MovePrint(startY, startX+x, "-")
	}
	gi.gameWindow.MovePrint(startY, startX+mapSize+1, "+")

	// eujk направления игрока
	dir := gi.game.State.PlayerDir
	planeX := -dir.Y
	planeY := dir.X

	for y := 1; y <= mapSize; y++ {
		gi.gameWindow.MovePrint(startY+y, startX, "|")

		// Заполняем содержимое мини-карты
		for x := 1; x <= mapSize; x++ {
			dx := float64(x - 1 - mapSize/2)
			dy := float64(y - 1 - mapSize/2)

			// поворот на за ировком
			rotX := dx*planeX + (-dy)*dir.X
			rotY := dx*planeY + (-dy)*dir.Y

			mapX := state.Player.Position.X + int(math.Round(rotX))
			mapY := state.Player.Position.Y + int(math.Round(rotY))

			var ch string
			var color int16

			if x-1 == mapSize/2 && y-1 == mapSize/2 {
				ch = "^"
				color = ColorPlayer
			} else if mapX >= 0 && mapX < len(state.Level.Grid) &&
				mapY >= 0 && mapY < len(state.Level.Grid[0]) {
				if gi.game.State.Level.Visible[mapX][mapY] {
					cell := state.Level.Grid[mapX][mapY]
					if cell.IsWall {
						ch = "#"
						color = ColorWallNear
					} else if cell.IsCorridor {
						ch = "."
						color = ColorFloor
					} else {
						ch = " "
						color = ColorCeiling // черныйй
					}

					// предметы
					for _, item := range state.Items {
						if item.Position.X == mapX && item.Position.Y == mapY {
							ch = gi.game.GetItemSymbol(item)
							color = gi.game.GetItemColor(item)
							break
						}
					}

					// враги
					for _, enemy := range state.Enemies {
						if enemy.Position.X == mapX && enemy.Position.Y == mapY && enemy.IsVisible {
							ch = gi.game.GetEnemySymbol(enemy)
							color = gi.game.GetEnemyColor(enemy)
							break
						}
					}

					// двери
					for _, door := range state.Level.Doors {
						if door.Position.X == mapX && door.Position.Y == mapY {
							ch = "D"
							color = gi.game.GetKeyColor(door.Color)
							break
						}
					}
				}
			} else {
				ch = " "
				color = ColorCeiling
			}

			if color > 0 {
				gi.gameWindow.ColorOn(color)
			}
			gi.gameWindow.MovePrint(startY+y, startX+x, ch)
			if color > 0 {
				gi.gameWindow.ColorOff(color)
			}
		}

		gi.gameWindow.MovePrint(startY+y, startX+mapSize+1, "|")
	}

	// нижняя граница
	gi.gameWindow.MovePrint(startY+mapSize+1, startX, "+")
	for x := 1; x < mapSize+1; x++ {
		gi.gameWindow.MovePrint(startY+mapSize+1, startX+x, "-")
	}
	gi.gameWindow.MovePrint(startY+mapSize+1, startX+mapSize+1, "+")

	gi.gameWindow.ColorOff(ColorDefault)
}
