package main

import (
	"math/rand"
	"strconv"

	"github.com/rthornton128/goncurses"
)

// GameInterface реализует пользовательский интерфейс
type GameInterface struct {
	stdscr          *goncurses.Window
	gameWindow      *goncurses.Window
	statusWindow    *goncurses.Window
	messageWindow   *goncurses.Window
	inventoryWindow *goncurses.Window
}

func (gi *GameInterface) Initialize() error {
	stdscr, err := goncurses.Init()
	if err != nil {
		return err
	}


	goncurses.Raw(true)
	goncurses.Echo(false)
	goncurses.Cursor(0)
	stdscr.Keypad(true)


	gi.stdscr = stdscr
	gi.createWindows()
	gi.initColors()


	return nil
}

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

	// Устанавливаем размер камеры равным usable area игрового окна

	// Устанавливаем размер камеры
	if CurrentGameState.Camera != nil {
		maxY, maxX := gi.gameWindow.MaxYX()
		CurrentGameState.Camera.Width = maxX - 2
		CurrentGameState.Camera.Height = maxY - 2
	}
}

func (gi *GameInterface) Run() {
	defer goncurses.End()


	gi.renderInitialScreen()

	// Неблокирующий ввод
	gi.stdscr.Timeout(100) // 100ms timeout


	// Неблокирующий ввод - переделал под блокирующий ввод (для пошагового режима)
	gi.stdscr.Timeout(-1)

	for CurrentGameState.IsRunning {
		gi.renderGame()
		gi.handleInput()

		if !CurrentGameState.GameOver && !CurrentGameState.GameWon {
			GameActionsImpl.UpdateGameState()

		if CurrentGameState.IsPlayerTurn {
			gi.handleInput()
		} else { // весь мир ждет хода игорока ))
			if !CurrentGameState.GameOver {
				GameActionsImpl.UpdateGameState()
			}
		}


		if CurrentGameState.GameOver {
			gi.showGameOver()
			break
		}

		if CurrentGameState.GameWon {
			gi.showVictory()
			break
		}

		// Небольшая задержка для снижения нагрузки на CPU
		time.Sleep(50 * time.Millisecond)

		// УБИРАЕМ проверку GameWon здесь, так как переход на следующий уровень
		// теперь обрабатывается через nextLevel()

		// time.Sleep(50 * time.Millisecond) // не нужен
	}
}

func (gi *GameInterface) renderInitialScreen() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== ROGUE GAME ===\n")
	gi.stdscr.Print("Press any key to start...")
	gi.stdscr.Refresh()

	// Ожидаем ввода с блокировкой для начального экрана
	gi.stdscr.Timeout(-1)  // Блокирующий режим
	gi.stdscr.GetChar()    // Ждем любого нажатия
	gi.stdscr.Timeout(100) // Возвращаем неблокирующий режим

	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) renderGame() {
	gi.renderMap()
	gi.renderStatus()
	gi.renderMessageBar()
	gi.renderInventory()
	gi.stdscr.Refresh()
}

