package presentation

import (
	"math"
	"rogue-go/internal/domain/entity"
	"rogue-go/internal/presentation/sprites"
)

func (gi *GameInterface) render3D() {
	gi.gameWindow.Erase()

	player := gi.game.State.Player
	level := gi.game.State.Level
	maxY, maxX := gi.gameWindow.MaxYX()
	width, height := maxX-2, maxY-2

	// Кэшируем часто используемые значения
	posX := float64(player.Position.X) + 0.5
	posY := float64(player.Position.Y) + 0.5
	dirX := gi.game.State.PlayerDir.X
	dirY := gi.game.State.PlayerDir.Y
	planeX := gi.game.State.PlayerPlane.X
	planeY := gi.game.State.PlayerPlane.Y

	// Предварительно вычисляем видимых врагов и предметы
	visibleEnemies := gi.getVisibleEnemies()
	visibleItems := gi.getVisibleItems()

	zBuffer := make([]float64, width)
	for i := range zBuffer {
		zBuffer[i] = 30.0 // maxViewDistance
	}

	// Основной цикл по x-координатам экрана
	for x := 0; x < width; x++ {
		cameraX := 2*float64(x)/float64(width) - 1
		rayDirX := dirX + planeX*cameraX
		rayDirY := dirY + planeY*cameraX

		// БЫСТРЫЙ АЛГОРИТМ DDA
		mapX, mapY := int(posX), int(posY)

		var deltaDistX, deltaDistY float64
		if math.Abs(rayDirX) > 1e-9 {
			deltaDistX = math.Abs(1.0 / rayDirX)
		} else {
			deltaDistX = 1e30
		}
		if math.Abs(rayDirY) > 1e-9 {
			deltaDistY = math.Abs(1.0 / rayDirY)
		} else {
			deltaDistY = 1e30
		}

		var stepX, stepY int
		var sideDistX, sideDistY float64

		if rayDirX < 0 {
			stepX = -1
			sideDistX = (posX - float64(mapX)) * deltaDistX
		} else {
			stepX = 1
			sideDistX = (float64(mapX) + 1.0 - posX) * deltaDistX
		}
		if rayDirY < 0 {
			stepY = -1
			sideDistY = (posY - float64(mapY)) * deltaDistY
		} else {
			stepY = 1
			sideDistY = (float64(mapY) + 1.0 - posY) * deltaDistY
		}

		// DDA алгоритм
		hit := false
		side := 0
		var doorHit *entity.Door

		for !hit {
			if sideDistX < sideDistY {
				sideDistX += deltaDistX
				mapX += stepX
				side = 0
			} else {
				sideDistY += deltaDistY
				mapY += stepY
				side = 1
			}

			// Проверка границ
			if mapX < 0 || mapX >= len(level.Grid) || mapY < 0 || mapY >= len(level.Grid[0]) {
				break
			}

			// Проверка стены или двери
			door := gi.getDoorAt(mapX, mapY)
			isClosedDoor := door != nil && door.IsLocked

			if level.Grid[mapX][mapY].IsWall {
				hit = true
			} else if isClosedDoor {
				hit = true
				doorHit = door
			}
		}

		if !hit {
			continue
		}

		// Вычисление расстояния
		var perpWallDist float64
		if side == 0 {
			perpWallDist = (float64(mapX) - posX + float64(1-stepX)/2) / rayDirX
		} else {
			perpWallDist = (float64(mapY) - posY + float64(1-stepY)/2) / rayDirY
		}

		if perpWallDist <= 0 || perpWallDist > 30.0 {
			continue
		}

		zBuffer[x] = perpWallDist

		// Отрисовка стены
		lineHeight := int(float64(height) / perpWallDist)
		if lineHeight > height {
			lineHeight = height
		}

		drawStart := -lineHeight/2 + height/2
		if drawStart < 0 {
			drawStart = 0
		}
		drawEnd := lineHeight/2 + height/2
		if drawEnd >= height {
			drawEnd = height - 1
		}

		// Отрисовка стены (оптимизированная) - и добавил двери

		if doorHit != nil {
			doorHeight := len(sprites.DoorSprite)
			if doorHeight == 0 {
				doorHeight = 1
			}
			doorWidth := maxDoorSpriteWidth()
			if doorWidth == 0 {
				doorWidth = 1
			}
			var wallX float64
			if side == 0 {
				wallX = posY + perpWallDist*rayDirY
			} else {
				wallX = posX + perpWallDist*rayDirX
			}
			wallX -= math.Floor(wallX)

			texX := int(wallX * float64(doorWidth))
			if texX < 0 {
				texX = 0
			}
			if texX >= doorWidth {
				texX = doorWidth - 1
			}

			wallColor := gi.getDoorColorWithDistance(doorHit, perpWallDist, 30.0, side)
			for y := drawStart; y <= drawEnd; y++ {
				relativeY := y - drawStart
				texY := int(float64(relativeY) * float64(doorHeight) / float64(lineHeight))
				if texY >= doorHeight {
					texY = doorHeight - 1
				}
				symbol := gi.getDoorSymbolSafe(texX, texY)
				if symbol != " " {
					gi.gameWindow.ColorOn(wallColor)
					gi.gameWindow.MovePrint(y+1, x+1, symbol)
					gi.gameWindow.ColorOff(wallColor)
				}
			}

		} else {

			textureSymbol := gi.getWallTexture(mapX, mapY)
			wallColor := gi.getDistanceColor(perpWallDist, 30.0, side)
			if side == 1 {
				wallColor += 3 // Затемнение для вертикальных сторон
			}
			// Быстрая отрисовка столбца стены
			for y := drawStart; y <= drawEnd; y++ {
				gi.gameWindow.ColorOn(wallColor)
				gi.gameWindow.MovePrint(y+1, x+1, textureSymbol)
				gi.gameWindow.ColorOff(wallColor)
			}
		}

		// Отрисовка пола и потолка
		for y := 0; y < height; y++ {
			if y < drawStart {
				gi.gameWindow.ColorOn(8) // Потолок
				gi.gameWindow.MovePrint(y+1, x+1, " ")
				gi.gameWindow.ColorOff(8)
			} else if y > drawEnd {
				rowDist := float64(height) / (2.0*float64(y) - float64(height))
				floorX := posX + rowDist*(rayDirX)
				floorY := posY + rowDist*(rayDirY)
				cellX := int(floorX)
				cellY := int(floorY)

				color := int16(9)
				symbol := "."
				if gi.isExitRoomCenter(cellX, cellY) {
					// texX := int((floorX - float64(cellX))*float64(len(stairSprite[0])))
					// texY := int((floorY - float64(cellY))*float64(len(stairSprite)))
					symbol = "#" //gi.getStairSymbol(texX, texY) говно со прайтом получается
					color = 24
				}

				gi.gameWindow.ColorOn(color) // Пол
				gi.gameWindow.MovePrint(y+1, x+1, symbol)
				gi.gameWindow.ColorOff(color)
			}
		}
	}

	// Отрисовка спрайтов (оптимизированная)
	gi.drawSpritesOptimized(visibleEnemies, visibleItems, zBuffer, width, height)

	if gi.showMinimap {
		gi.renderMinimap()
	}
	gi.render3DHUD()
	gi.gameWindow.Refresh()
}

