package rules

import (
	"math"
	"rogue-go/internal/domain/entity"
)

type MiniMapCell struct {
	X, Y  int
	Ch    string
	Color int16
}

type MiniMapData struct {
	Size  int
	Cells []MiniMapCell
}

// СТроим миникарту
func BuildMiniMap(state *entity.GameState, size int, getItemSymbol func(*entity.Item) string, getItemColor func(*entity.Item) int16, getEnemySymbol func(*entity.Enemy) string, getEnemyColor func(*entity.Enemy) int16, getKeyColor func(entity.ItemSubType) int16) MiniMapData {
	var miniMap MiniMapData
	miniMap.Size = size
	if state == nil || state.Player == nil || state.Level == nil {
		return miniMap
	}

	dir := state.PlayerDir
	planeX := -dir.Y
	planeY := dir.X

	for y := 1; y <= size; y++ {
		for x := 1; x <= size; x++ {
			dx := float64(x - 1 - size/2)
			dy := float64(y - 1 - size/2)

			rotX := dx*planeX + (-dy)*dir.X
			rotY := dx*planeY + (-dy)*dir.Y

			mapX := state.Player.Position.X + int(math.Round(rotX))
			mapY := state.Player.Position.Y + int(math.Round(rotY))

			cell := MiniMapCell{X: x, Y: y, Ch: " ", Color: 0}

			if x-1 == size/2 && y-1 == size/2 {
				cell.Ch = "@"
				cell.Color = entity.ColorPlayer
			} else if mapX >= 0 && mapX < len(state.Level.Grid) &&
				mapY >= 0 && mapY < len(state.Level.Grid[0]) {
				if state.Level.Visible[mapX][mapY] {
					gridCell := state.Level.Grid[mapX][mapY]
					if gridCell.IsWall {
						cell.Ch = "#"
						cell.Color = entity.ColorWallNear
					} else if gridCell.IsCorridor {
						cell.Ch = "."
						cell.Color = entity.ColorFloor
					}

					// предметы
					for _, item := range state.Items {
						if item.Position.X == mapX && item.Position.Y == mapY {
							cell.Ch = getItemSymbol(item)
							cell.Color = getItemColor(item)
							break
						}
					}

					// враги
					for _, enemy := range state.Enemies {
						if enemy.Position.X == mapX && enemy.Position.Y == mapY && enemy.IsVisible {
							cell.Ch = getEnemySymbol(enemy)
							cell.Color = getEnemyColor(enemy)
							break
						}
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
						if door.Position.X == mapX && door.Position.Y == mapY {
							cell.Ch = "D"
							cell.Color = getKeyColor(door.Color)
							break
						}
					}
				}
			}
			miniMap.Cells = append(miniMap.Cells, cell)
		}
	}

	return miniMap
}