func (gi *GameInterface) renderMap() {
	gi.gameWindow.Erase()
	gi.gameWindow.Box(0, 0)

	// Обновляем позицию камеры для следования за игроком

	player := CurrentGameState.Player
	camera := CurrentGameState.Camera


	if camera == nil || CurrentGameState.Level == nil {
		return
	}

	// Получаем размеры игрового окна (без учета интерфейса справа)

	maxWinY, maxWinX := gi.gameWindow.MaxYX()

	// Учитываем рамку окна (2 символа с каждой стороны)
	usableWidth := maxWinX - 2  // минус левая и правая рамка
	usableHeight := maxWinY - 2 // минус верхняя и нижняя рамка

	// Центрируем камеру на игроке, но в пределах игрового окна
	usableWidth := maxWinX - 2
	usableHeight := maxWinY - 2

	camera.X = player.Position.X - usableWidth/2
	camera.Y = player.Position.Y - usableHeight/2

	// Ограничиваем камеру пределами карты

	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}

	// Получаем реальные размеры карты

	mapWidth := len(CurrentGameState.Level.Grid)
	if mapWidth == 0 {
		return
	}
	mapHeight := len(CurrentGameState.Level.Grid[0])

	// Корректно ограничиваем камеру справа и снизу
	if camera.X+usableWidth > mapWidth {

	if camera.X+usableWidth > mapWidth {
		camera.X = mapWidth - usableWidth
	}
	if camera.Y+usableHeight > mapHeight {
	if camera.Y+usableHeight > mapHeight {
		camera.Y = mapHeight - usableHeight
	}

	// Убедимся, что камера не выходит за границы после корректировки
	if camera.X < 0 {
		camera.X = 0
	}
	if camera.Y < 0 {
		camera.Y = 0
	}
	if camera.X > mapWidth-usableWidth && mapWidth > usableWidth {

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
	if camera.Y > mapHeight-usableHeight && mapHeight > usableHeight {
		camera.Y = mapHeight - usableHeight
	}

	// Рендерим только видимую часть карты

	for x := camera.X; x < camera.X+usableWidth && x < len(CurrentGameState.Level.Grid); x++ {
		for y := camera.Y; y < camera.Y+usableHeight && y < len(CurrentGameState.Level.Grid[0]); y++ {
			screenX := x - camera.X + 1 // +1 для рамки
			screenY := y - camera.Y + 1 // +1 для рамки

			// Проверяем границы внутри рамки более тщательно
			if screenX <= 0 || screenY <= 0 ||
				screenX >= maxWinX-1 || screenY >= maxWinY-1 {
				continue // Пропускаем, если выходим за границы окна
			screenX := x - camera.X + 1
			screenY := y - camera.Y + 1

			if screenX <= 0 || screenY <= 0 ||
				screenX >= maxWinX-1 || screenY >= maxWinY-1 {
				continue
			}


			if CurrentGameState.Level.Visible[x][y] {
				cell := CurrentGameState.Level.Grid[x][y]
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
				for _, item := range CurrentGameState.Items {
					if item.Position.X == x && item.Position.Y == y {
						// ОСОБАЯ ОБРАБОТКА ДЛЯ КЛЮЧЕЙ
						if item.Type == Scroll && (item.SubType == KeyRed || item.SubType == KeyBlue ||
							item.SubType == KeyGreen || item.SubType == KeyYellow) {
							ch = gi.getKeySymbol(item.SubType)
							color = gi.getKeyColor(item.SubType)
						} else {
							ch = gi.getItemSymbol(item)
							color = gi.getItemColor(item)
						}
					}
				}


				// Проверка врагов
				for _, enemy := range CurrentGameState.Enemies {
					if enemy.Position.X == x && enemy.Position.Y == y && (enemy.IsVisible || enemy.Type == Mimic) {

						ch = gi.getEnemySymbol(enemy)
						color = gi.getEnemyColor(enemy)
						break
					}
				}


				// Проверка игрока
				if CurrentGameState.Player.Position.X == x &&
					CurrentGameState.Player.Position.Y == y {
				if CurrentGameState.Player.Position.X == x &&
					CurrentGameState.Player.Position.Y == y {
					ch = "@"
					color = 6
				}

				// Проверка выхода
				if CurrentGameState.Level.ExitRoom != nil {
					exitX := CurrentGameState.Level.ExitRoom.X + CurrentGameState.Level.ExitRoom.Width/2
					exitY := CurrentGameState.Level.ExitRoom.Y + CurrentGameState.Level.ExitRoom.Height/2
					if x == exitX && y == exitY && CurrentGameState.Level.Visible[x][y] {
						ch = ">"
						color = 3
					}
				}


				if color > 0 {
					gi.gameWindow.ColorOn(color)
				}
				gi.gameWindow.MovePrint(screenY, screenX, ch)
				if color > 0 {
					gi.gameWindow.ColorOff(color)
				}
			} else if CurrentGameState.Level.Discovered[x][y].IsWall || CurrentGameState.Level.Discovered[x][y].IsCorridor {
				cell := CurrentGameState.Level.Grid[x][y]
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

// Функция для получения символа ключа
func (gi *GameInterface) getKeySymbol(keyType ItemSubType) string {
	switch keyType {
	case KeyRed:
		return "R"
	case KeyBlue:
		return "B"
	case KeyGreen:
		return "G"
	case KeyYellow:
		return "Y"
	default:
		return "K"
	}
}

// Функция для получения цвета ключа
func (gi *GameInterface) getKeyColor(keyType ItemSubType) int16 {
	switch keyType {
	case KeyRed:
		return 1 // Красный
	case KeyBlue:
		return 4 // Синий
	case KeyGreen:
		return 2 // Зеленый
	case KeyYellow:
		return 3 // Желтый
	default:
		return 7 // Белый
	}
}

func (gi *GameInterface) renderStatus() {
	gi.statusWindow.Erase()
	gi.statusWindow.Box(0, 0)

	player := CurrentGameState.Player
	gi.statusWindow.MovePrint(1, 1, "Level:"+strconv.Itoa(CurrentGameState.CurrentLevel))
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

	gi.statusWindow.Refresh()
	gi.statusWindow.Erase()
	gi.statusWindow.Box(0, 0)

	player := CurrentGameState.Player
	gi.statusWindow.MovePrint(1, 1, "Level:"+strconv.Itoa(CurrentGameState.CurrentLevel))
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
	if CurrentGameState.SleepTurns > 0 {
		gi.statusWindow.MovePrint(15, 1, "Sleep: "+strconv.Itoa(CurrentGameState.SleepTurns)+" turns")
	}

	gi.statusWindow.Refresh()
}

func (gi *GameInterface) renderStatusMessages() {
	if CurrentGameState.StatusMessages != nil && len(CurrentGameState.StatusMessages) > 0 {
		// Получаем размеры окна статуса
		maxY, maxX := gi.statusWindow.MaxYX()
		startRow := 12 // Начинаем ниже эффектов

		gi.statusWindow.MovePrint(startRow, 1, "Recent items:")
		startRow++

		// Показываем последние 5 сообщений
		startIdx := max(0, len(CurrentGameState.StatusMessages)-5)
		for i, msg := range CurrentGameState.StatusMessages[startIdx:] {
			if startRow+i < maxY-2 { // Оставляем место для рамки
				// Обрезаем длинные сообщения
				maxLen := maxX - 4
				if len(msg) > maxLen {
					msg = msg[:maxLen-3] + "..."
				}
				gi.statusWindow.MovePrint(startRow+i, 1, msg)
			}
		}
	}
	if CurrentGameState.StatusMessages != nil && len(CurrentGameState.StatusMessages) > 0 {
		startRow := 12
		maxY, _ := gi.statusWindow.MaxYX() // Добавляем получение maxY

		gi.statusWindow.MovePrint(startRow, 1, "Recent items:")
		startRow++

		startIdx := max(0, len(CurrentGameState.StatusMessages)-5)
		for i := startIdx; i < len(CurrentGameState.StatusMessages) && startRow < maxY-2; i++ {
			gi.statusWindow.MovePrint(startRow, 1, CurrentGameState.StatusMessages[i])
			startRow++
		}
	}
}

func (gi *GameInterface) renderEffects() {
	player := CurrentGameState.Player
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
	gi.messageWindow.Erase()
	gi.messageWindow.Box(0, 0)

	// Получаем размеры окна сообщений
	_, maxX := gi.messageWindow.MaxYX()

	// Всегда показываем управление в первой строке
	controlText := "Move: WASD | Items: H-J-K-E | Pickup: P"
	gi.messageWindow.MovePrint(1, 1, controlText)

	// Во второй строке показываем только не-предметные сообщения
	if CurrentGameState.Messages != nil && len(CurrentGameState.Messages) > 0 {
		// Берем последнее сообщение, если оно не о предмете
		lastMsg := CurrentGameState.Messages[len(CurrentGameState.Messages)-1]
		if !isItemMessage(lastMsg) && len(lastMsg) > 0 {
			// Обрезаем если слишком длинное
			maxLen := maxX - 4
			if len(lastMsg) > maxLen {
				lastMsg = lastMsg[:maxLen-3] + "..."
			}
			gi.messageWindow.MovePrint(2, 1, lastMsg)
		}
	}

	gi.messageWindow.Refresh()
	gi.messageWindow.Erase()
	gi.messageWindow.Box(0, 0)

	if CurrentGameState.Messages != nil && len(CurrentGameState.Messages) > 0 {
		lastMessage := CurrentGameState.Messages[len(CurrentGameState.Messages)-1]
		gi.messageWindow.MovePrint(1, 1, lastMessage)
	}

	gi.messageWindow.Refresh()
}

// Вспомогательная функция для определения сообщений о предметах
func isItemMessage(message string) bool {
	itemKeywords := []string{
		"Picked up", "Equipped", "treasure", "food", "potion", "scroll",
		"weapon", "gold", "HP", "STR", "AGI", "MaxHP", "Health", "Agility",
		"Strength", "item", "backpack", "full", "No items", "pick up",
	}
	messageLower := strings.ToLower(message)
	for _, keyword := range itemKeywords {
		if strings.Contains(messageLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (gi *GameInterface) renderEffects() {
	effects := CurrentGameState.Player.ActiveEffects
	if len(effects) > 0 {
		gi.statusWindow.MovePrint(8, 1, "Effects:")
		for i, effect := range effects {
			if i < 3 {
				gi.statusWindow.MovePrint(9+i, 1,
					gi.getEffectSymbol(effect)+" "+strconv.Itoa(effect.Duration))
			}
		}
	}
}

func (gi *GameInterface) renderInventory() {
	gi.inventoryWindow.Erase()
	gi.inventoryWindow.Box(0, 0)

	gi.inventoryWindow.MovePrint(1, 1, "Inventory:")

	backpack := CurrentGameState.Backpack
	gi.inventoryWindow.MovePrint(1, 1, "F:"+strconv.Itoa(len(backpack.Food)))
	gi.inventoryWindow.MovePrint(1, 5, "P:"+strconv.Itoa(len(backpack.Potions)))
	gi.inventoryWindow.MovePrint(1, 9, "S:"+strconv.Itoa(len(backpack.Scrolls)))
	gi.inventoryWindow.MovePrint(1, 13, "W:"+strconv.Itoa(len(backpack.Weapons)))
	gi.inventoryWindow.MovePrint(1, 17, "T:"+strconv.Itoa(backpack.Treasure))

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

func (gi *GameInterface) handleInput() {
	char := gi.stdscr.GetChar()

	// Проверяем на ошибку ввода (таймаут) - в новой версии goncurses
	// GetChar() возвращает -1 при таймауте, но как rune
	if char == goncurses.Key(-1) {
		return // Нет ввода
	//только если ход игрока - избыточно но навярняка)
	if !CurrentGameState.IsPlayerTurn {
		return
	}

	if CurrentGameState.SleepTurns > 0 {
		GameActionsImpl.AddMessage("You are asleep! (" + strconv.Itoa(CurrentGameState.SleepTurns) + " turns remaining)")
		CurrentGameState.SleepTurns--
		return
	}


	switch char {
	// движение WASD + стрелки
	case goncurses.KEY_UP, 'w', 'W':
		GameActionsImpl.MovePlayer(0, -1)
	case goncurses.KEY_DOWN, 's', 'S':
		GameActionsImpl.MovePlayer(0, 1)
	case goncurses.KEY_LEFT, 'a', 'A':
		GameActionsImpl.MovePlayer(-1, 0)
	case goncurses.KEY_RIGHT, 'd', 'D':
		GameActionsImpl.MovePlayer(1, 0)

	// Управление инвентарём / использование предметов
	case 'h', 'H': // Оружие
		gi.showUseItemMenu(Weapon)
	case 'j', 'J': // Еда
		gi.showUseItemMenu(Food)
	case 'k', 'K': // Эликсиры / Potion
		gi.showUseItemMenu(Potion)
	case 'e', 'E': // Свитки / Scroll
		gi.showUseItemMenu(Scroll)

	// дополнительные удобства
	case 'i', 'I':
		gi.showInventory()
	case 'p', 'P':
		gi.showPickupMenu()
	case 'q', 'Q':
		gi.showQuitMenu()
	case 'v', 'V':
		gi.showSaveMenu()
	case 'o', 'O':
		gi.showLoadMenu()
	}
}

func (gi *GameInterface) showWeaponSelection() {
	gi.showSelection("Select weapon (0=Unequip, 1-9=Equip):", CurrentGameState.Backpack.Weapons, Weapon)
}

func (gi *GameInterface) showFoodSelection() {
	gi.showSelection("Select food (1-9):", CurrentGameState.Backpack.Food, Food)
}

func (gi *GameInterface) showPotionSelection() {
	gi.showSelection("Select potion (1-9):", CurrentGameState.Backpack.Potions, Potion)
}

func (gi *GameInterface) showScrollSelection() {
	gi.showSelection("Select scroll (1-9):", CurrentGameState.Backpack.Scrolls, Scroll)
}

func (gi *GameInterface) showSelection(prompt string, items []*Item, itemType ItemType) {
	if len(items) == 0 && (itemType != Weapon || CurrentGameState.Player.CurrentWeapon == nil) {
		gi.showMessage("No items available!")
		return
	}

func (gi *GameInterface) showUseItemMenu(itemType ItemType) {
	gi.stdscr.Erase()

	bp := CurrentGameState.Backpack
	if bp == nil {
		gi.showMessage("Backpack empty")
		return
	}

	// Временно переключаемся на блокирующий ввод
	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100) // Возвращаем неблокирующий ввод в конце

	switch itemType {
	case Weapon:
		gi.stdscr.Print("=== WEAPONS (0 - unequip) ===\n\n")
		gi.stdscr.Print("0. Unequip current weapon\n")
		if len(bp.Weapons) == 0 {
			gi.stdscr.Print("No weapons in backpack\n")
		} else {
			for i, w := range bp.Weapons {
				gi.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, w.StrengthBoost)
			}
		}
		gi.stdscr.Print("\nChoose 0-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch == '0' {
			GameActionsImpl.UseItem(Weapon, 0)
			return
		}
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '0')
			if idx-1 < len(bp.Weapons) {
				GameActionsImpl.UseItem(Weapon, idx)
				return
			}
		}

	case Food:
		if len(bp.Food) == 0 {
			gi.showMessage("No food in backpack")
			return
		}
		gi.stdscr.Print("=== FOOD ===\n\n")
		for i, f := range bp.Food {
			gi.stdscr.Printf("%d. Food (HP+%d)\n", i+1, f.HealthBoost)
		}
		gi.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Food) {
				GameActionsImpl.UseItem(Food, idx)
				return
			}
		}

	case Potion:
		if len(bp.Potions) == 0 {
			gi.showMessage("No potions in backpack")
			return
		}
		gi.stdscr.Print("=== POTIONS ===\n\n")
		for i := range bp.Potions {
			gi.stdscr.Printf("%d. Potion\n", i+1)
		}
		gi.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Potions) {
				GameActionsImpl.UseItem(Potion, idx)
				return
			}
		}

	case Scroll:
		if len(bp.Scrolls) == 0 {
			gi.showMessage("No scrolls in backpack")
			return
		}
		gi.stdscr.Print("=== SCROLLS ===\n\n")
		for i := range bp.Scrolls {
			gi.stdscr.Printf("%d. Scroll\n", i+1)
		}
		gi.stdscr.Print("\nChoose 1-9 or any other key to cancel: ")
		gi.stdscr.Refresh()

		ch := gi.stdscr.GetChar()
		if ch >= '1' && ch <= '9' {
			idx := int(ch - '1')
			if idx < len(bp.Scrolls) {
				GameActionsImpl.UseItem(Scroll, idx)
				return
			}
		}
	}

	// Если нажата любая другая клавиша - просто возвращаемся без сообщения
}

func (gi *GameInterface) showInventory() {
	gi.stdscr.Erase()
	gi.stdscr.Print(prompt + "\n\n")

	// Показываем текущее экипированное оружие
	if itemType == Weapon && CurrentGameState.Player.CurrentWeapon != nil {
		gi.stdscr.Print("Current: " + gi.getItemDescription(CurrentGameState.Player.CurrentWeapon) + "\n\n")
	}

	// Отображаем предметы с правильными номерами (начиная с 1)
	for i, item := range items {
		gi.stdscr.Printf("%d: %s\n", i+1, gi.getItemDescription(item)) // i+1 чтобы номера начинались с 1
	}

	if itemType == Weapon {
		gi.stdscr.Print("0: Unequip current weapon\n")
	}

	gi.stdscr.Print("\nPress number or ESC to cancel")
	gi.stdscr.Print("=== INVENTORY ===\n\n")

	backpack := CurrentGameState.Backpack

	if backpack != nil {
		gi.stdscr.Print("Weapons:\n")
		for i := range backpack.Weapons {
			gi.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, backpack.Weapons[i].StrengthBoost)
		}

		gi.stdscr.Print("\nFood:\n")
		for i := range backpack.Food {
			gi.stdscr.Printf("%d. Food (HP+%d)\n", i+1, backpack.Food[i].HealthBoost)
		}

		gi.stdscr.Print("\nPotions:\n")
		for i := range backpack.Potions {
			gi.stdscr.Printf("%d. Potion\n", i+1)
		}

		gi.stdscr.Print("\nScrolls and Keys:\n")
		for i, scroll := range backpack.Scrolls {
			if scroll.SubType == KeyRed || scroll.SubType == KeyBlue ||
				scroll.SubType == KeyGreen || scroll.SubType == KeyYellow {
				gi.stdscr.Printf("%d. %s\n", i+1, gi.getKeyName(scroll.SubType))
			} else {
				gi.stdscr.Printf("%d. Scroll\n", i+1)
			}
		}
	}

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода для меню
	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100)

	for {
		char := gi.stdscr.GetChar()
		if char == goncurses.KEY_ESC {
			break
		}

		if digit := int(char - '0'); digit >= 0 {
			if itemType == Weapon && digit == 0 {
				GameActionsImpl.UseItem(itemType, 0) // снять оружие
				break
			} else if digit >= 1 && digit <= len(items) {
				if itemType == Weapon {
					// Для оружия передаем digit (1,2,3...) а не digit-1
					GameActionsImpl.UseItem(itemType, digit)
				} else {
					// Для остальных предметов передаем digit-1 (0,1,2...)
					GameActionsImpl.UseItem(itemType, digit-1)
				}
				break
			}
	gi.stdscr.GetChar()
}

func (gi *GameInterface) getKeyName(keyType ItemSubType) string {
	switch keyType {
	case KeyRed:
		return "Red Key"
	case KeyBlue:
		return "Blue Key"
	case KeyGreen:
		return "Green Key"
	case KeyYellow:
		return "Yellow Key"
	default:
		return "Key"
	}
}

func (gi *GameInterface) showEquipment() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== EQUIPMENT ===\n\n")

	player := CurrentGameState.Player

	if player.CurrentWeapon != nil {
		gi.stdscr.Printf("Current weapon: Weapon (Str+%d)\n", player.CurrentWeapon.StrengthBoost)
	} else {
		gi.stdscr.Print("Current weapon: None\n")
	}

	backpack := CurrentGameState.Backpack
	if backpack != nil && len(backpack.Weapons) > 0 {
		gi.stdscr.Print("\nAvailable weapons:\n")
		for i, weapon := range backpack.Weapons {
			gi.stdscr.Printf("%d. Weapon (Str+%d)\n", i+1, weapon.StrengthBoost)
		}

		gi.stdscr.Print("\nSelect weapon to equip (number) or any key to cancel: ")
		gi.stdscr.Refresh()

		char := gi.stdscr.GetChar()
		if char >= '1' && char <= '9' {
			index := int(char - '1')
			if index < len(backpack.Weapons) {
				// Используем существующий метод UseItem вместо EquipWeapon
				GameActionsImpl.UseItem(Weapon, index+1) // +1 потому что 0 - это unequip
				gi.showMessage("Equipped weapon")
			}
		}
	} else {
		gi.stdscr.Print("\nNo weapons in inventory\n")
		gi.stdscr.Print("\nPress any key to continue...")
		gi.stdscr.Refresh()
		gi.stdscr.GetChar()
	}
}

func (gi *GameInterface) showPickupMenu() {
	player := CurrentGameState.Player
	var itemsToPick []*Item

	for _, item := range CurrentGameState.Items {
		if item.Position.X == player.Position.X &&
			item.Position.Y == player.Position.Y {
			itemsToPick = append(itemsToPick, item)
		}
	}

	if len(itemsToPick) == 0 {
		GameActionsImpl.AddMessage("No items here to pick up")
		return
	}

	if len(itemsToPick) == 1 {
		// Используем существующий метод PickUpItem вместо PickupItem
		GameActionsImpl.PickUpItem()
		return
	}

	gi.stdscr.Erase()
	gi.stdscr.Print("=== PICK UP ITEM ===\n\n")

	for i, item := range itemsToPick {
		gi.stdscr.Printf("%d. %s\n", i+1, gi.getItemName(item))
	}

	gi.stdscr.Print("\nSelect item to pick up (number) or any key to cancel: ")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода для меню
	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100)


	char := gi.stdscr.GetChar()
	if char >= '1' && char <= '9' {
		index := int(char - '1')
		if index < len(itemsToPick) {
			GameActionsImpl.PickUpItem()
		}
	}
}

func (gi *GameInterface) getItemName(item *Item) string {
	switch item.Type {
	case Weapon:
		return "Weapon (Str+" + strconv.Itoa(item.StrengthBoost) + ")"
	case Food:
		return "Food (HP+" + strconv.Itoa(item.HealthBoost) + ")"
	case Potion:
		switch item.SubType {
		case HealthPotion:
			return "Health Potion"
		case AgilityPotion:
			return "Agility Potion"
		case StrengthPotion:
			return "Strength Potion"
		default:
			return "Potion"
		}
	case Scroll:
		switch item.SubType {
		case MaxHealthScroll:
			return "Max Health Scroll"
		case AgilityScroll:
			return "Agility Scroll"
		case StrengthScroll:
			return "Strength Scroll"
		case KeyRed:
			return "Red Key"
		case KeyBlue:
			return "Blue Key"
		case KeyGreen:
			return "Green Key"
		case KeyYellow:
			return "Yellow Key"
		default:
			return "Scroll"
		}
	case Treasure:
		return "Treasure (" + strconv.Itoa(item.Value) + " gold)"
	default:
		return "Unknown item"
	}
}

func (gi *GameInterface) showQuitMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== QUIT GAME ===\n\n")
	gi.stdscr.Print("Save game before quitting? (y/n/c): ")
	gi.stdscr.Refresh()

	// Устанавливаем блокирующий ввод
	gi.stdscr.Timeout(-1)
	char := gi.stdscr.GetChar()
	gi.stdscr.Timeout(100) // Возвращаем неблокирующий ввод

	switch char {
	case 'y', 'Y':
		gi.showSaveMenu()
		CurrentGameState.IsRunning = false
	case 'n', 'N':
		CurrentGameState.IsRunning = false
	}
}

func (gi *GameInterface) showSaveMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("Enter filename to save: ")
	gi.stdscr.Refresh()


	filename := gi.getInput()
	if filename != "" {
		err := GameActionsImpl.SaveGame(filename)
		if err != nil {
			gi.showMessage("Save error: " + err.Error())
		} else {
			gi.showMessage("Game saved successfully!")
		}
	}
}

func (gi *GameInterface) showLoadMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("Enter filename to load: ")
	gi.stdscr.Refresh()


	filename := gi.getInput()
	if filename != "" {
		err := GameActionsImpl.LoadGame(filename)
		if err != nil {
			gi.showMessage("Load error: " + err.Error())
		} else {
			gi.showMessage("Game loaded successfully!")
		}
	}
}

func (gi *GameInterface) showGameOver() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== HIGH SCORES ===\n\n")

	if CurrentGameState.Statistics == nil || len(CurrentGameState.Statistics.BestScores) == 0 {
		gi.stdscr.Print("No high scores yet!\n")
	} else {
		for i, score := range CurrentGameState.Statistics.BestScores {
			if i >= 10 { // Показываем топ-10
				break
			}
			gi.stdscr.Printf("%d. Level: %d, Treasures: %d\n",
				i+1, score.LevelReached, score.Treasures)
		}
	}

	gi.stdscr.Print("=== GAME OVER ===\n\n")
	gi.stdscr.Print("You have been defeated!\n\n")

	player := CurrentGameState.Player
	gi.stdscr.Printf("Level reached: %d\n", CurrentGameState.CurrentLevel)
	gi.stdscr.Printf("Treasures collected: %d\n", player.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", player.DefeatedEnemies)

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) showVictory() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== VICTORY! ===\n\n")
	gi.stdscr.Print("You have completed the game!\n\n")

	player := CurrentGameState.Player
	gi.stdscr.Printf("Final level: %d\n", CurrentGameState.CurrentLevel)
	gi.stdscr.Printf("Treasures collected: %d\n", player.Treasures)
	gi.stdscr.Printf("Enemies defeated: %d\n", player.DefeatedEnemies)

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода
	gi.stdscr.GetChar()
}

