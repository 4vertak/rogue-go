package presentation

import (
	"fmt"
	"math"
	"rogue-go/internal/domain/entity"
	"strconv"

	"github.com/rthornton128/goncurses"
)

const (
	// стены и пол
	ColorWallNear = 1
	ColorWallMid  = 2
	ColorWallFar  = 3
	ColorFloor    = 4
	ColorCeiling  = 5

	// игрок
	ColorPlayer = 6

	// предметы
	ColorWeapon    = 7
	ColorPotion    = 8
	ColorMimic     = 9
	ColorKeyRed    = 10
	ColorKeyBlue   = 11
	ColorKeyGreen  = 12
	ColorKeyYellow = 13
	ColorDefault   = 14

	// двери (три оттенка на каждый цвет - надо выставить)
	ColorDoorRedNear    = 15
	ColorDoorRedMid     = 16
	ColorDoorRedFar     = 17
	ColorDoorBlueNear   = 18
	ColorDoorBlueMid    = 19
	ColorDoorBlueFar    = 20
	ColorDoorGreenNear  = 21
	ColorDoorGreenMid   = 22
	ColorDoorGreenFar   = 23
	ColorDoorYellowNear = 24
	ColorDoorYellowMid  = 25
	ColorDoorYellowFar  = 26

	// монстры
	ColorZombie    = 27
	ColorVampire   = 28
	ColorGhost     = 29
	ColorOgre      = 30
	ColorSnakeMage = 31
)

func (gi *GameInterface) createWindows() {
	rows, cols := gi.stdscr.MaxYX()

	// Главное игровое окно (центр экрана)
	gameRows := rows - 6 // Оставляем место для HUD в 3D режиме
	gameCols := cols - 31

	if gameRows < 10 {
		gameRows = 10
	}
	if gameCols < 20 {
		gameCols = 20
	}

	gi.gameWindow = gi.stdscr.Sub(gameRows, gameCols, 0, 0)
	gi.gameWindow.Keypad(true)

	// Окно статуса (справа)
	gi.statusWindow = gi.stdscr.Sub(gameRows, 30, 0, gameCols)
	gi.statusWindow.Keypad(true)

	// Окно сообщений (снизу слева)
	gi.messageWindow = gi.stdscr.Sub(3, gameCols, gameRows, 0)
	gi.messageWindow.Keypad(true)

	// Окно инвентаря (снизу справа)
	gi.inventoryWindow = gi.stdscr.Sub(3, 30, gameRows, gameCols)
	gi.inventoryWindow.Keypad(true)

	// Камера
	if gi.game.State.Camera != nil {
		maxY, maxX := gi.gameWindow.MaxYX()
		gi.game.State.Camera.Width = maxX - 2
		gi.game.State.Camera.Height = maxY - 2
	}
}

func (gi *GameInterface) renderInitialScreen() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== ROGUE GAME ===\n")
	gi.stdscr.Print("Press any key to start...")
	gi.stdscr.Refresh()

	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) renderGame() {
	state := gi.game.State
	if state.RenderMode == entity.Render3D {
		gi.render3D()
	} else {
		gi.renderMap()
	}

	// Всегда рендерим остальные окна независимо от режима
	gi.renderStatus()
	gi.renderMessageBar()
	gi.renderInventory()
	gi.stdscr.Refresh()
}

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

const hugeFloat = 1e30 //константа для очень большого числа

