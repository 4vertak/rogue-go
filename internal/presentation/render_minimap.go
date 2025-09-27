package presentation

func (gi *GameInterface) renderMinimap() {
	state := gi.game.State
	_, maxX := gi.gameWindow.MaxYX()
	mapSize := 8

	// Рисуем мини-карту в правом верхнем углу (а не нижнем)
	startX := maxX - mapSize - 2
	startY := 1 // Верхний угол

	// Рисуем рамку мини-карты
	gi.gameWindow.ColorOn(ColorDefault) // Белый цвет

	// Верхняя граница
	gi.gameWindow.MovePrint(startY, startX, "+")
	for x := 1; x < mapSize+1; x++ {
		gi.gameWindow.MovePrint(startY, startX+x, "-")
	}
	gi.gameWindow.MovePrint(startY, startX+mapSize+1, "+")

	// Боковые границы и содержимое
	for y := 1; y <= mapSize; y++ {
		gi.gameWindow.MovePrint(startY+y, startX, "|")

		// Заполняем содержимое мини-карты
		for x := 1; x <= mapSize; x++ {
			mapX := state.Player.Position.X - mapSize/2 + x - 1
			mapY := state.Player.Position.Y - mapSize/2 + y - 1

			var ch string
			var color int16 = 0

			if mapX >= 0 && mapX < len(state.Level.Grid) &&
				mapY >= 0 && mapY < len(state.Level.Grid[0]) {

				cell := state.Level.Grid[mapX][mapY]

				if cell.IsWall {
					ch = "#"
					color = ColorWallNear
				} else {
					ch = "."
					color = ColorFloor
				}

				// Проверяем дверь
				for _, door := range state.Level.Doors {
					if door.Position.X == mapX && door.Position.Y == mapY {
						ch = "D"
						color = gi.game.GetKeyColor(door.Color)
						break
					}
				}

				// Проверяем предметы
				for _, item := range state.Items {
					if item.Position.X == mapX && item.Position.Y == mapY {
						ch = gi.game.GetItemSymbol(item)
						color = gi.game.GetItemColor(item)
						break
					}
				}

				// Проверяем врагов
				for _, enemy := range state.Enemies {
					if enemy.Position.X == mapX && enemy.Position.Y == mapY && enemy.IsVisible {
						ch = gi.game.GetEnemySymbol(enemy)
						color = gi.game.GetEnemyColor(enemy)
						break
					}
				}

				// Игрок в центре мини-карты
				if x-1 == mapSize/2 && y-1 == mapSize/2 {
					ch = "@"
					color = ColorPlayer
				}
			} else {
				// За пределами карты
				ch = " "
				color = ColorCeiling //черный как потолок
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

	// Нижняя граница
	gi.gameWindow.MovePrint(startY+mapSize+1, startX, "+")
	for x := 1; x < mapSize+1; x++ {
		gi.gameWindow.MovePrint(startY+mapSize+1, startX+x, "-")
	}
	gi.gameWindow.MovePrint(startY+mapSize+1, startX+mapSize+1, "+")

	gi.gameWindow.ColorOff(7)
}