func (gi *GameInterface) showMessage(message string) {
	gi.stdscr.Erase()
	gi.stdscr.Print(message)
	gi.stdscr.Print("\n\nPress any key to continue...")
	gi.stdscr.Refresh()

	// Устанавливаем блокирующий ввод
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100) // Возвращаем неблокирующий ввод
}

func (gi *GameInterface) getInput() string {
	// Временное включение эха и блокирующего ввода
	goncurses.Echo(true)
	gi.stdscr.Timeout(-1)
	defer func() {
		goncurses.Echo(false)
		gi.stdscr.Timeout(100)
	}()

	// Явно отключаем echo чтобы избежать дублирования
	goncurses.Echo(false)

	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100)

	input := ""
	cursorPos := 0

	// Очищаем строку и показываем курсор
	y, x := gi.stdscr.CursorYX()
	gi.stdscr.Print("> ")
	cursorPos = x + 2
	gi.stdscr.Refresh()

	for {
		char := gi.stdscr.GetChar()

		if char == '\n' || char == '\r' {
			break
		} else if char == 127 || char == 8 { // Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
				cursorPos--
				gi.stdscr.Move(y, cursorPos)
				gi.stdscr.Print(" ")
				gi.stdscr.Move(y, cursorPos)
			}
		} else {
			input += string(char)
		}
	}
	return input
}