func (gi *GameInterface) getWallTexture(mapX, mapY int) string {
	// Абсолютно безопасная версия - всегда возвращает валидный символ
	// Используем хэш для генерации текстуры на основе координат

	// Приводим координаты к положительным значениям
	safeX := uint32(mapX)
	safeY := uint32(mapY)

	// Простой хэш для детерминированной текстуры
	hash := (safeX * 73856093) ^ (safeY * 19349663)

	// Выбираем тип текстуры на основе хэша
	textureType := hash % NumTextures

	// Выбираем символ текстуры
	switch textureType {
	case TextureBrick:
		symbols := []string{"#", "▒", "▓", "░"}
		return symbols[hash%uint32(len(symbols))]

	case TextureStone:
		symbols := []string{"•", "∘", "∙", "○"}
		return symbols[hash%uint32(len(symbols))]

	case TextureWood:
		symbols := []string{"=", "≡", "║", "│"}
		return symbols[hash%uint32(len(symbols))]

	case TextureMarble:
		symbols := []string{"□", "■", "▢", "▣"}
		return symbols[hash%uint32(len(symbols))]

	default:
		return "#"
	}
}

func (gi *GameInterface) isExitRoomCenter(x, y int) bool {
	room := gi.game.State.Level.ExitRoom
	if room == nil {
		return false
	}
	centerX := room.X + room.Width/2
	centerY := room.Y + room.Height/2
	return x == centerX && y == centerY
}

func (gi *GameInterface) getDistanceColor(distance float64, maxDistance float64, side int) int16 {
	var base int16
	switch {
	case distance < maxDistance/3:
		base = ColorWallNear
	case distance < maxDistance*2/3:
		base = ColorWallMid
	default:
		base = ColorWallFar
	}

	if side == 1 {
		switch base {
		case ColorWallNear:
			return ColorWallNear
		case ColorWallMid:
			return ColorWallMid
		case ColorWallFar:
			return ColorWallFar
		}
	}
	return base
}

func (gi *GameInterface) getDoorAt(x, y int) *entity.Door {
	for _, door := range gi.game.State.Level.Doors {
		if door.Position.X == x && door.Position.Y == y {
			return door
		}
	}
	return nil
}

func (gi *GameInterface) getDoorColorWithDistance(door *entity.Door, distance float64, maxDistance float64, side int) int16 {
	base := gi.game.GetDoorColor(door)

	switch {
	case distance < maxDistance/3:
	case distance < maxDistance*2/3:
		base += 1
	default:
		base += 2
	}
	if side == 1 {
		if base < ColorDoorYellowFar {
			base += 1
		}
	}

	return base
}
