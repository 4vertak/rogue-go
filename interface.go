package main

import (
	"os"
	"strconv"
	"strings"
	"time"

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
	if CurrentGameState.Camera != nil {
		maxY, maxX := gi.gameWindow.MaxYX()
		CurrentGameState.Camera.Width = maxX - 2  // usable width
		CurrentGameState.Camera.Height = maxY - 2 // usable height
	}
}

func (gi *GameInterface) Run() {
	defer goncurses.End()

	gi.renderInitialScreen()

	// Неблокирующий ввод
	gi.stdscr.Timeout(100) // 100ms timeout

	for CurrentGameState.IsRunning {
		gi.renderGame()
		gi.handleInput()

		if !CurrentGameState.GameOver && !CurrentGameState.GameWon {
			GameActionsImpl.UpdateGameState()
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
		camera.X = mapWidth - usableWidth
	}
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
		camera.X = mapWidth - usableWidth
	}
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
						ch = gi.getItemSymbol(item)
						color = gi.getItemColor(item)
					}
				}

				// Проверка врагов
				for _, enemy := range CurrentGameState.Enemies {
					if enemy.Position.X == x && enemy.Position.Y == y && (enemy.IsVisible || enemy.Type == Mimic) {

						ch = gi.getEnemySymbol(enemy)
						color = gi.getEnemyColor(enemy)

					}
				}

				// Проверка игрока
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
						color = 3 // Yellow
					}
				}

				if color > 0 {
					gi.gameWindow.ColorOn(color)
				}
				gi.gameWindow.MovePrint(screenY, screenX, ch)
				if color > 0 {
					gi.gameWindow.ColorOff(color)
				}
			} else if CurrentGameState.Level.Discovered[x][y] {
				gi.gameWindow.MovePrint(screenY, screenX, " ")
			} else {
				gi.gameWindow.MovePrint(screenY, screenX, " ")
			}
		}
	}

	gi.gameWindow.Refresh()
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

	backpack := CurrentGameState.Backpack
	gi.inventoryWindow.MovePrint(1, 1, "F:"+strconv.Itoa(len(backpack.Food)))
	gi.inventoryWindow.MovePrint(1, 5, "P:"+strconv.Itoa(len(backpack.Potions)))
	gi.inventoryWindow.MovePrint(1, 9, "S:"+strconv.Itoa(len(backpack.Scrolls)))
	gi.inventoryWindow.MovePrint(1, 13, "W:"+strconv.Itoa(len(backpack.Weapons)))
	gi.inventoryWindow.MovePrint(1, 17, "T:"+strconv.Itoa(backpack.Treasure))

	gi.inventoryWindow.Refresh()
}

func (gi *GameInterface) handleInput() {
	char := gi.stdscr.GetChar()

	// Проверяем на ошибку ввода (таймаут) - в новой версии goncurses
	// GetChar() возвращает -1 при таймауте, но как rune
	if char == goncurses.Key(-1) {
		return // Нет ввода
	}

	switch char {
	case 'w', 'W':
		GameActionsImpl.MovePlayer(0, -1)
	case 's', 'S':
		GameActionsImpl.MovePlayer(0, 1)
	case 'a', 'A':
		GameActionsImpl.MovePlayer(-1, 0)
	case 'd', 'D':
		GameActionsImpl.MovePlayer(1, 0)
	case 'p', 'P': // Поднять предмет
		GameActionsImpl.PickUpItem()
	case 'h', 'H': // Использовать оружие
		gi.showWeaponSelection()
	case 'j', 'J': // Использовать еду
		gi.showFoodSelection()
	case 'k', 'K': // Использовать зелье
		gi.showPotionSelection()
	case 'e', 'E': // Использовать свиток
		gi.showScrollSelection()
	case goncurses.KEY_ESC:
		gi.showMenu()
	}
}

// Остальные функции остаются без изменений...

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
		}
	}
}

func (gi *GameInterface) showMenu() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== MENU ===\n\n")
	gi.stdscr.Print("1. Continue game\n")
	gi.stdscr.Print("2. Save game\n")
	gi.stdscr.Print("3. Load game\n")
	gi.stdscr.Print("4. High scores\n")
	gi.stdscr.Print("5. Return to main menu\n")
	gi.stdscr.Print("6. Exit game\n\n")
	gi.stdscr.Print("Select option: ")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода для меню
	gi.stdscr.Timeout(-1)
	defer gi.stdscr.Timeout(100)

	char := gi.stdscr.GetChar()
	switch char {
	case '1':
		return // Продолжить игру
	case '2':
		gi.saveGameDialog()
	case '3':
		gi.loadGameDialog()
	case '4':
		gi.showHighScores()
	case '5':
		CurrentGameState.IsRunning = false
	case '6':
		CurrentGameState.IsRunning = false
		goncurses.End()
		os.Exit(0)
	}
}