func (gi *GameInterface) render3D() {
	gi.gameWindow.Erase()

	player := gi.game.State.Player
	level := gi.game.State.Level
	maxY, maxX := gi.gameWindow.MaxYX()
	width, height := maxX-2, maxY-2
	// Максимальное расстояние видимости
	maxViewDistance := 30.0

	zBuffer := make([]float64, width)
	for i := range zBuffer {
		zBuffer[i] = maxViewDistance
	}

	for x := 0; x < width; x++ {

		cameraX := 2*float64(x)/float64(width) - 1
		rayDirX := gi.game.State.PlayerDir.X + gi.game.State.PlayerPlane.X*cameraX
		rayDirY := gi.game.State.PlayerDir.Y + gi.game.State.PlayerPlane.Y*cameraX

		posX := float64(player.Position.X) + 0.5
		posY := float64(player.Position.Y) + 0.5

		var deltaDistX, deltaDistY float64
		if math.Abs(rayDirX) < 1e-9 {
			deltaDistX = hugeFloat
		} else {
			deltaDistX = math.Abs(1.0 / rayDirX)
		}
		if math.Abs(rayDirY) < 1e-9 {
			deltaDistY = hugeFloat
		} else {
			deltaDistY = math.Abs(1.0 / rayDirY)
		}

		// шаги
		var stepX, stepY int
		var sideDistX, sideDistY float64

		if rayDirX < 0 {
			stepX = -1
			sideDistX = (posX - float64(int(posX))) * deltaDistX
		} else {
			stepX = 1
			sideDistX = (float64(int(posX)+1) - posX) * deltaDistX
		}
		if rayDirY < 0 {
			stepY = -1
			sideDistY = (posY - float64(int(posY))) * deltaDistY
		} else {
			stepY = 1
			sideDistY = (float64(int(posY)+1) - posY) * deltaDistY
		}

		mapX, mapY := int(posX), int(posY)
		hit, side := false, 0

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

			// границы
			if mapX < 0 || mapX >= len(level.Grid) ||
				mapY < 0 || mapY >= len(level.Grid[0]) {
				hit = true
				break
			}

			door := gi.getDoorAt(mapX, mapY)
			isClosedDoor := door != nil && door.IsLocked

			if level.Grid[mapX][mapY].IsWall || (isClosedDoor) {
				hit = true
			}
		}

		if !hit {
			continue
		}

		if mapX < 0 || mapX >= len(level.Grid) ||
			mapY < 0 || mapY >= len(level.Grid[0]) {
			continue
		}

		var perpWallDist float64
		if side == 0 {
			if math.Abs(rayDirX) < 1e-9 {
				continue
			}
			perpWallDist = (float64(mapX) - posX + float64(1-stepX)/2) / rayDirX
		} else {
			if math.Abs(rayDirY) < 1e-9 {
				continue
			}
			perpWallDist = (float64(mapY) - posY + float64(1-stepY)/2) / rayDirY
		}
		if perpWallDist <= 0 {
			continue
		}
		if perpWallDist > maxViewDistance {
			continue
		}
		zBuffer[x] = perpWallDist
		drawWall := true

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

		for _, enemy := range gi.game.State.Enemies {
			if !enemy.IsVisible || !gi.game.State.Level.Visible[enemy.Position.X][enemy.Position.Y] {
				continue
			}

			relX := float64(enemy.Position.X) - posX + 0.5
			relY := float64(enemy.Position.Y) - posY + 0.5

			transformX := relY*gi.game.State.PlayerDir.X - relX*gi.game.State.PlayerDir.Y
			transformY := relY*gi.game.State.PlayerDir.Y + relX*gi.game.State.PlayerDir.X
			if transformY <= 0 {
				continue
			}

			enemyScreenX := int(float64(width) / 2 * (1 + transformX/transformY))
			sprite := gi.GetEnemySprite(enemy)
			spriteW := len(sprite[0])

			spriteScreenWidth := int(float64(height) / transformY)
			spriteStartX := enemyScreenX - spriteScreenWidth/2
			spriteEndX := enemyScreenX + spriteScreenWidth/2

			if x >= spriteStartX && x < spriteEndX {
				texX := (x - spriteStartX) * spriteW / spriteScreenWidth

				if transformY < perpWallDist {
					gi.drawSpriteColumn(
						x, transformY, sprite, texX, height, gi.game.GetEnemyColor(enemy), zBuffer, drawStart, drawEnd)
					drawWall = false
					break
				}
			}
		}

		if drawWall {

			var wallX float64
			if side == 0 {
				wallX = posY + perpWallDist*rayDirY
			} else {
				wallX = posX + perpWallDist*rayDirX
			}
			wallX -= math.Floor(wallX)

			door := gi.getDoorAt(mapX, mapY)
			isClosedDoor := door != nil && door.IsLocked
			var textureSymbol string
			if isClosedDoor {
				textureSymbol = "#"
			} else {
				textureSymbol = gi.getWallTexture(mapX, mapY)
			}

			var wallColor int16
			if isClosedDoor {
				wallColor = gi.getDoorColorWithDistance(door, perpWallDist, maxViewDistance, side)
			} else {
				wallColor = gi.getDistanceColor(perpWallDist, maxViewDistance, side)
				if side == 1 {
					wallColor += 3
				}
			}

			if isClosedDoor {
				doorHeight := len(doorSprite)
				if doorHeight == 0 {
					doorHeight = 1
				}
				doorWidth := maxDoorSpriteWidth()
				if doorWidth == 0 {
					doorWidth = 1
				}
				texX := int(wallX * float64(doorWidth))
				if texX < 0 {
					texX = 0
				}
				if texX >= doorWidth {
					texX = doorWidth - 1
				}

				for y := drawStart; y <= drawEnd; y++ {
					relativeY := y - drawStart
					texY := int(float64(relativeY) * float64(doorHeight) / float64(lineHeight))
					if texY < 0 {
						texY = 0
					}
					if texY >= doorHeight {
						texY = doorHeight - 1
					}

					symbol := gi.getDoorSymbolSafe(texX, texY)
					gi.gameWindow.ColorOn(wallColor)
					gi.gameWindow.MovePrint(y+1, x+1, symbol)
					gi.gameWindow.ColorOff(wallColor)
				}
			} else {
				for y := drawStart; y <= drawEnd; y++ {
					gi.gameWindow.ColorOn(wallColor)
					gi.gameWindow.MovePrint(y+1, x+1, textureSymbol)
					gi.gameWindow.ColorOff(wallColor)
				}
			}

			// потолок / пол
			for y := 0; y < height; y++ {
				if y < drawStart {
					gi.gameWindow.ColorOn(8)
					gi.gameWindow.MovePrint(y+1, x+1, " ")
					gi.gameWindow.ColorOff(8)
				} else if y > drawEnd {
					gi.gameWindow.ColorOn(9)
					gi.gameWindow.MovePrint(y+1, x+1, ".")
					gi.gameWindow.ColorOff(9)
				}
			}
		}

		for _, item := range gi.game.State.Items {
			if !gi.game.State.Level.Visible[item.Position.X][item.Position.Y] {
				continue
			}

			relX := float64(item.Position.X) - posX + 0.5
			relY := float64(item.Position.Y) - posY + 0.5

			transformX := relY*gi.game.State.PlayerDir.X - relX*gi.game.State.PlayerDir.Y
			transformY := relY*gi.game.State.PlayerDir.Y + relX*gi.game.State.PlayerDir.X
			if transformY <= 0 {
				continue
			}

			spriteScreenX := int(float64(width) / 2 * (1 + transformX/transformY))
			sprite := gi.GetItemSprite(item)
			spriteW := len(sprite[0])

			spriteScreenWidth := int(float64(height) / transformY)
			spriteStartX := spriteScreenX - spriteScreenWidth/2
			spriteEndX := spriteScreenX + spriteScreenWidth/2

			if x >= spriteStartX && x < spriteEndX {
				texX := (x - spriteStartX) * spriteW / spriteScreenWidth

				// предмет ближе, чем стена?
				if transformY < zBuffer[x] {
					gi.drawItemColumn(
						x, transformY, sprite, texX, height, gi.game.GetItemColor(item),
						zBuffer, drawStart, drawEnd)
				}
			}
		}
	}

	if gi.showMinimap {
		gi.renderMinimap()
	}
	gi.render3DHUD()
	gi.gameWindow.Refresh()
}