func (gi *GameInterface) showMessage(message string) {
	gi.stdscr.Erase()
	gi.stdscr.Print(message + "\n\nPress any key to continue...")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) showGameOver() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== GAME OVER ===\n\n")
	gi.stdscr.Print("Level reached: " + strconv.Itoa(CurrentGameState.CurrentLevel) + "\n")
	gi.stdscr.Print("Treasures collected: " + strconv.Itoa(CurrentGameState.Player.Treasures) + "\n")
	gi.stdscr.Print("Enemies defeated: " + strconv.Itoa(CurrentGameState.Player.DefeatedEnemies) + "\n\n")
	gi.stdscr.Print("Press any key to exit...")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) showVictory() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== VICTORY! ===\n\n")
	gi.stdscr.Print("You completed all 21 levels!\n")
	gi.stdscr.Print("Total treasures: " + strconv.Itoa(CurrentGameState.Player.Treasures) + "\n")
	gi.stdscr.Print("Total enemies defeated: " + strconv.Itoa(CurrentGameState.Player.DefeatedEnemies) + "\n\n")
	gi.stdscr.Print("Press any key to exit...")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
		} else if char >= 32 && char <= 126 {
			input += string(byte(char))
			gi.stdscr.Print(string(byte(char)))
			cursorPos++
		}

		gi.stdscr.Refresh()
	}

	return input
}

