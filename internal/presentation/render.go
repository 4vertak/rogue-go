package presentation

import (
	"math"
	"rogue-go/internal/domain/entity"
	"strconv"

	"github.com/rthornton128/goncurses"
)

func (gi *GameInterface) createWindows() {
	rows, cols := gi.stdscr.MaxYX()

	// Главное игровое окно (центр экрана)
	gameRows := rows - 6
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

	gi.renderStatus()
	gi.renderMessageBar()
	gi.renderInventory()
	gi.stdscr.Refresh()
}

func (gi *GameInterface) renderMap() {
	gi.gameWindow.Erase()
	gi.gameWindow.Box(0, 0)

	state := gi.game.State
	player := state.Player
	camera := state.Camera

	if camera == nil || state.Level == nil {
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

	mapWidth := len(state.Level.Grid)
	if mapWidth == 0 {
		return
	}
	mapHeight := len(state.Level.Grid[0])

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

	for x := camera.X; x < camera.X+usableWidth && x < len(state.Level.Grid); x++ {
		for y := camera.Y; y < camera.Y+usableHeight && y < len(state.Level.Grid[0]); y++ {
			screenX := x - camera.X + 1
			screenY := y - camera.Y + 1

			if screenX <= 0 || screenY <= 0 ||
				screenX >= maxWinX-1 || screenY >= maxWinY-1 {
				continue
			}

			if state.Level.Visible[x][y] {
				cell := state.Level.Grid[x][y]
				var ch string
				var color int16 = 0

				if cell.IsWall {
					ch = "#"
					color = 1
				} else if cell.IsCorridor {
					ch = "."
					color = 2
				} else {
					ch = " "
				}

				// Проверка предметов
				for _, item := range state.Items {
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
				for _, enemy := range state.Enemies {
					if enemy.Position.X == x && enemy.Position.Y == y && enemy.IsVisible {
						ch = gi.game.GetEnemySymbol(enemy)
						color = gi.game.GetEnemyColor(enemy)
						break
					}
				}

				// Проверка игрока
				if state.Player.Position.X == x &&
					state.Player.Position.Y == y {
					ch = "@"
					color = 6
				}

				// Проверка выхода
				if state.Level.ExitRoom != nil {
					exitX := state.Level.ExitRoom.X + state.Level.ExitRoom.Width/2
					exitY := state.Level.ExitRoom.Y + state.Level.ExitRoom.Height/2
					if x == exitX && y == exitY && state.Level.Visible[x][y] {
						ch = ">"
						color = 3
					}
				}

				// Проверка дверей
				for _, door := range state.Level.Doors {
					if door.Position.X == x && door.Position.Y == y {
						if state.Level.Visible[x][y] {
							ch = gi.game.GetDoorSymbol(door) // Используем символ двери
							color = gi.game.GetDoorColor(door) // Используем цвет двери
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
			} else if state.Level.Discovered[x][y].IsWall || state.Level.Discovered[x][y].IsCorridor {
				cell := state.Level.Grid[x][y]
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

func (gi *GameInterface) render3D() {
	state := gi.game.State
	gi.gameWindow.Erase()

	player := state.Player
	level := state.Level
	maxY, maxX := gi.gameWindow.MaxYX()
	width, height := maxX-2, maxY-2

	// Максимальное расстояние видимости
	maxViewDistance := 15.0

	// Рендерим столбы для каждой колонки экрана
	for x := 0; x < width; x++ {
		// Вычисляем положение луча
		cameraX := 2*float64(x)/float64(width) - 1
		rayDirX := state.PlayerDir.X + state.PlayerPlane.X*cameraX
		rayDirY := state.PlayerDir.Y + state.PlayerPlane.Y*cameraX

		// Позиция игрока на карте
		posX := float64(player.Position.X) + 0.5
		posY := float64(player.Position.Y) + 0.5

		// Длина луча от текущей позиции до следующей x или y-стороны
		deltaDistX := math.Abs(1 / rayDirX)
		deltaDistY := math.Abs(1 / rayDirY)

		// Направление шага и начальное sideDist
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

		// DDA алгоритм
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

			// Проверяем границы перед обращением к массиву
			if mapX < 0 || mapX >= len(level.Grid) ||
				mapY < 0 || mapY >= len(level.Grid[0]) {
				hit = true
				break
			}

			// Проверяем стену
			if level.Grid[mapX][mapY].IsWall {
				hit = true
			}
		}

		if !hit {
			continue
		}

		// Проверяем границы еще раз перед рендерингом
		if mapX < 0 || mapX >= len(level.Grid) ||
			mapY < 0 || mapY >= len(level.Grid[0]) {
			continue
		}

		// Вычисляем расстояние до стены
		var perpWallDist float64
		if side == 0 {
			perpWallDist = (float64(mapX) - posX + float64(1-stepX)/2) / rayDirX
		} else {
			perpWallDist = (float64(mapY) - posY + float64(1-stepY)/2) / rayDirY
		}

		// Ограничиваем максимальное расстояние
		if perpWallDist > maxViewDistance {
			continue
		}

		// Вычисляем высоту стены
		lineHeight := int(float64(height) / perpWallDist)
		if lineHeight > height {
			lineHeight = height
		}

		// Вычисляем начальную и конечную точки отрисовки
		drawStart := -lineHeight/2 + height/2
		if drawStart < 0 {
			drawStart = 0
		}
		drawEnd := lineHeight/2 + height/2
		if drawEnd >= height {
			drawEnd = height - 1
		}

		// Получаем текстуру для стены
		textureSymbol := gi.getWallTexture(mapX, mapY)

		// Применяем градиент based on distance
		wallColor := gi.getDistanceColor(perpWallDist, maxViewDistance)

		// Для боковых стен делаем темнее
		if side == 1 {
			wallColor += 3 // Сдвигаем цвет для более темного оттенка
		}

		// Рисуем текстурированный столб
		for y := drawStart; y <= drawEnd; y++ {
			gi.gameWindow.ColorOn(wallColor)
			gi.gameWindow.MovePrint(y+1, x+1, textureSymbol)
			gi.gameWindow.ColorOff(wallColor)
		}

		// Рисуем пол и потолок с градиентом
		for y := 0; y < height; y++ {
			if y < drawStart {
				// Потолок - простой цвет
				gi.gameWindow.ColorOn(8)
				gi.gameWindow.MovePrint(y+1, x+1, " ")
				gi.gameWindow.ColorOff(8)
			} else if y > drawEnd {
				// Пол - простой цвет с точками
				gi.gameWindow.ColorOn(9)
				gi.gameWindow.MovePrint(y+1, x+1, ".")
				gi.gameWindow.ColorOff(9)
			}
		}
	}

	// Рендерим мини-карту ПОСЛЕ стен, чтобы она была поверх них
	if gi.showMinimap {
		gi.renderMinimap()
	}

	// Рендерим двери поверх всего
	gi.renderDoors3D()

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

// Функция для получения градиентного цвета на основе расстояния
func (gi *GameInterface) getDistanceColor(distance float64, maxDistance float64) int16 {
	if distance < maxDistance/3 {
		return 24 // Яркий цвет для близких стен
	} else if distance < maxDistance*2/3 {
		return 25 // Средний цвет
	} else {
		return 26 // Темный цвет для далеких стен
	}
}

func getWallColor(cell entity.Cell, side int) int16 {
	// Разные цвета для разных сторон
	if side == 1 {
		// Темнее для y-сторон
		switch {
		case cell.IsCorridor:
			return 3 // Темно-желтый
		default:
			return 11 // Темно-синий
		}
	}

	// Обычные цвета для x-сторон
	switch {
	case cell.IsCorridor:
		return 2 // Желтый
	default:
		return 4 // Синий
	}
}

func (gi *GameInterface) renderMinimap() {
	state := gi.game.State
	_, maxX := gi.gameWindow.MaxYX() // Правильно получаем оба значения
	mapSize := 8                     // Уменьшим размер мини-карты для 3D режима

	// Рисуем мини-карту в правом верхнем углу (а не нижнем)
	startX := maxX - mapSize - 2
	startY := 1 // Верхний угол

	// Рисуем рамку мини-карты
	gi.gameWindow.ColorOn(7) // Белый цвет

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
					color = 1 // Красный для стен
				} else {
					ch = "."
					color = 2 // Зеленый для пола
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
					color = 6 // Голубой для игрока
				}
			} else {
				// За пределами карты
				ch = " "
				color = 8
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

// Добавим функцию для рендеринга дверей в 3D
func (gi *GameInterface) renderDoors3D() {
	state := gi.game.State
	maxY, maxX := gi.gameWindow.MaxYX()
	screenWidth := maxX - 2
	screenHeight := maxY - 2

	for _, door := range state.Level.Doors {
		// Проверяем видимость двери
		if !state.Level.Visible[door.Position.X][door.Position.Y] {
			continue
		}

		// Вычисляем позицию двери относительно игрока
		relX := door.Position.X - state.Player.Position.X
		relY := door.Position.Y - state.Player.Position.Y

		// Преобразуем в экранные координаты (упрощенная версия)
		angle := math.Atan2(float64(relY), float64(relX)) -
			math.Atan2(state.PlayerDir.Y, state.PlayerDir.X)

		// Нормализуем угол
		for angle < -math.Pi {
			angle += 2 * math.Pi
		}
		for angle > math.Pi {
			angle -= 2 * math.Pi
		}

		// Если дверь в поле зрения
		if math.Abs(angle) < math.Pi/3 {
			screenX := int((angle + math.Pi/3) * float64(screenWidth) / (2 * math.Pi / 3))
			distance := math.Sqrt(float64(relX*relX + relY*relY))

			// Вычисляем высоту двери
			height := int(float64(screenHeight) / distance)
			if height > screenHeight {
				height = screenHeight
			}

			// Центрируем дверь по вертикали
			startY := (screenHeight - height) / 2
			if startY < 0 {
				startY = 0
			}

			// Рисуем дверь
			doorColor := gi.game.GetKeyColor(door.Color)
			for y := startY; y < startY+height && y < screenHeight; y++ {
				gi.gameWindow.ColorOn(doorColor)
				gi.gameWindow.MovePrint(y+1, screenX+1, "D")
				gi.gameWindow.ColorOff(doorColor)
			}
		}
	}
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
			effectText := string(effect.Type) + " (" + strconv.Itoa(effect.Duration) + ")"
			gi.statusWindow.MovePrint(startRow, 1, effectText)
			startRow++
		}
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
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
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
	goncurses.InitPair(1, goncurses.C_RED, goncurses.C_BLACK)     // Стены
	goncurses.InitPair(2, goncurses.C_GREEN, goncurses.C_BLACK)   // Коридоры, еда
	goncurses.InitPair(3, goncurses.C_YELLOW, goncurses.C_BLACK)  // Выход, свитки
	goncurses.InitPair(4, goncurses.C_BLUE, goncurses.C_BLACK)    // Оружие
	goncurses.InitPair(5, goncurses.C_MAGENTA, goncurses.C_BLACK) // Зелья, змей-маг
	goncurses.InitPair(6, goncurses.C_CYAN, goncurses.C_BLACK)    // Игрок, сокровища
	goncurses.InitPair(7, goncurses.C_WHITE, goncurses.C_BLACK)   // Призрак, по умолчанию

	// Для Mimic используем другую пару или меняем подход
	goncurses.InitPair(13, goncurses.C_BLACK, goncurses.C_WHITE) // Mimic

	// Цвета для 3D режима
	goncurses.InitPair(8, goncurses.C_BLACK, goncurses.C_BLACK)  // Потолок
	goncurses.InitPair(9, goncurses.C_WHITE, goncurses.C_BLACK)  // Пол
	goncurses.InitPair(10, goncurses.C_RED, goncurses.C_BLACK)   // Текстура стен 1
	goncurses.InitPair(11, goncurses.C_BLUE, goncurses.C_BLACK)  // Текстура стен 2
	goncurses.InitPair(12, goncurses.C_GREEN, goncurses.C_BLACK) // Текстура стен 3

	// Изменяем инициализацию цветов для текстур
	goncurses.InitPair(20, goncurses.C_RED, goncurses.C_BLACK)    // Кирпич
	goncurses.InitPair(21, goncurses.C_WHITE, goncurses.C_BLACK)  // Камень (бывший GRAY)
	goncurses.InitPair(22, goncurses.C_YELLOW, goncurses.C_BLACK) // Дерево
	goncurses.InitPair(23, goncurses.C_CYAN, goncurses.C_BLACK)   // Мрамор (бывший WHITE)

	// Градиентные цвета для стен
	goncurses.InitPair(24, goncurses.C_RED, goncurses.C_BLACK)     // Яркий
	goncurses.InitPair(25, goncurses.C_MAGENTA, goncurses.C_BLACK) // Средний
	goncurses.InitPair(26, goncurses.C_BLUE, goncurses.C_BLACK)    // Темный
}