func (gi *GameInterface) drawSpriteColumn(
	screenX int,
	distance float64,
	sprite [][]rune,
	texX int,
	screenHeight int,
	color int16,
	zBuffer []float64,
	drawStartWall, drawEndWall int,
) {
	spriteH := len(sprite)
	lineHeight := int(float64(screenHeight) / distance)

	drawStart := -lineHeight/2 + screenHeight/2
	if drawStart < 0 {
		drawStart = 0
	}
	drawEnd := lineHeight/2 + screenHeight/2
	if drawEnd >= screenHeight {
		drawEnd = screenHeight - 1
	}

	for y := 0; y < screenHeight; y++ {
		var symbol string
		if y >= drawStart && y <= drawEnd {
			texY := (y - drawStart) * spriteH / lineHeight
			symbol = gi.getMonsterSymbolSafe(sprite, texX, texY)
		} else {
			symbol = " "
		}

		if symbol != " " {
			// монстр
			gi.gameWindow.ColorOn(color)
			gi.gameWindow.MovePrint(y+1, screenX+1, symbol)
			gi.gameWindow.ColorOff(color)
		} else {
			if y >= drawStartWall && y <= drawEndWall {
				// стена
				wallDepth := zBuffer[screenX]
				gi.gameWindow.ColorOn(gi.getDistanceColor(wallDepth, 15.0, 0)) // костыли
				gi.gameWindow.MovePrint(y+1, screenX+1, "#")
				gi.gameWindow.ColorOff(gi.getDistanceColor(wallDepth, 15.0, 0))
			} else if y < drawStartWall {
				// потолок
				gi.gameWindow.ColorOn(8)
				gi.gameWindow.MovePrint(y+1, screenX+1, " ")
				gi.gameWindow.ColorOff(8)
			} else if y > drawEndWall {
				// пол
				gi.gameWindow.ColorOn(9)
				gi.gameWindow.MovePrint(y+1, screenX+1, ".")
				gi.gameWindow.ColorOff(9)
			}
		}
	}
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

var doorSprite = [][]rune{
	[]rune("MMMMMMMNdlolclclllllllllllllcckWMMMMMMMM"),
	[]rune("MMMMMMMNdcl:cdkkkxoccoxkkkdc::kWMMMMMMMM"),
	[]rune("MMMMMMMNd:::l0WMMNkllkNMMMKo::kWMMMMMMMM"),
	[]rune("MMMMMMMNdcc:lxO00OdlldkOOOko::kWMMMMMMMM"),
	[]rune("MMMMMMMNdcl:cdxkkxolloxxxxdl::kWMMMMMMMM"),
	[]rune("MMMMMMMNdll:l0WWWNkllkNWWW0l::kWMMMMMMMM"),
	[]rune("MMMMMMMNdcl:lk000OdlldO000kl::kWMMMMMMMM"),
	[]rune("MMMMMMMNdcl::::::cccc::c::::;:kWMMMMMMMM"),
	[]rune("MMMMMMMNdcl::ccccc::::cc:;,,,:kWMMMMMMMM"),
	[]rune("MMMMMMMNd::::cooooc::cllc;;;,:kWMMMMMMMM"),
	[]rune("MMMMMMMNdclc:cllllc:::llllc:::kWMMMMMMMM"),
	[]rune("MMMMMMMNdclc::::::c:::::::::::kWMMMMMMMM"),
	[]rune("MMMMMMMNdcl::clllll:::clllc:;:kWMMMMMMMM"),
	[]rune("MMMMMMMNdcl::looool::cloool:;:kWMMMMMMMM"),
	[]rune("MMMMMMMNdclc:llllllc:clllll:::kWMMMMMMMM"),
	[]rune("MMMMMMMNdclccccccccc:ccccccc::kWMMMMMMMM"),
	[]rune("MMMMMMMNd:cc:clllll:::lllllc::kWMMMMMMMM"),
	[]rune("MMMMMMMNdclcclolollc:clooolc::kWMMMMMMMM"),
	[]rune("MMMMMMMNdcllllllllllllllllllcckWMMMMMMMM"),
	[]rune("MMMMMMMNo;:::::::::::::::::c::kWMMMMMMMM"),
}

func maxDoorSpriteWidth() int {
	maxW := 0
	for _, row := range doorSprite {
		if len(row) > maxW {
			maxW = len(row)
		}
	}
	return maxW
}

func (gi *GameInterface) getDoorSymbolSafe(x, y int) string {
	if y < 0 || y >= len(doorSprite) {
		return " "
	}
	row := doorSprite[y]
	if x < 0 || x >= len(row) {
		return " "
	}
	return string(row[x])
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

func (gi *GameInterface) getDoorAt(x, y int) *entity.Door {
	for _, door := range gi.game.State.Level.Doors {
		if door.Position.X == x && door.Position.Y == y {
			return door
		}
	}
	return nil
}

var zombieSprite = [][]rune{
	[]rune("        WXK0OkkxxxxkkO0KN                "),
	[]rune("  WNX0kdllcccccccc::::::cxX              "),
	[]rune("0dlc:,''''''''.',;;;,,,,',dN             "),
	[]rune("d,...''''''',,'.''''''''''cK             "),
	[]rune("N0o:loollllll:;;;'.......'lX             "),
	[]rune(" WkcllldddddokKXX0o,.....,kW             "),
	[]rune(" W0k00xldddlxNWNXNKc',;,;kW              "),
	[]rune(" N00XKxcllccokXNNKd:;;:,:xKW             "),
	[]rune("  Xxll:cllllc:lolc:;;;,...lX             "),
	[]rune("   XOkdoc,,,,;:::::;,'....,cON           "),
	[]rune("      Xo,.':::;:c:;,''.,oxc,:kN          "),
	[]rune("      Kdcclllc::;'.'',',::;:olkW         "),
	[]rune("       XOkkkkOOOl'.'''.,odc:odxK         "),
	[]rune("              W0l;'',;:;:::dko:oX        "),
	[]rune("              XkxOxc,::cxd;;;;lldKW      "),
	[]rune("             Wk:;:cc:;',c:;:cc:;,c0      "),
	[]rune("              Kl,',ld:',,..''....,kW     "),
	[]rune("              Nkc:xXNkcoo;.......'xW     "),
	[]rune("              Xxll0  Klc:....,;,.;O      "),
	[]rune("            XOxocxNNXkol;..''::,:OW      "),
	[]rune("           Xxllc:xXOoooo:,dkc;;,dW       "),
	[]rune("           0l;;;;okdodddlcxOl;;'oN       "),
	[]rune("           Xxc;;:odooooll::lc;;':OW      "),
	[]rune("            WKOOK0dc:;:c::dKkc;'.:kN     "),
	[]rune("                Wk;,,''cx0W WXo'..,kW    "),
	[]rune("                WO:,,;:xKNWNXXx;,;c0W    "),
	[]rune("              W0dc;;::cokOOxdl;,;;;ldOX  "),
	[]rune("             NklcloooccxOOOkxxdoloolccdK "),
	[]rune("             0c;::c:;,,;::ccccc:ccccc:;o "),
	[]rune("             N0kdoollcccc:::;;;;:ccccox0 "),
}

var vampirSprite = [][]rune{
	[]rune("                   NNNXNN               "),
	[]rune("           M X0xolc:::::clldkK          "),
	[]rune("       M    Xl'..''..........lN         "),
	[]rune("            k,;odxxdoccodxoc':K         "),
	[]rune("         XXXdo0KKKKK0kOKKKKKd:kNN       "),
	[]rune("         0ddokK00KK0kdOKKK0K0oldkN M    "),
	[]rune("         NOolO0kxkXOxxOXKkxOOocdK       "),
	[]rune("          KolOK00KK0kxxkOKKK0oo0N       "),
	[]rune("          NklO0O0KKK0kddk0O0kdK         "),
	[]rune("         0l',xd,collol:ld:;xo:xX        "),
	[]rune("        0;...;dl;,.....':ldl'..;xN      "),
	[]rune("        Xkl,..lko::::::coxc..,loxK      "),
	[]rune("       NOc'...,l:,;;;;;;c:'..'l0        "),
	[]rune("      0c'......d0Okddkxxx,.....,oX      "),
	[]rune("     k,.......,0  KkkK  Nl.......:0    M"),
	[]rune("    k,....... ;K  X0OK   d........;O    "),
	[]rune("   O,.........;K  KOkK   k.........;O   "),
	[]rune("  Xc......... 'xKXNNNNXK0l..........cX  "),
	[]rune("  x.......,;'...',,,,,''...,:,..'....x  "),
	[]rune(" Xc......;oo,..............:dl'....'.:K "),
	[]rune(" O,.,,...'''........'.......'......,',O "),
	[]rune(" 0oxKKk;';lo;......',:l:'...,c:..,o00x0 "),
	[]rune("       XOX  d.:kd:;xKN Xd..lK XxxKN     "),
	[]rune("          Kkc.:X XX     k..ldON         "),
	[]rune("          kc:ckN        KdlloO          "),
	[]rune("         M                    M         "),
	[]rune("                                        "),
	[]rune("                                        "),
}

var ghostSprite = [][]rune{
	[]rune("M      M     MN0OkO00KKKOkkK       NN   "),
	[]rune("            NK0KXN MMMMMM Kk0    X00KK0K"),
	[]rune("           XO0 MMMMMMMMMMNKO0   KOK M Xk"),
	[]rune("          NO0 MMMMMM KxkNXOOX X0OXMMMMNk"),
	[]rune("X00KK0X   0OKxo0 MM0d, '0M XKKKK MMMMMNk"),
	[]rune("OX M X0KK00NNl..lKKl'.'dNMMMMMMMMMMMMMNO"),
	[]rune("O MMMMM NN MMNkoxK NXXN MMMMMMMMMMM XK00"),
	[]rune("k0 MMMMMMMMMMMMMMMMMMMMMMMMMMMMM   KOKN "),
	[]rune("XOXMMMMMMMMMMMMMMMMMMMMMMMM XKK00KKKX   "),
	[]rune("NOOK0K M NNMMMMMMMMMMMMMMMM OdxOKKKN    "),
	[]rune(" XKKK00K0kONMMMMMMMMMMMMMMMM XK0K0k0    "),
	[]rune("      XKXN0KMMMMMMMMMMMMMMMMMMMM 0ON    "),
	[]rune("          KONMMMMMMMMMMMMMMMMM XO0N     "),
	[]rune("          NO0  MMMMMMMMMMMMMMKO0X       "),
	[]rune("           X0O0NMMMMM NN  NK00X         "),
	[]rune("             XOxO0KKK0xxkkkOXN M        "),
}

var ogreSprite = [][]rune{
	[]rune("                    NX00XN              "),
	[]rune("              XKKKOdlldk0KX             "),
	[]rune("           N0d:,;ccccclccd0N            "),
	[]rune("         NKd:'.,cooodoooodxk0XN         "),
	[]rune("       NKkc:c:':oloooooollllcldON       "),
	[]rune("        Nx;;c;':lllcccc::clc;:dKN       "),
	[]rune("        Xc,cc'.cdoolllc:;:c:::dO0N      "),
	[]rune("        O,.',..cxxoloooolcclllloON      "),
	[]rune("      Kx:......:loddddolloodoolxN       "),
	[]rune("   NKkoooc,''..,;;;;::ccllcc:lON        "),
	[]rune("  Kxdoooollll::c:;,,,,,,,,,,';o0        "),
	[]rune(" Kdooolod:;looooollc::;;;:ll;,,oX       "),
	[]rune(" 0ddooldd;,cooooooolllccccloc;';O       "),
	[]rune(" klooooo:,,;:ccclllllloooolllc:,cK      "),
	[]rune(" Xkolooo:',,,;clloooooooooooooo;;O      "),
	[]rune("  N0xoooc:cc::ldoooodooolcloool;;kN     "),
	[]rune("    Xxoooool:;;lodoodddoolllc:,'';k     "),
	[]rune("    0dooolll:'.';:::::::;:;,,'''',x     "),
	[]rune("   N0doc:;,,,'.''';;;;,;;;,,,''',c0     "),
	[]rune("     N0kxdoc'',,'ldd0Oxxd;',',oO0X      "),
	[]rune("          0c'',,;x0KN  NKo,,;dX         "),
	[]rune("        NXKkc,,,oX     Xo;:cldxxkO0XN   "),
	[]rune("     X0xdooolc::oK     0lcodddooolccldON"),
	[]rune("   N0xlcccclodkKN   M   NXKXNNKOOkxxk0X "),
	[]rune("     NK0OkxkOKN  M    M            M    "),
}

var snakeMageSprite = [][]rune{
	[]rune("    X0kxxxkkkxloxxxxxxxxxxoldkkxxxxkOKX "),
	[]rune("  X0xoollllllc;:clooollooc:;:lllllloodkX"),
	[]rune(" XOoc:;;,,,;:cc::cccccccc:::cc;;,,,;::lk"),
	[]rune(" Ko;;;;,''',;::::cc;,,;cc::::;,''',;;;;c"),
	[]rune(" 0c',;:;,;:;,;;,':l,...cl,,;;,,:;,;::;,;"),
	[]rune(" Xo,,;;::;::;,;;;;;....;;,;;;;::;:::;;'c"),
	[]rune("  Ko,;:;;:;;:;;:c:;,''',:cloc::;::;:;,c0"),
	[]rune("   Xx:;:;;:;;::cdoc::;::loxdc;;:;;:;:dK "),
	[]rune("     0dc::;:::;;lxdolccdxxxxl:::::coOX  "),
	[]rune("  X    0xl:::::,;oxxxollooodl:::cd0X    "),
	[]rune("0KXKkxxOKKxc;;;;;coxxxdol::lc;:d0       "),
	[]rune("0kdl:;,,:oxo:;;;,,codxxxdddoloOX        "),
	[]rune("  K000x:'''''',,,,;lxxxxxxxdlcdOK       "),
	[]rune("    0l;'''',,;;;,,,:odddddddc,:ldK      "),
	[]rune("   Xo''''';cccllc;,;:ldddddo:;cclk      "),
	[]rune("    x,''''',;:coko;;coxxxxxdc;:lxX   X  "),
	[]rune(" Xkoc,',,,;;,,,,;c::coddxxxolx0X        "),
	[]rune("0l,''';::cllcllcclc:coddxxdl;:oxK       "),
	[]rune("c''',;cccc::clllllcclodddxdc::;;:xX     "),
	[]rune("''',;cllclddlcclolccoodddxocclll:;dX    "),
	[]rune(",'',:oooldOOkxollccloooooolcllllc,lK    "),
	[]rune("d,',:loooloolllc::cllooooc::lc;:;cOX  X "),
	[]rune("Xkc,,:lllllccl:;::loooodxdollooxOKX     "),
	[]rune("  Kkdlc::;;:;;;clooddxkK   XXX          "),
	[]rune("     X0xl:;;::coddxO0K                  "),
}

var mimicSprite = [][]rune{
	[]rune("    WXkxkOOOkxxk0OOOOOOkxxkOOOkxkXW     "),
	[]rune("    Kc..,,,,,..',,'',,,'..',,;,..cK     "),
	[]rune("   No..',,,'...'''...','..'',,,'..lN    "),
	[]rune("   0, ........................... 'O    "),
	[]rune("  Wd...............................oW   "),
	[]rune("  Wd..'..'''...............'''..'..oW   "),
	[]rune("   Kllocdo;lkl,cxo'':oxc,lkl;odcollK    "),
	[]rune("    KOOxOo.cO: ,kc  .ck, :Oc.o0xkOKW    "),
	[]rune("    W0xxc. ... .'.   .'. ... .cxxOW     "),
	[]rune("     Xl;'                     ',cK      "),
	[]rune("     0:::. .'...       ..... .::;0      "),
	[]rune("     kcdc..,::::;;;;;;;::::,..cd:k      "),
	[]rune("    Wxlkoc:'cl;;ol,,;lo,;lc':clkldW     "),
	[]rune("    Nl:dk0d;x0c;kO;:okk:l0x;o0kd:lX     "),
	[]rune("    K:.,;:;,;:;;;:,,;;:;;:;,;:;,.;0     "),
	[]rune("  WWK:.......................'...;K     "),
	[]rune("  W Wx'.',,'...''.....''...',,,..xW     "),
	[]rune("     K:.',;,'..','''',,'..',;;'.;K      "),
	[]rune("     Wx,,,;;,'',,,''',,,'',;;,,,dN      "),
	[]rune("     WN0000000OOOOOOOO0OO0000000X       "),
}

func (gi *GameInterface) drawItemColumn(
	screenX int,
	distance float64,
	sprite [][]rune,
	texX int,
	screenHeight int,
	color int16,
	zBuffer []float64,
	drawStartWall, drawEndWall int,
) {
	spriteH := len(sprite)
	lineHeight := int(float64(screenHeight) / distance)

	drawStart := -lineHeight/2 + screenHeight/2
	if drawStart < 0 {
		drawStart = 0
	}
	drawEnd := lineHeight/2 + screenHeight/2
	if drawEnd >= screenHeight {
		drawEnd = screenHeight - 1
	}

	for y := drawStart; y <= drawEnd; y++ {
		texY := (y - drawStart) * spriteH / lineHeight
		symbol := gi.getItemSymbolSafe(sprite, texX, texY)
		if symbol != " " {
			gi.gameWindow.ColorOn(color)
			gi.gameWindow.MovePrint(y+1, screenX+1, symbol)
			gi.gameWindow.ColorOff(color)
		}
	}
}

func (gi *GameInterface) GetEnemySprite(enemy *entity.Enemy) [][]rune {
	switch enemy.Type {
	case entity.Zombie:
		return zombieSprite
	case entity.Vampire:
		return vampirSprite
	case entity.Ghost:
		return ghostSprite
	case entity.Ogre:
		return ogreSprite
	case entity.SnakeMage:
		return snakeMageSprite
	case entity.Mimic:
		return mimicSprite
	default:
		return [][]rune{[]rune("E")}
	}
}

func (gi *GameInterface) getMonsterSymbolSafe(sprite [][]rune, x, y int) string {
	if y < 0 || y >= len(sprite) {
		return " "
	}
	row := sprite[y]
	if x < 0 || x >= len(row) {
		return " "
	}
	return string(row[x])
}

var foodSprite = [][]rune{
	[]rune("                    KK                  "),
	[]rune("          K00kkOkdddddxkOOOOK           "),
	[]rune("        0kxdooddddoooooddooddx0         "),
	[]rune("      0xddddddddddddddddddddxxdxO       "),
	[]rune("     0xxxddddxxxdxxxxxddxxddxxxxxO      "),
	[]rune("    KkkkxxxxxdxdddxddddddddxxxkkkO0     "),
	[]rune("    0O0OOkkkkkxxxxxxxxxxxxkkkkOO0O0     "),
	[]rune("    K0KKK0000OOkkOkkkkkkkOOO00KKK0O     "),
	[]rune("   KOkxxOKKKKKKKKKKKKKKKKKKKKK0kxkk0    "),
	[]rune("  Kdcl;,lkOkkkkkkkOOOOO0OkkkOOd:;cclO   "),
	[]rune("  0l:ccccodoc;,,cddddddc;;codolccoc:x   "),
	[]rune("  Oo:;;::::ccccok00O000xollccccc:::cxK  "),
	[]rune("  Odxdccc:::;;;:cxO00kocc::::cccldkkx0  "),
	[]rune("     Kdoddddoolllldxdoccloooodxxok      "),
	[]rune("      0xddxkxdddxxxdoodxxxkxddxdk       "),
	[]rune("        0OkxddddddddoodxxxxxxkOK        "),
	[]rune("            KK0OOOOOOOOO00K             "),
}
var treasureSprite = [][]rune{
	[]rune("NXKKKKKKKK00KKKKKKKK000OOO0XN           "),
	[]rune("XOdllllllllllllllccccc:::cxKN           "),
	[]rune("KOxl:::cccccccccccccccccccxKN           "),
	[]rune("XOkoccccccccccccccccccccccxXN           "),
	[]rune("X0OdccccccccccccccccccccccdK            "),
	[]rune("X00xlcllllllllllcccloololcxX            "),
	[]rune("X00OollllllllllllllooldxlokKXNN         "),
	[]rune("NK0Oocccccccccccccloxx0Ok0OOXNXX        "),
	[]rune(" XXKxlllooooxkkkkkkxddkOKNXKK00N        "),
	[]rune("  NX000000KKXXXXKKKOxxdx0XXKOOKNN       "),
	[]rune("  N0kk0XXNXXXKKKXK0000OO0KX0O0XNNN      "),
	[]rune("  N0xox0KXXXKKKXXKXKKKXXXXKXXNNXXKXNN   "),
	[]rune("   Kxc:loOKXXXKKK0OOOOkkOKXK000OkxOXN   "),
	[]rune("   Xkl::cx000xddollx0OkolkKkllllcoON    "),
	[]rune("   N0dlc:dO0koclllloddxdoooolllcld0N    "),
	[]rune("   NXOkdoxO0Ooclccccccllllllllllox0N    "),
	[]rune("     NXKKK000kdddddooddxkO0OOOOOOk0NNN  "),
	[]rune("         NX00KKK0OOkkO0KXNNNXKKKK0KXKXXN"),
	[]rune("          NKKNNNNN           NXXXXXXXXXN"),
}

var scrollSprite = [][]rune{
	[]rune("                         N0OKKKXNN      "),
	[]rune("                       XxldkKXXNXKKK    "),
	[]rune("                      KocokkOKKKXK0xON  "),
	[]rune("                    N0dddk00KXKOkkdoxX  "),
	[]rune("                  NOxddkOOKXKK0kxl:oO0N "),
	[]rune("                 KxdxkO000KK0OkxocldolO "),
	[]rune("              NKOdloddxOKKOOkxdoodoc:c0 "),
	[]rune("             Kdll:::c::lkkdooodolcldOX  "),
	[]rune("            Kocdl;,,,,,;oxollolldOK     "),
	[]rune("         N0kxkOK0xlcc:;cxdcclox0N       "),
	[]rune("       KkddxO0KKKKKOOkocc:,:xX          "),
	[]rune("    N0xoodk0000K0OOkxdl:,,;xN           "),
	[]rune("  Xxl:lxOKXKK00Okddocccc:o0             "),
	[]rune(" Xo...,::cddx00xolccllco0               "),
	[]rune(" Xl..'',cll;:dkdc:llccdX                "),
	[]rune(" Nx,..''',:c:,;clolclxX                 "),
	[]rune("  Xx,.,::;,,'..'c::xKN                  "),
	[]rune("    Kd;,;;;,'...'cxN                    "),
	[]rune("      Kxl;''....,kN                     "),
	[]rune("        N0xlccccxN                      "),
}

var potionSprite = [][]rune{
	[]rune("                                        "),
	[]rune("               XOkOkkk0N                "),
	[]rune("              Xoccclcccx                "),
	[]rune("              Nd:cc:;;;k                "),
	[]rune("               Ooxxlc;cK                "),
	[]rune("               Ookxlc;lK                "),
	[]rune("               kokxlc;c0                "),
	[]rune("             Kkxkkxolc:cdK              "),
	[]rune("           Xo,cOkc,'''...;kN            "),
	[]rune("          Nd'..,'.........,O            "),
	[]rune("           k:'............;0            "),
	[]rune("           Nkc;,........'c0             "),
	[]rune("             N0OxooooodkKN              "),
	[]rune("                                        "),
}

var weaponSprite = [][]rune{
	[]rune(",odoodxk                                "),
	[]rune(",dk00Okkxx                              "),
	[]rune(":::lxOO0K0xx                            "),
	[]rune("kcclclxKXXKOxx                          "),
	[]rune("  dlcccoOXXKOkxx                        "),
	[]rune("    ocllxO00OO0Oxx                      "),
	[]rune("      dlddoldOKKOxdxk                   "),
	[]rune("        dolcclOK0kkkxxk                 "),
	[]rune("          oc:cdkO00KXKOxx               "),
	[]rune("           OdlodooxKXXKOkxx     klck    "),
	[]rune("              doolcd0KXKOkkdx0Od:,;x    "),
	[]rune("                dlldxkOOOOkxoodl:lO     "),
	[]rune("                  dooddllxxddo:;dX      "),
	[]rune("                    kdlc:ldo:'';:dK     "),
	[]rune("                     Koccc;...';:::o    "),
	[]rune("                   Odl;;:ll,....';:,;okK"),
	[]rune("                  d;,,;lO   o;'....,lddc"),
	[]rune("                  KOOO        o,.':ddl;."),
	[]rune("                               k::c;';c:"),
	[]rune("                               Xx:'..,cx"),
}

var keySprite = [][]rune{
	[]rune("    X0kkkkOKN                           "),
	[]rune("   0xkkkkOkxOX NN                       "),
	[]rune("XOkxxkkkkkOkxxkxk0XKKXXXXXXXXXXXXXXXXXXN"),
	[]rune("dloxdoddddddddddooooooloolooooooooooooox"),
	[]rune("ddkXOdxkxxxxxxxxxxdooooolllloooolllloooo"),
	[]rune("kkO0O0KKK0KK0000OkkxxxxkO0OOkxxxkOkddxkK"),
	[]rune("X000KKKXXKXXOO0OOKXXKXXN    NNNN  NXKX  "),
	[]rune("   KxxkOOkkxOX NN                       "),
	[]rune("    XOkkkkO0N                           "),
	[]rune("         N                              "),
}

func (gi *GameInterface) GetItemSprite(item *entity.Item) [][]rune {
	switch item.Type {
	case entity.Weapon:
		return weaponSprite
	case entity.Food:
		return foodSprite
	case entity.Potion:
		return potionSprite
	case entity.Scroll:
		switch item.SubType {
		case entity.MaxHealthScroll, entity.AgilityScroll, entity.StrengthScroll:
			return scrollSprite
		case entity.KeyRed, entity.KeyBlue, entity.KeyGreen, entity.KeyYellow:
			return keySprite
		default:
			return [][]rune{[]rune("?")}
		}
	case entity.Treasure:
		return treasureSprite
	default:
		return [][]rune{[]rune("?")}
	}
}

func (gi *GameInterface) getItemSymbolSafe(sprite [][]rune, x, y int) string {
	if y < 0 || y >= len(sprite) {
		return " "
	}
	row := sprite[y]
	if x < 0 || x >= len(row) {
		return " "
	}
	return string(row[x])
}

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

func (gi *GameInterface) renderStatus() {
	state := gi.game.State
	gi.statusWindow.Erase()
	gi.statusWindow.Box(0, 0)

	player := state.Player
	gi.statusWindow.MovePrint(1, 1, "Level:"+strconv.Itoa(state.CurrentLevel))
	gi.statusWindow.MovePrint(2, 1, "Health:"+strconv.Itoa(player.Health)+"/"+strconv.Itoa(player.MaxHealth))
	gi.statusWindow.MovePrint(3, 1, "Str:"+strconv.Itoa(player.Strength))
	if player.CurrentWeapon != nil {
		gi.statusWindow.MovePrint(3, 10, "("+strconv.Itoa(player.Strength+player.CurrentWeapon.StrengthBoost)+")")
	}
	gi.statusWindow.MovePrint(4, 1, "Agi:"+strconv.Itoa(player.Agility))
	gi.statusWindow.MovePrint(5, 1, "Gold:"+strconv.Itoa(player.Treasures))

	if player.CurrentWeapon != nil {
		gi.statusWindow.MovePrint(6, 1, "Weapon: Equipped")
	} else {
		gi.statusWindow.MovePrint(6, 1, "Weapon: None")
	}

	// Отображаем активные эффекты
	gi.renderEffects()

	// Отображаем сообщения о предметах в статусе
	gi.renderStatusMessages()

	// Отображаем статус сна
	if state.SleepTurns > 0 {
		gi.statusWindow.MovePrint(15, 1, "Sleep: "+strconv.Itoa(state.SleepTurns)+" turns")
	}

	gi.statusWindow.Refresh()
}

func (gi *GameInterface) renderStatusMessages() {
	state := gi.game.State

	if state.StatusMessages != nil && len(state.StatusMessages) > 0 {
		startRow := 12
		maxY, _ := gi.statusWindow.MaxYX() // Добавляем получение maxY

		gi.statusWindow.MovePrint(startRow, 1, "Recent items:")
		startRow++

		startIdx := max(0, len(state.StatusMessages)-5)
		for i := startIdx; i < len(state.StatusMessages) && startRow < maxY-2; i++ {
			gi.statusWindow.MovePrint(startRow, 1, state.StatusMessages[i])
			startRow++
		}
	}
}

func (gi *GameInterface) renderEffects() {
	state := gi.game.State
	player := state.Player
	if player.ActiveEffects != nil && len(player.ActiveEffects) > 0 {
		maxY, _ := gi.statusWindow.MaxYX()
		startRow := 8

		gi.statusWindow.MovePrint(startRow, 1, "Effects:")
		startRow++

		for _, effect := range player.ActiveEffects {
			if startRow >= maxY-2 {
				break
			}

			effectName := getEffectName(effect.Type)
			effectText := effectName + " (" + strconv.Itoa(effect.Duration) + ")"
			gi.statusWindow.MovePrint(startRow, 1, effectText)
			startRow++
		}
	}
}

// Добавьте вспомогательную функцию для преобразования типа эффекта в название
func getEffectName(effectType entity.EffectType) string {
	switch effectType {
	case entity.HealthEffect:
		return "Health"
	case entity.AgilityEffect:
		return "Agility"
	case entity.StrengthEffect:
		return "Strength"
	case entity.MaxHealthEffect:
		return "Max Health"
	default:
		return "Unknown"
	}
}

func (gi *GameInterface) renderMessageBar() {
	state := gi.game.State
	gi.messageWindow.Erase()
	gi.messageWindow.Box(0, 0)

	if state.Messages != nil && len(state.Messages) > 0 {
		lastMessage := state.Messages[len(state.Messages)-1]
		gi.messageWindow.MovePrint(1, 1, lastMessage)
	}

	gi.messageWindow.Refresh()
}

func (gi *GameInterface) renderInventory() {
	state := gi.game.State
	gi.inventoryWindow.Erase()
	gi.inventoryWindow.Box(0, 0)
	gi.inventoryWindow.MovePrint(1, 1, "Inventory:")

	backpack := state.Backpack
	if backpack != nil {
		weaponCount := len(backpack.Weapons)
		foodCount := len(backpack.Food)
		potionCount := len(backpack.Potions)
		scrollCount := len(backpack.Scrolls)

		gi.inventoryWindow.MovePrint(2, 1, "W:"+strconv.Itoa(weaponCount))
		gi.inventoryWindow.MovePrint(2, 5, "F:"+strconv.Itoa(foodCount))
		gi.inventoryWindow.MovePrint(2, 9, "P:"+strconv.Itoa(potionCount))
		gi.inventoryWindow.MovePrint(2, 13, "S:"+strconv.Itoa(scrollCount))
	}

	gi.inventoryWindow.Refresh()
}

func (gi *GameInterface) showGameOver() {
	state := gi.game.State
	gi.stdscr.Erase()
	gi.stdscr.Print("=== GAME OVER ===\n\n")
	gi.stdscr.Print("You have been defeated!\n\n")

	player := state.Player
	gi.stdscr.Printf("Level reached: %d\n", state.CurrentLevel)
	gi.stdscr.Printf("Treasures collected: %d\n", player.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", player.DefeatedEnemies)

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
}

func (gi *GameInterface) showVictory() {
	state := gi.game.State
	gi.stdscr.Erase()
	gi.stdscr.Print("=== VICTORY! ===\n\n")
	gi.stdscr.Print("You have completed the game!\n\n")

	player := state.Player
	gi.stdscr.Printf("Final level: %d\n", state.CurrentLevel)
	gi.stdscr.Printf("Treasures collected: %d\n", player.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", player.DefeatedEnemies)

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
	gi.stdscr.GetChar()
}

func (gi *GameInterface) initColors() {
	goncurses.StartColor()

	goncurses.InitPair(1, goncurses.C_WHITE, goncurses.C_BLACK) // стены блидние
	goncurses.InitPair(2, goncurses.C_CYAN, goncurses.C_BLACK)  // стены средние
	goncurses.InitPair(3, goncurses.C_BLUE, goncurses.C_BLACK)  // стеныдальние
	goncurses.InitPair(4, goncurses.C_WHITE, goncurses.C_BLACK) // пол
	goncurses.InitPair(5, goncurses.C_BLACK, goncurses.C_BLACK) // потолок

	goncurses.InitPair(6, goncurses.C_BLUE, goncurses.C_BLACK)

	goncurses.InitPair(7, goncurses.C_YELLOW, goncurses.C_BLACK)  // оружие
	goncurses.InitPair(8, goncurses.C_CYAN, goncurses.C_BLACK)    // зелье
	goncurses.InitPair(9, goncurses.C_WHITE, goncurses.C_BLACK)   // мимик
	goncurses.InitPair(10, goncurses.C_RED, goncurses.C_BLACK)    // Ключ и дверь
	goncurses.InitPair(11, goncurses.C_BLUE, goncurses.C_BLACK)   //  Ключ и дверь
	goncurses.InitPair(12, goncurses.C_GREEN, goncurses.C_BLACK)  //  Ключ и дверь
	goncurses.InitPair(13, goncurses.C_YELLOW, goncurses.C_BLACK) //  Ключ и дверь
	goncurses.InitPair(14, goncurses.C_WHITE, goncurses.C_BLACK)  // деволтный белый на всякий

	// двери с градиентом
	// Красная дверь (15–17)
	goncurses.InitPair(15, goncurses.C_RED, goncurses.C_BLACK)
	goncurses.InitPair(16, goncurses.C_RED, goncurses.C_BLACK)
	goncurses.InitPair(17, goncurses.C_RED, goncurses.C_BLACK)

	// Синяя дверь (18–20)
	goncurses.InitPair(18, goncurses.C_BLUE, goncurses.C_BLACK)
	goncurses.InitPair(19, goncurses.C_BLUE, goncurses.C_BLACK)
	goncurses.InitPair(20, goncurses.C_BLUE, goncurses.C_BLACK)
	// Зелёная дверь (21–23)
	goncurses.InitPair(21, goncurses.C_GREEN, goncurses.C_BLACK)
	goncurses.InitPair(22, goncurses.C_GREEN, goncurses.C_BLACK)
	goncurses.InitPair(23, goncurses.C_GREEN, goncurses.C_BLACK)

	// Жёлтая дверь (24–26)
	goncurses.InitPair(24, goncurses.C_YELLOW, goncurses.C_BLACK)
	goncurses.InitPair(25, goncurses.C_YELLOW, goncurses.C_BLACK)
	goncurses.InitPair(26, goncurses.C_YELLOW, goncurses.C_BLACK)

	goncurses.InitPair(27, goncurses.C_GREEN, goncurses.C_BLACK)  // Зомби
	goncurses.InitPair(28, goncurses.C_RED, goncurses.C_BLACK)    // Вампир
	goncurses.InitPair(29, goncurses.C_WHITE, goncurses.C_BLACK)  // Привидение
	goncurses.InitPair(30, goncurses.C_YELLOW, goncurses.C_BLACK) // Огр
	goncurses.InitPair(31, goncurses.C_WHITE, goncurses.C_BLACK)  // Змей
}

// Рендерим HUD поверх 3D сцены
func (gi *GameInterface) render3DHUD() {
	state := gi.game.State
	maxY, maxX := gi.gameWindow.MaxYX()

	// Полупрозрачный фон для HUD
	gi.gameWindow.ColorOn(7)
	for x := 0; x < maxX; x++ {
		gi.gameWindow.MovePrint(maxY-3, x, " ")
	}
	gi.gameWindow.ColorOff(7)

	// Информация о здоровье
	healthPercent := float64(state.Player.Health) / float64(state.Player.MaxHealth)
	healthBar := "["
	barWidth := 20
	filled := int(healthPercent * float64(barWidth))

	for i := 0; i < barWidth; i++ {
		if i < filled {
			healthBar += "="
		} else {
			healthBar += " "
		}
	}
	healthBar += "]"

	gi.gameWindow.MovePrint(maxY-2, 1, "HP: "+healthBar)
	gi.gameWindow.MovePrint(maxY-2, 25,
		fmt.Sprintf("%d/%d", state.Player.Health, state.Player.MaxHealth))

	// Текущее оружие
	if state.Player.CurrentWeapon != nil {
		gi.gameWindow.MovePrint(maxY-1, 1,
			"Weapon: "+gi.game.GetItemName(state.Player.CurrentWeapon))
	} else {
		gi.gameWindow.MovePrint(maxY-1, 1, "Weapon: None")
	}

	// Подсказки управления
	controls := "WASD:Move  Mouse:Look  E:Inventory  Space:Interact  T:2D/3D"
	gi.gameWindow.MovePrint(maxY-1, maxX-len(controls)-1, controls)
}
