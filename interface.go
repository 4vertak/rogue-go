package main

import (
    "math"
	"math/rand"
	"strconv"
	"time"

	"github.com/rthornton128/goncurses"
)

// Добавим константы для текстур
const (
    TextureBrick = iota
    TextureStone
    TextureWood
    TextureMarble
    NumTextures
)

// Добавим структуру для текстуры
type Texture struct {
    Symbols []string
    Colors  []int16
}

// GameInterface реализует пользовательский интерфейс
type GameInterface struct {
	stdscr          *goncurses.Window
	gameWindow      *goncurses.Window
	statusWindow    *goncurses.Window
	messageWindow   *goncurses.Window
	inventoryWindow *goncurses.Window
    showMinimap     bool // Добавляем флаг для мини-карты
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

    gi.showMinimap = true
	
	return nil
}

func (gi *GameInterface) createWindows() {
	rows, cols := gi.stdscr.MaxYX()
	
	// Главное игровое окно (центр экрана)
	gameRows := rows - 6
	gameCols := cols - 31
	if gameRows < 10 { gameRows = 10 }
	if gameCols < 20 { gameCols = 20 }
	
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
    gi.stdscr.Timeout(100)
    
    for CurrentGameState.IsRunning {
        gi.renderGame()
        gi.handleInput()
        
        if !CurrentGameState.GameOver {
            GameActionsImpl.UpdateGameState()
        }
        
        if CurrentGameState.GameOver {
            gi.showGameOver()
            break
        }
        
        // УБИРАЕМ проверку GameWon здесь, так как переход на следующий уровень
        // теперь обрабатывается через nextLevel()
        
        time.Sleep(50 * time.Millisecond)
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
    if CurrentGameState.RenderMode == Render3D {
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
	
	player := CurrentGameState.Player
	camera := CurrentGameState.Camera
	
	if camera == nil || CurrentGameState.Level == nil {
		return
	}
	
	maxWinY, maxWinX := gi.gameWindow.MaxYX()
	usableWidth := maxWinX - 2
	usableHeight := maxWinY - 2
	
	camera.X = player.Position.X - usableWidth/2
	camera.Y = player.Position.Y - usableHeight/2
	
	if camera.X < 0 { camera.X = 0 }
	if camera.Y < 0 { camera.Y = 0 }
	
	mapWidth := len(CurrentGameState.Level.Grid)
	if mapWidth == 0 { return }
	mapHeight := len(CurrentGameState.Level.Grid[0])
	
	if camera.X + usableWidth > mapWidth {
		camera.X = mapWidth - usableWidth
	}
	if camera.Y + usableHeight > mapHeight {
		camera.Y = mapHeight - usableHeight
	}
	
	if camera.X < 0 { camera.X = 0 }
	if camera.Y < 0 { camera.Y = 0 }
	if camera.X > mapWidth - usableWidth && mapWidth > usableWidth {
		camera.X = mapWidth - usableWidth
	}
	if camera.Y > mapHeight - usableHeight && mapHeight > usableHeight {
		camera.Y = mapHeight - usableHeight
	}
	
	for x := camera.X; x < camera.X+usableWidth && x < len(CurrentGameState.Level.Grid); x++ {
		for y := camera.Y; y < camera.Y+usableHeight && y < len(CurrentGameState.Level.Grid[0]); y++ {
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
					if enemy.Position.X == x && enemy.Position.Y == y && enemy.IsVisible {
						ch = gi.getEnemySymbol(enemy)
						color = gi.getEnemyColor(enemy)
						break
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
			} else if CurrentGameState.Level.Discovered[x][y] {
				gi.gameWindow.MovePrint(screenY, screenX, " ")
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
    
    // Отображаем статус сна
    if CurrentGameState.SleepTurns > 0 {
        gi.statusWindow.MovePrint(15, 1, "Sleep: "+strconv.Itoa(CurrentGameState.SleepTurns)+" turns")
    }
    
    gi.statusWindow.Refresh()
}

func (gi *GameInterface) renderStatusMessages() {
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
	
	if CurrentGameState.Messages != nil && len(CurrentGameState.Messages) > 0 {
		lastMessage := CurrentGameState.Messages[len(CurrentGameState.Messages)-1]
		gi.messageWindow.MovePrint(1, 1, lastMessage)
	}
	
	gi.messageWindow.Refresh()
}

func (gi *GameInterface) renderInventory() {
	gi.inventoryWindow.Erase()
	gi.inventoryWindow.Box(0, 0)
	gi.inventoryWindow.MovePrint(1, 1, "Inventory:")
	
	backpack := CurrentGameState.Backpack
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

    if CurrentGameState.RenderMode == Render3D {
        gi.handle3DInput(char)
        return
    }
    
    // Оригинальная обработка для 2D режима
    switch char {
    case goncurses.KEY_UP, 'w', 'W':
        GameActionsImpl.MovePlayer(0, -1)
    case goncurses.KEY_DOWN, 's', 'S':
        GameActionsImpl.MovePlayer(0, 1)
    case goncurses.KEY_LEFT, 'a', 'A':
        GameActionsImpl.MovePlayer(-1, 0)
    case goncurses.KEY_RIGHT, 'd', 'D':
        GameActionsImpl.MovePlayer(1, 0)
    case 'h', 'H':
        gi.showUseItemMenu(Weapon)
    case 'j', 'J':
        gi.showUseItemMenu(Food)
    case 'k', 'K':
        gi.showUseItemMenu(Potion)
    case 'e', 'E':
        gi.showUseItemMenu(Scroll)
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
    case 't', 'T': // Переключение в 3D режим
        CurrentGameState.RenderMode = Render3D
        // Инициализируем направление взгляда, если не инициализировано
        if CurrentGameState.PlayerDir.X == 0 && CurrentGameState.PlayerDir.Y == 0 {
            CurrentGameState.PlayerDir = PositionF{X: -1, Y: 0}
            CurrentGameState.PlayerPlane = PositionF{X: 0, Y: 0.66}
        }
        gi.AddMessage("Switched to 3D mode")
    }
}

func (gi *GameInterface) handle3DInput(char goncurses.Key) {
    moveSpeed := 0.1
    rotSpeed := 0.05

    switch char {
    case 'w', 'W':
        gi.moveIn3D(moveSpeed)
    case 's', 'S':
        gi.moveIn3D(-moveSpeed)
    case 'a', 'A':
        // Страф влево
        gi.strafe(-moveSpeed)
        
    case 'd', 'D':
        // Страф вправо
        gi.strafe(moveSpeed)
    
    case goncurses.KEY_LEFT:
        // Поворот налево
        oldDirX := CurrentGameState.PlayerDir.X
        CurrentGameState.PlayerDir.X = CurrentGameState.PlayerDir.X*math.Cos(-rotSpeed) - CurrentGameState.PlayerDir.Y*math.Sin(-rotSpeed)
        CurrentGameState.PlayerDir.Y = oldDirX*math.Sin(-rotSpeed) + CurrentGameState.PlayerDir.Y*math.Cos(-rotSpeed)
        
        oldPlaneX := CurrentGameState.PlayerPlane.X
        CurrentGameState.PlayerPlane.X = CurrentGameState.PlayerPlane.X*math.Cos(-rotSpeed) - CurrentGameState.PlayerPlane.Y*math.Sin(-rotSpeed)
        CurrentGameState.PlayerPlane.Y = oldPlaneX*math.Sin(-rotSpeed) + CurrentGameState.PlayerPlane.Y*math.Cos(-rotSpeed)
        
    case goncurses.KEY_RIGHT:
        // Поворот направо
        oldDirX := CurrentGameState.PlayerDir.X
        CurrentGameState.PlayerDir.X = CurrentGameState.PlayerDir.X*math.Cos(rotSpeed) - CurrentGameState.PlayerDir.Y*math.Sin(rotSpeed)
        CurrentGameState.PlayerDir.Y = oldDirX*math.Sin(rotSpeed) + CurrentGameState.PlayerDir.Y*math.Cos(rotSpeed)
        
        oldPlaneX := CurrentGameState.PlayerPlane.X
        CurrentGameState.PlayerPlane.X = CurrentGameState.PlayerPlane.X*math.Cos(rotSpeed) - CurrentGameState.PlayerPlane.Y*math.Sin(rotSpeed)
        CurrentGameState.PlayerPlane.Y = oldPlaneX*math.Sin(rotSpeed) + CurrentGameState.PlayerPlane.Y*math.Cos(rotSpeed)
    
    case 't', 'T':
        CurrentGameState.RenderMode = Render2D
        gi.AddMessage("Switched to 2D mode")

    case 'm', 'M':
        gi.showMinimap = !gi.showMinimap
        if gi.showMinimap {
            gi.AddMessage("Mini-map enabled")
        } else {
            gi.AddMessage("Mini-map disabled")
        }
    
    // Действия с инвентарем - меняем клавиши для 3D режима
    case '1':
        gi.showUseItemMenu(Weapon)
    case '2':
        gi.showUseItemMenu(Food)
    case '3':
        gi.showUseItemMenu(Potion)
    case '4':
        gi.showUseItemMenu(Scroll)
    case '5':
        GameActionsImpl.PickUpItem()
    case 'i', 'I':
        gi.showInventory()
    
    // Быстрое поворачивание (новые клавиши)
    case 'q', 'Q':
        CurrentGameState.PlayerAngle -= 45 // Быстрый поворот налево
        if CurrentGameState.PlayerAngle < 0 {
            CurrentGameState.PlayerAngle += 360
        }
    case 'r', 'R': // Вместо 'e' для быстрого поворота направо
        CurrentGameState.PlayerAngle += 45 
        if CurrentGameState.PlayerAngle >= 360 {
            CurrentGameState.PlayerAngle -= 360
        }
    
    // Взаимодействие с дверьми
    case ' ', '\n', '\r': // Пробел или Enter
        gi.interactWithDoor()
    
    // Бег
    case goncurses.KEY_UP:
        gi.moveIn3D(moveSpeed) // Бег вперед
    case goncurses.KEY_DOWN:
        gi.moveIn3D(-moveSpeed) // Бег назад
    
    // Выход из игры
    case goncurses.KEY_ESC:
        gi.showQuitMenu()
    }
}

func (gi *GameInterface) moveIn3D(distance float64) {
    moveSpeed := distance * 0.1 // Увеличим скорость движения
    player := CurrentGameState.Player
    dir := CurrentGameState.PlayerDir

    // Правильно вычисляем новую позицию с учетом направления взгляда
    newX := float64(player.Position.X) + dir.X*moveSpeed
    newY := float64(player.Position.Y) + dir.Y*moveSpeed

    // Проверяем возможность движения
    if gi.canMoveTo3D(int(newX), int(newY)) {
        player.Position.X = int(newX)
        player.Position.Y = int(newY)
        GameActionsImpl.AddMessage("Moved forward")
    } else {
        GameActionsImpl.AddMessage("Cannot move - obstacle ahead")
    }
}

// Добавим функцию для движения назад
func (gi *GameInterface) moveBackward(distance float64) {
    moveSpeed := distance * 0.1
    player := CurrentGameState.Player
    dir := CurrentGameState.PlayerDir

    // Двигаемся в противоположном направлении
    newX := float64(player.Position.X) - dir.X*moveSpeed
    newY := float64(player.Position.Y) - dir.Y*moveSpeed

    if gi.canMoveTo3D(int(newX), int(newY)) {
        player.Position.X = int(newX)
        player.Position.Y = int(newY)
        GameActionsImpl.AddMessage("Moved backward")
    }
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
    gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
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
        return 13 // Используем новую пару
    default:
        return 7
    }
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
    goncurses.InitPair(13, goncurses.C_BLACK, goncurses.C_WHITE)  // Mimic
    
    // Цвета для 3D режима
    goncurses.InitPair(8, goncurses.C_BLACK, goncurses.C_BLACK)   // Потолок
    goncurses.InitPair(9, goncurses.C_WHITE, goncurses.C_BLACK)   // Пол
    goncurses.InitPair(10, goncurses.C_RED, goncurses.C_BLACK)    // Текстура стен 1
    goncurses.InitPair(11, goncurses.C_BLUE, goncurses.C_BLACK)   // Текстура стен 2  
    goncurses.InitPair(12, goncurses.C_GREEN, goncurses.C_BLACK)  // Текстура стен 3

    // Изменяем инициализацию цветов для текстур
    goncurses.InitPair(20, goncurses.C_RED, goncurses.C_BLACK)       // Кирпич
    goncurses.InitPair(21, goncurses.C_WHITE, goncurses.C_BLACK)     // Камень (бывший GRAY)
    goncurses.InitPair(22, goncurses.C_YELLOW, goncurses.C_BLACK)    // Дерево
    goncurses.InitPair(23, goncurses.C_CYAN, goncurses.C_BLACK)      // Мрамор (бывший WHITE)
    
    // Градиентные цвета для стен
    goncurses.InitPair(24, goncurses.C_RED, goncurses.C_BLACK)       // Яркий
    goncurses.InitPair(25, goncurses.C_MAGENTA, goncurses.C_BLACK)   // Средний
    goncurses.InitPair(26, goncurses.C_BLUE, goncurses.C_BLACK)      // Темный
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

func (gi *GameInterface) render3D() {
    gi.gameWindow.Erase()
    
    player := CurrentGameState.Player
    level := CurrentGameState.Level
    maxY, maxX := gi.gameWindow.MaxYX()
    width, height := maxX-2, maxY-2

    // Максимальное расстояние видимости
    maxViewDistance := 15.0

    // Рендерим столбы для каждой колонки экрана
    for x := 0; x < width; x++ {
        // Вычисляем положение луча
        cameraX := 2*float64(x)/float64(width) - 1
        rayDirX := CurrentGameState.PlayerDir.X + CurrentGameState.PlayerPlane.X*cameraX
        rayDirY := CurrentGameState.PlayerDir.Y + CurrentGameState.PlayerPlane.Y*cameraX

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

// Добавим функцию для рендеринга дверей в 3D
func (gi *GameInterface) renderDoors3D() {
    maxY, maxX := gi.gameWindow.MaxYX()
    screenWidth := maxX - 2
    screenHeight := maxY - 2
    
    for _, door := range CurrentGameState.Level.Doors {
        // Проверяем видимость двери
        if !CurrentGameState.Level.Visible[door.Position.X][door.Position.Y] {
            continue
        }

        // Вычисляем позицию двери относительно игрока
        relX := door.Position.X - CurrentGameState.Player.Position.X
        relY := door.Position.Y - CurrentGameState.Player.Position.Y
        
        // Преобразуем в экранные координаты (упрощенная версия)
        angle := math.Atan2(float64(relY), float64(relX)) - 
                math.Atan2(CurrentGameState.PlayerDir.Y, CurrentGameState.PlayerDir.X)
        
        // Нормализуем угол
        for angle < -math.Pi {
            angle += 2 * math.Pi
        }
        for angle > math.Pi {
            angle -= 2 * math.Pi
        }

        // Если дверь в поле зрения
        if math.Abs(angle) < math.Pi/3 {
            screenX := int((angle + math.Pi/3) * float64(screenWidth) / (2*math.Pi/3))
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
            doorColor := gi.getKeyColor(door.Color)
            for y := startY; y < startY+height && y < screenHeight; y++ {
                gi.gameWindow.ColorOn(doorColor)
                gi.gameWindow.MovePrint(y+1, screenX+1, "D")
                gi.gameWindow.ColorOff(doorColor)
            }
        }
    }
}

func (gi *GameInterface) castRay(startPos Position, angle float64, level *Level) (float64, Cell) {
    directionX := math.Cos(angle * math.Pi / 180)
    directionY := math.Sin(angle * math.Pi / 180)
    
    posX := float64(startPos.X) + 0.5
    posY := float64(startPos.Y) + 0.5
    
    // Простой алгоритм для начала
    for distance := 0.1; distance < 20.0; distance += 0.1 {
        checkX := int(posX + distance*directionX)
        checkY := int(posY + distance*directionY)
        
        if checkX < 0 || checkX >= len(level.Grid) || 
           checkY < 0 || checkY >= len(level.Grid[0]) {
            return distance, Cell{IsWall: true}
        }
        
        cell := level.Grid[checkX][checkY]
        if cell.IsWall {
            return distance, cell
        }
    }
    
    return 20.0, Cell{IsWall: false}
}

func (gi *GameInterface) renderMinimap() {
    _, maxX := gi.gameWindow.MaxYX() // Правильно получаем оба значения
    mapSize := 8 // Уменьшим размер мини-карты для 3D режима
    
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
            mapX := CurrentGameState.Player.Position.X - mapSize/2 + x - 1
            mapY := CurrentGameState.Player.Position.Y - mapSize/2 + y - 1
            
            var ch string
            var color int16 = 0
            
            if mapX >= 0 && mapX < len(CurrentGameState.Level.Grid) &&
               mapY >= 0 && mapY < len(CurrentGameState.Level.Grid[0]) {
                
                cell := CurrentGameState.Level.Grid[mapX][mapY]
                
                if cell.IsWall {
                    ch = "#"
                    color = 1 // Красный для стен
                } else {
                    ch = "."
                    color = 2 // Зеленый для пола
                }
                
                // Проверяем дверь
                for _, door := range CurrentGameState.Level.Doors {
                    if door.Position.X == mapX && door.Position.Y == mapY {
                        ch = "D"
                        color = gi.getKeyColor(door.Color)
                        break
                    }
                }
                
                // Проверяем предметы
                for _, item := range CurrentGameState.Items {
                    if item.Position.X == mapX && item.Position.Y == mapY {
                        ch = gi.getItemSymbol(item)
                        color = gi.getItemColor(item)
                        break
                    }
                }
                
                // Проверяем врагов
                for _, enemy := range CurrentGameState.Enemies {
                    if enemy.Position.X == mapX && enemy.Position.Y == mapY && enemy.IsVisible {
                        ch = gi.getEnemySymbol(enemy)
                        color = gi.getEnemyColor(enemy)
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

func (gi *GameInterface) getWallColor(cell Cell, side int) int16 {
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

// Исправим функцию strafe для бокового движения
func (gi *GameInterface) strafe(distance float64) {
    moveSpeed := distance * 0.1
    player := CurrentGameState.Player
    dir := CurrentGameState.PlayerDir

    // Перпендикулярное направление для страфа
    perpX := -dir.Y
    perpY := dir.X

    newX := float64(player.Position.X) + perpX*moveSpeed
    newY := float64(player.Position.Y) + perpY*moveSpeed

    if gi.canMoveTo3D(int(newX), int(newY)) {
        player.Position.X = int(newX)
        player.Position.Y = int(newY)
        GameActionsImpl.AddMessage("Strafed")
    }
}

func (gi *GameInterface) interactWithDoor() {
    // Проверяем дверь перед игроком
    angleRad := CurrentGameState.PlayerAngle * math.Pi / 180
    checkX := CurrentGameState.Player.Position.X + int(1.5*math.Cos(angleRad))
    checkY := CurrentGameState.Player.Position.Y + int(1.5*math.Sin(angleRad))
    
    // Проходим по всем дверям без использования индекса
    doorFound := false
    for _, door := range CurrentGameState.Level.Doors {
        if door.Position.X == checkX && door.Position.Y == checkY {
            doorFound = true
            if door.IsLocked {
                gi.AddMessage("The door is locked")
            } else {
                // Открываем/закрываем дверь
                gi.AddMessage("You open the door")
            }
            break
        }
    }
    
    if !doorFound {
        gi.AddMessage("No door nearby")
    }
}

func (gi *GameInterface) canMoveTo3D(x, y int) bool {
    // Проверяем границы карты
    if x < 0 || x >= len(CurrentGameState.Level.Grid) ||
       y < 0 || y >= len(CurrentGameState.Level.Grid[0]) {
        GameActionsImpl.AddMessage("Cannot move - edge of map")
        return false
    }
    
    // Проверяем стену
    if CurrentGameState.Level.Grid[x][y].IsWall {
        return false
    }
    
    // Проверяем заблокированную дверь
    for _, door := range CurrentGameState.Level.Doors {
        if door.Position.X == x && door.Position.Y == y && door.IsLocked {
            return false
        }
    }
    
    // Проверяем врагов
    for _, enemy := range CurrentGameState.Enemies {
        if enemy.Position.X == x && enemy.Position.Y == y {
            GameActionsImpl.AddMessage("Enemy blocks the way!")
            return false
        }
    }
    
    return true
}

func (gi *GameInterface) AddMessage(message string) {
    GameActionsImpl.AddMessage(message)
}