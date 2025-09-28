package rules

import "rogue-go/internal/domain/entity"

type MapCell struct {
	ScreenX, ScreenY int
	Ch               string
	Color            int16
}

type MapRenderData struct {
	Width, Height int
	Cells         []MapCell
}

// BuildMapData готовит DTO для рендера карты
func BuildMapData(state *entity.GameState,
	getItemSymbol func(*entity.Item) string,
	getItemColor func(*entity.Item) int16,
	getEnemySymbol func(*entity.Enemy) string,
	getEnemyColor func(*entity.Enemy) int16,
	getKeySymbol func(entity.ItemSubType) string,
	getKeyColor func(entity.ItemSubType) int16,
	getDoorSymbol func(*entity.Door) string,
	getDoorColor func(*entity.Door) int16,
) MapRenderData {
	data := MapRenderData{}
	if state == nil || state.Level == nil || state.Player == nil || state.Camera == nil {
		return data
	}

	player := state.Player
	camera := state.Camera

	mapWidth := len(state.Level.Grid)
	if mapWidth == 0 {
		return data
	}
	mapHeight := len(state.Level.Grid[0])

	// размеры окна (камера)
	usableWidth := camera.Width
	usableHeight := camera.Height

	camera.X = player.Position.X - usableWidth/2
	camera.Y = player.Position.Y - usableHeight/2

	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}
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

	data.Width = usableWidth
	data.Height = usableHeight

	for x := camera.X; x < camera.X+usableWidth && x < mapWidth; x++ {
		for y := camera.Y; y < camera.Y+usableHeight && y < mapHeight; y++ {
			screenX := x - camera.X + 1
			screenY := y - camera.Y + 1

			cell := MapCell{ScreenX: screenX, ScreenY: screenY, Ch: " "}

			if state.Level.Visible[x][y] {
				grid := state.Level.Grid[x][y]
				if grid.IsWall {
					cell.Ch = "#"
					cell.Color = entity.ColorWallNear
				} else if grid.IsCorridor {
					cell.Ch = "."
					cell.Color = entity.ColorFloor
				}

				// предметы
				for _, item := range state.Items {
					if item.Position.X == x && item.Position.Y == y {
						if item.Type == entity.Scroll &&
							(item.SubType == entity.KeyRed || item.SubType == entity.KeyBlue ||
								item.SubType == entity.KeyGreen || item.SubType == entity.KeyYellow) {
							cell.Ch = getKeySymbol(item.SubType)
							cell.Color = getKeyColor(item.SubType)
						} else {
							cell.Ch = getItemSymbol(item)
							cell.Color = getItemColor(item)
						}
					}
				}

				// враги
				for _, enemy := range state.Enemies {
					if enemy.Position.X == x && enemy.Position.Y == y && enemy.IsVisible {
						cell.Ch = getEnemySymbol(enemy)
						cell.Color = getEnemyColor(enemy)
						break
					}
				}

				// игрок
				if state.Player.Position.X == x && state.Player.Position.Y == y {
					cell.Ch = "@"
					cell.Color = entity.ColorPlayer
				}

				// выход
				if state.Level.ExitRoom != nil {
					exitX := state.Level.ExitRoom.X + state.Level.ExitRoom.Width/2
					exitY := state.Level.ExitRoom.Y + state.Level.ExitRoom.Height/2
					if x == exitX && y == exitY {
						cell.Ch = ">"
						cell.Color = entity.ColorWallNear
					}
				}

				// двери
				for _, door := range state.Level.Doors {
					if door.Position.X == x && door.Position.Y == y {
						cell.Ch = getDoorSymbol(door)
						cell.Color = getDoorColor(door)
						break
					}
				}

			} else if state.Level.Discovered[x][y].IsWall || state.Level.Discovered[x][y].IsCorridor {
				grid := state.Level.Grid[x][y]
				if grid.IsWall {
					cell.Ch = "#"
					cell.Color = entity.ColorWallNear
				}
			}

			data.Cells = append(data.Cells, cell)
		}
	}

	return data
}