func (gi *GameInterface) saveGameDialog() {
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

func (gi *GameInterface) loadGameDialog() {
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

func (gi *GameInterface) showHighScores() {
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

	gi.stdscr.Print("\nPress any key to continue...")
	gi.stdscr.Refresh()

	// Временное включение блокирующего ввода
	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) getInput() string {
	// Временное включение эха и блокирующего ввода
	goncurses.Echo(true)
	gi.stdscr.Timeout(-1)
	defer func() {
		goncurses.Echo(false)
		gi.stdscr.Timeout(100)
	}()

	input := ""
	for {
		char := gi.stdscr.GetChar()
		if char == '\n' || char == '\r' {
			break
		} else if char == goncurses.KEY_BACKSPACE || char == 127 {
			if len(input) > 0 {
				input = input[:len(input)-1]
				gi.stdscr.MovePrint(gi.stdscr.CursorYX())
				gi.stdscr.Print(" ")
				gi.stdscr.Move(gi.stdscr.CursorYX())
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
}

func (gi *GameInterface) getItemSymbol(item *Item) string {
	switch item.Type {
	case Treasure:
		return "$"
	case Food:
		return "f"
	case Potion:
		return "p"
	case Scroll:
		return "s"
	case Weapon:
		return "w"
	default:
		return "?"
	}
}

func (gi *GameInterface) getItemColor(item *Item) int16 {
	switch item.Type {
	case Treasure:
		return 3 // Yellow
	case Food:
		return 4 // Green
	case Potion:
		return 5 // Blue
	case Scroll:
		return 6 // Purple
	case Weapon:
		return 7 // White
	default:
		return 0
	}
}

func (gi *GameInterface) getEnemySymbol(enemy *Enemy) string {
	switch enemy.Type {
	case Zombie:
		return "z"
	case Vampire:
		return "v"
	case Ghost:
		return "g"
	case Ogre:
		return "O"
	case SnakeMage:
		return "s"
	case Mimic:
		if enemy.IsVisible {
			return "M"
		} else {
			koef := (enemy.Health - 20) / 5
			if koef < 5 {
				return "$"
			} else if koef < 9 {
				return "f"
			} else if koef < 14 {
				return "p"
			} else if koef < 18 {
				return "s"
			} else {
				return "w"
			}
		}
	default:
		return "e"
	}
}

func (gi *GameInterface) getEnemyColor(enemy *Enemy) int16 {
	switch enemy.Type {
	case Zombie:
		return 2 // Green
	case Vampire:
		return 1 // Red
	case Ghost:
		return 7 // White
	case Ogre:
		return 3 // Yellow
	case SnakeMage:
		return 6 // Purple
	default:
		return 0
	}
}

func (gi *GameInterface) getEffectSymbol(effect *Effect) string {
	switch effect.Type {
	case AgilityEffect:
		return "*"
	case StrengthEffect:
		return "+"
	case MaxHealthEffect:
		return "H"
	default:
		return "!"
	}
}

func (gi *GameInterface) getItemDescription(item *Item) string {
	switch item.Type {
	case Food:
		return "Food (+" + strconv.Itoa(item.HealthBoost) + " HP)"
	case Potion:
		switch item.SubType {
		case HealthPotion:
			return "Health Potion (+" + strconv.Itoa(item.HealthBoost) + " HP)"
		case AgilityPotion:
			return "Agility Potion (+" + strconv.Itoa(item.AgilityBoost) + " AGI)"
		case StrengthPotion:
			return "Strength Potion (+" + strconv.Itoa(item.StrengthBoost) + " STR)"
		default:
			return "Unknown Potion"
		}
	case Scroll:
		switch item.SubType {
		case MaxHealthScroll:
			return "Health Scroll (+" + strconv.Itoa(item.MaxHealthBoost) + " MaxHP)"
		case AgilityScroll:
			return "Agility Scroll (+" + strconv.Itoa(item.AgilityBoost) + " AGI)"
		case StrengthScroll:
			return "Strength Scroll (+" + strconv.Itoa(item.StrengthBoost) + " STR)"
		default:
			return "Unknown Scroll"
		}
	case Weapon:
		return "Weapon (+" + strconv.Itoa(item.StrengthBoost) + " STR)"
	case Treasure:
		return "Treasure (" + strconv.Itoa(item.Value) + " gold)"
	default:
		return "Unknown Item"
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