func (gi *GameInterface) getItemSymbol(item *Item) string {
	switch item.Type {
	case Weapon:
		return "!"
	case Food:
		return "%"
	case Potion:
		return "&"
	case Scroll:
		return "?"
	case Treasure:
		return "$"
	default:
		return "?"
	}
}

func (gi *GameInterface) getItemColor(item *Item) int16 {
	switch item.Type {
	case Weapon:
		return 4
	case Food:
		return 2
	case Potion:
		return 5
	case Scroll:
		return 3
	case Treasure:
		return 6
	default:
		return 7
	}
}

func (gi *GameInterface) getEnemySymbol(enemy *Enemy) string {
	if enemy.Type == Mimic && enemy.IsDisguised {
		// Мимик маскируется под случайный предмет
		itemTypes := []string{"!", "%", "&", "?", "$"}
		return itemTypes[rand.Intn(len(itemTypes))]
	}

	switch enemy.Type {
	case Zombie:
		return "Z"
	case Vampire:
		return "V"
	case Ghost:
		return "G"
	case Ogre:
		return "O"
	case SnakeMage:
		return "S"
	case Mimic:
		return "M"
	default:
		return "E"
	}
}

func (gi *GameInterface) getEnemyColor(enemy *Enemy) int16 {
	switch enemy.Type {
	case Zombie:
		return 2
	case Vampire:
		return 1
	case Ghost:
		return 7
	case Ogre:
		return 4
	case SnakeMage:
		return 5
	case Mimic:
		return 8 // Новый цвет для мимика
	default:
		return 7
	}
}

