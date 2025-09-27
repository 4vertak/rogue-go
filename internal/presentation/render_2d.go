package presentation

import "rogue-go/internal/domain/entity"

func (gi *GameInterface) renderMap() {
	gi.gameWindow.Erase()
	gi.gameWindow.Box(0, 0)

	player := gi.game.State.Player
	camera := gi.game.State.Camera

	if camera == nil || gi.game.State.Level == nil {
		return
	}

	maxWinY, maxWinX := gi.gameWindow.MaxYX()
	usableWidth := maxWinX - 2
	usableHeight := maxWinY - 2

	camera.X = player.Position.X - usableWidth/2
	camera.Y = player.Position.Y - usableHeight/2

	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}

	mapWidth := len(gi.game.State.Level.Grid)
	if mapWidth == 0 {
		return
	}
	mapHeight := len(gi.game.State.Level.Grid[0])

	if camera.X+usableWidth > mapWidth {
		camera.X = mapWidth - usableWidth
	}
	if camera.Y+usableHeight > mapHeight {
		camera.Y = mapHeight - usableHeight
	}

	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}
	if camera.X > mapWidth-usableWidth && mapWidth > usableWidth {
		camera.X = mapWidth - usableWidth
	}
	if camera.Y > mapHeight-usableHeight && mapHeight > usableHeight {
		camera.Y = mapHeight - usableHeight
	}

	for x := camera.X; x < camera.X+usableWidth && x < len(gi.game.State.Level.Grid); x++ {
		for y := camera.Y; y < camera.Y+usableHeight && y < len(gi.game.State.Level.Grid[0]); y++ {
			screenX := x - camera.X + 1
			screenY := y - camera.Y + 1

			if screenX <= 0 || screenY <= 0 ||
				screenX >= maxWinX-1 || screenY >= maxWinY-1 {
				continue
			}

			if gi.game.State.Level.Visible[x][y] {
				cell := gi.game.State.Level.Grid[x][y]
				var ch string
				var color int16 = 0

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

				// Проверка предметов
				for _, item := range gi.game.State.Items {
					if item.Position.X == x && item.Position.Y == y {
						// ОСОБАЯ ОБРАБОТКА ДЛЯ КЛЮЧЕЙ
						if item.Type == entity.Scroll && (item.SubType == entity.KeyRed || item.SubType == entity.KeyBlue ||
							item.SubType == entity.KeyGreen || item.SubType == entity.KeyYellow) {
							ch = gi.game.GetKeySymbol(item.SubType)
							color = gi.game.GetKeyColor(item.SubType)
						} else {
							ch = gi.game.GetItemSymbol(item)
							color = gi.game.GetItemColor(item)
						}
					}
				}

				// Проверка врагов
				for _, enemy := range gi.game.State.Enemies {
					if enemy.Position.X == x && enemy.Position.Y == y && enemy.IsVisible {
						ch = gi.game.GetEnemySymbol(enemy)
						color = gi.game.GetEnemyColor(enemy)
						break
					}
				}

				// Проверка игрока
				if gi.game.State.Player.Position.X == x &&
					gi.game.State.Player.Position.Y == y {
					ch = "@"
					color = 6
				}

				// Проверка выхода
				if gi.game.State.Level.ExitRoom != nil {
					exitX := gi.game.State.Level.ExitRoom.X + gi.game.State.Level.ExitRoom.Width/2
					exitY := gi.game.State.Level.ExitRoom.Y + gi.game.State.Level.ExitRoom.Height/2
					if x == exitX && y == exitY && gi.game.State.Level.Visible[x][y] {
						ch = ">"
						color = 3
					}
				}

				// Проверка дверей
				for _, door := range gi.game.State.Level.Doors {
					if door.Position.X == x && door.Position.Y == y {
						if gi.game.State.Level.Visible[x][y] {
							ch = gi.game.GetDoorSymbol(door)
							color = gi.game.GetDoorColor(door)
						}
						break
					}
				}

				if color > 0 {
					gi.gameWindow.ColorOn(color)
				}
				gi.gameWindow.MovePrint(screenY, screenX, ch)
				if color > 0 {
					gi.gameWindow.ColorOff(color)
				}
			} else if gi.game.State.Level.Discovered[x][y].IsWall || gi.game.State.Level.Discovered[x][y].IsCorridor {
				cell := gi.game.State.Level.Grid[x][y]
				var ch string
				var color int16 = 0

				if cell.IsWall {
					ch = "#"
					color = 1
				} else {
					ch = " "
				}
				if color > 0 {
					gi.gameWindow.ColorOn(color)
				}
				gi.gameWindow.MovePrint(screenY, screenX, ch)
				if color > 0 {
					gi.gameWindow.ColorOff(color)
				}
			} else {
				gi.gameWindow.MovePrint(screenY, screenX, " ")
			}
		}
	}

	gi.gameWindow.Refresh()
}