func (gi *GameInterface) initColors() {
	goncurses.InitPair(1, goncurses.C_RED, goncurses.C_BLACK)     // Red - vampires
	goncurses.InitPair(2, goncurses.C_GREEN, goncurses.C_BLACK)   // Green - zombies, corridors
	goncurses.InitPair(3, goncurses.C_YELLOW, goncurses.C_BLACK)  // Yellow - ogres, treasures
	goncurses.InitPair(4, goncurses.C_CYAN, goncurses.C_BLACK)    // Cyan - food
	goncurses.InitPair(5, goncurses.C_BLUE, goncurses.C_BLACK)    // Blue - potions
	goncurses.InitPair(6, goncurses.C_MAGENTA, goncurses.C_BLACK) // Purple - snake mages, scrolls
	goncurses.InitPair(7, goncurses.C_WHITE, goncurses.C_BLACK)   // White - ghosts, weapons
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

	goncurses.InitPair(1, goncurses.C_RED, goncurses.C_BLACK)     // Стены
	goncurses.InitPair(2, goncurses.C_GREEN, goncurses.C_BLACK)   // Коридоры, еда
	goncurses.InitPair(3, goncurses.C_YELLOW, goncurses.C_BLACK)  // Выход, свитки
	goncurses.InitPair(4, goncurses.C_BLUE, goncurses.C_BLACK)    // Оружие
	goncurses.InitPair(5, goncurses.C_MAGENTA, goncurses.C_BLACK) // Зелья, змей-маг
	goncurses.InitPair(6, goncurses.C_CYAN, goncurses.C_BLACK)    // Игрок, сокровища
	goncurses.InitPair(7, goncurses.C_WHITE, goncurses.C_BLACK)   // Призрак, по умолчанию
	goncurses.InitPair(8, goncurses.C_BLACK, goncurses.C_WHITE)   // Mimic (инвертированные цвета)
}
