package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"
	"time"
)

// GameLogic реализует бизнес-логику игры
type GameLogic struct{}

func init() {
	GameActionsImpl = &GameLogic{}
	rand.Seed(time.Now().UnixNano())
}

func (gl *GameLogic) MovePlayer(dx, dy int) {
	newX := CurrentGameState.Player.Position.X + dx
	newY := CurrentGameState.Player.Position.Y + dy

	if gl.canMoveTo(newX, newY) {
		CurrentGameState.Player.Position.X = newX
		CurrentGameState.Player.Position.Y = newY
		CurrentGameState.GameSession.CellsWalked++
		gl.updateVisibility()
		gl.checkForItems()
		gl.checkForEnemies() // Бой при движении игрока
		gl.checkForExit()
	}
	// ждем свой очереди для хода
	if !CurrentGameState.IsPlayerTurn {
		gl.AddMessage("Wait your turn!")
		return
	}

	// Проверяем, не усыплен ли игрок
	if CurrentGameState.SleepTurns > 0 {
		CurrentGameState.SleepTurns--
		gl.AddMessage("You are asleep for " + strconv.Itoa(CurrentGameState.SleepTurns) + " more turns")
		return
	}

	newX := CurrentGameState.Player.Position.X + dx
	newY := CurrentGameState.Player.Position.Y + dy

	if gl.canMoveTo(newX, newY) {
		CurrentGameState.Player.Position.X = newX
		CurrentGameState.Player.Position.Y = newY
		CurrentGameState.GameSession.CellsWalked++
		gl.updateVisibility()

		// Автоматически подбираем ВСЕ предметы на клетке, КРОМЕ КЛЮЧЕЙ
		picked := false
		for i := len(CurrentGameState.Items) - 1; i >= 0; i-- {
			item := CurrentGameState.Items[i]
			if item.Position.X == newX && item.Position.Y == newY {
				// Ключи НЕ подбираются автоматически!
				if item.Type == Scroll && (item.SubType == KeyRed || item.SubType == KeyBlue ||
					item.SubType == KeyGreen || item.SubType == KeyYellow) {
					gl.AddMessage("You see a key here (press 'p' to pick up)")
					continue
				}

				if gl.canAddToBackpack(item) {
					gl.addToBackpack(item)
					CurrentGameState.Items = append(CurrentGameState.Items[:i], CurrentGameState.Items[i+1:]...)
					picked = true
				} else {
					gl.AddMessage("Item here but backpack is full for this type")
				}
			}
		}
		if picked {
			gl.AddMessage("You picked up items automatically")
		}

		// После движения проверяем автоматический переход на следующий уровень
		gl.checkForAutomaticExit()
		gl.checkForEnemies() // Бой при движении игрока
		gl.checkForExit()

	}
	CurrentGameState.IsPlayerTurn = false
}

func (gl *GameLogic) EnemyTurn() {
	if CurrentGameState.IsPlayerTurn {
		return
	}

	// Логика движения врагов
	gl.moveEnemies()

	// После хода врагов
	CurrentGameState.IsPlayerTurn = true
}

func (gl *GameLogic) checkForAutomaticExit() {
	if CurrentGameState.SleepTurns > 0 {
		return
	}

	playerPos := CurrentGameState.Player.Position
	exitRoom := CurrentGameState.Level.ExitRoom

	if exitRoom == nil {
		return
	}

	// Проверяем конкретную клетку выхода, а не всю комнату
	exitX := exitRoom.X + exitRoom.Width/2
	exitY := exitRoom.Y + exitRoom.Height/2

	playerAtExit := playerPos.X == exitX && playerPos.Y == exitY

	if playerAtExit {
		// Проверяем, что все враги побеждены
		if len(CurrentGameState.Enemies) == 0 {
			// АВТОМАТИЧЕСКИ переходим на следующий уровень
			gl.nextLevel()
		} else {
			gl.AddMessage("You must defeat all enemies before exiting! " +
				strconv.Itoa(len(CurrentGameState.Enemies)) + " enemies remaining")
		}
	}
}

// Функция для проверки близости к выходу
func (gl *GameLogic) checkExitProximity() {
	playerPos := CurrentGameState.Player.Position
	exitRoom := CurrentGameState.Level.ExitRoom

	if exitRoom == nil {
		return
	}

	exitX := exitRoom.X + exitRoom.Width/2
	exitY := exitRoom.Y + exitRoom.Height/2

	// Проверяем, виден ли выход и близко ли к нему игрок
	if CurrentGameState.Level.Visible[exitX][exitY] {
		distance := abs(playerPos.X-exitX) + abs(playerPos.Y-exitY)

		if distance <= 3 {
			gl.AddMessage("You see the exit '>' nearby!")
		}
	}
}

func (gl *GameLogic) UseItem(itemType ItemType, index int) {
	// Проверяем, не усыплен ли игрок
	if CurrentGameState.SleepTurns > 0 {
		CurrentGameState.SleepTurns--
		gl.AddMessage("You are asleep, can't use items")
		return
	}

	// Проверяем, что бэкпек существует
	if CurrentGameState.Backpack == nil {
		gl.AddMessage("No backpack available")
		return
	}

	switch itemType {
	case Food:
		if index < len(CurrentGameState.Backpack.Food) {
			food := CurrentGameState.Backpack.Food[index]
			CurrentGameState.Player.Health += food.HealthBoost
			if CurrentGameState.Player.Health > CurrentGameState.Player.MaxHealth {
				CurrentGameState.Player.Health = CurrentGameState.Player.MaxHealth
			}
			CurrentGameState.Backpack.Food = append(
				CurrentGameState.Backpack.Food[:index],
				CurrentGameState.Backpack.Food[index+1:]...,
			)
			CurrentGameState.GameSession.FoodConsumed++
		}

		if index < 0 || index >= len(CurrentGameState.Backpack.Food) {
			gl.AddMessage("Invalid food selection")
			return
		}
		food := CurrentGameState.Backpack.Food[index]
		CurrentGameState.Player.Health += food.HealthBoost
		if CurrentGameState.Player.Health > CurrentGameState.Player.MaxHealth {
			CurrentGameState.Player.Health = CurrentGameState.Player.MaxHealth
		}
		CurrentGameState.Backpack.Food = append(
			CurrentGameState.Backpack.Food[:index],
			CurrentGameState.Backpack.Food[index+1:]...,
		)
		CurrentGameState.GameSession.FoodConsumed++
		gl.AddMessage("Ate food: +" + strconv.Itoa(food.HealthBoost) + " HP")

	case Potion:
		if index < len(CurrentGameState.Backpack.Potions) {
			potion := CurrentGameState.Backpack.Potions[index]
			gl.applyPotionEffect(potion)
			CurrentGameState.Backpack.Potions = append(
				CurrentGameState.Backpack.Potions[:index],
				CurrentGameState.Backpack.Potions[index+1:]...,
			)
			CurrentGameState.GameSession.PotionsUsed++
		}

		if index < 0 || index >= len(CurrentGameState.Backpack.Potions) {
			gl.AddMessage("Invalid potion selection")
			return
		}
		potion := CurrentGameState.Backpack.Potions[index]
		gl.applyPotionEffect(potion)
		CurrentGameState.Backpack.Potions = append(
			CurrentGameState.Backpack.Potions[:index],
			CurrentGameState.Backpack.Potions[index+1:]...,
		)
		CurrentGameState.GameSession.PotionsUsed++
		gl.AddMessage("Used potion")

	case Scroll:
		if index < 0 || index >= len(CurrentGameState.Backpack.Scrolls) {
			gl.AddMessage("Invalid scroll selection")
			return
		}
		scroll := CurrentGameState.Backpack.Scrolls[index]

		// Проверяем, является ли свиток ключом
		if scroll.SubType == KeyRed || scroll.SubType == KeyBlue ||
			scroll.SubType == KeyGreen || scroll.SubType == KeyYellow {
			gl.UseKey(scroll) // Используем ключ для открытия дверей
			// НЕ удаляем ключ из инвентаря здесь - это сделает UseKey
		} else {
			gl.applyScrollEffect(scroll) // Применяем эффект обычного свитка
			CurrentGameState.Backpack.Scrolls = append(
				CurrentGameState.Backpack.Scrolls[:index],
				CurrentGameState.Backpack.Scrolls[index+1:]...,
			)
			CurrentGameState.GameSession.ScrollsUsed++
			gl.AddMessage("Used scroll")
		}


	case Weapon:
		if index == 0 {
			// Unequip current weapon - бросаем на пол
			if CurrentGameState.Player.CurrentWeapon != nil {
				weapon := CurrentGameState.Player.CurrentWeapon

				// Ищем свободную соседнюю клетку
				playerPos := CurrentGameState.Player.Position
				directions := []Position{
					{1, 0}, {-1, 0}, {0, 1}, {0, -1},
					{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
				}

				for _, dir := range directions {
					newX := playerPos.X + dir.X
					newY := playerPos.Y + dir.Y

					if gl.canMoveTo(newX, newY) {
						// Кладем оружие на пол
						weapon.Position = Position{X: newX, Y: newY}
						CurrentGameState.Items = append(CurrentGameState.Items, weapon)
						gl.AddMessage("Weapon dropped on the floor")
						break
					}
				}

				CurrentGameState.Player.CurrentWeapon = nil
			} else {
				gl.AddMessage("No weapon equipped")
			}
		} else if index-1 < len(CurrentGameState.Backpack.Weapons) && index-1 >= 0 {
			newWeapon := CurrentGameState.Backpack.Weapons[index-1]

			// If player already has a weapon, put it back in backpack

			// Если у игрока уже есть оружие - бросаем его на пол
			if CurrentGameState.Player.CurrentWeapon != nil {
				currentWeapon := CurrentGameState.Player.CurrentWeapon
				CurrentGameState.Backpack.Weapons = append(
					CurrentGameState.Backpack.Weapons,
					currentWeapon,
				)
				gl.AddMessage("Swapped weapons")
			} else {
				gl.AddMessage("Equipped weapon: +" + strconv.Itoa(newWeapon.StrengthBoost) + " STR")
			}

			// Equip the new weapon

				// Ищем свободную соседнюю клетку
				playerPos := CurrentGameState.Player.Position
				directions := []Position{
					{1, 0}, {-1, 0}, {0, 1}, {0, -1},
					{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
				}

				for _, dir := range directions {
					newX := playerPos.X + dir.X
					newY := playerPos.Y + dir.Y

					if gl.canMoveTo(newX, newY) {
						// Кладем оружие на пол
						currentWeapon.Position = Position{X: newX, Y: newY}
						CurrentGameState.Items = append(CurrentGameState.Items, currentWeapon)
						gl.AddMessage("Old weapon dropped on the floor")
						break
					}
				}
			}

			// Экипируем новое оружие
			CurrentGameState.Player.CurrentWeapon = newWeapon

			// Remove the equipped weapon from backpack

			// Удаляем из рюкзака
			CurrentGameState.Backpack.Weapons = append(
				CurrentGameState.Backpack.Weapons[:index-1],
				CurrentGameState.Backpack.Weapons[index:]...,
			)

			gl.AddMessage("Equipped weapon: +" + strconv.Itoa(newWeapon.StrengthBoost) + " STR")
		} else {
			gl.AddMessage("Invalid weapon selection")
		}
	}
}

func (gl *GameLogic) AddMessage(message string) {
	// Основные сообщения (для нижней панели)
	if CurrentGameState.Messages == nil {
		CurrentGameState.Messages = make([]string, 0)
	}

	if len(CurrentGameState.Messages) >= 3 {
		CurrentGameState.Messages = CurrentGameState.Messages[1:]
	}
	CurrentGameState.Messages = append(CurrentGameState.Messages, message)

	// Сообщения о предметах - в статус (правую панель)
	if isItemMessage(message) {
		if CurrentGameState.StatusMessages == nil {
			CurrentGameState.StatusMessages = make([]string, 0)
		}

		if len(CurrentGameState.StatusMessages) >= 5 {
			CurrentGameState.StatusMessages = CurrentGameState.StatusMessages[1:]
		}
		CurrentGameState.StatusMessages = append(CurrentGameState.StatusMessages, message)
	}
	// Основные сообщения (для нижней панели)
	if CurrentGameState.Messages == nil {
		CurrentGameState.Messages = make([]string, 0)
	}

	if len(CurrentGameState.Messages) >= 3 {
		CurrentGameState.Messages = CurrentGameState.Messages[1:]
	}
	CurrentGameState.Messages = append(CurrentGameState.Messages, message)

	// Сообщения о предметах - в статус (правую панель)
	if isItemMessage(message) {
		if CurrentGameState.StatusMessages == nil {
			CurrentGameState.StatusMessages = make([]string, 0)
		}

		if len(CurrentGameState.StatusMessages) >= 5 {
			CurrentGameState.StatusMessages = CurrentGameState.StatusMessages[1:]
		}
		CurrentGameState.StatusMessages = append(CurrentGameState.StatusMessages, message)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func (gl *GameLogic) PickUpItem() {
	// Проверяем, не усыплен ли игрок
	if CurrentGameState.SleepTurns > 0 {
		CurrentGameState.SleepTurns--
		gl.AddMessage("You are asleep, can't pick up items")
		return
	}

	playerPos := CurrentGameState.Player.Position
	foundItem := false


	for i := len(CurrentGameState.Items) - 1; i >= 0; i-- {
		item := CurrentGameState.Items[i]
		if item.Position.X == playerPos.X && item.Position.Y == playerPos.Y {
			if gl.canAddToBackpack(item) {
				gl.addToBackpack(item)
				CurrentGameState.Items = append(
					CurrentGameState.Items[:i],
					CurrentGameState.Items[i+1:]...,
				)
				foundItem = true
				break
			} else {
				gl.AddMessage("Backpack full for this item type!")
			}
		}
	}


	if !foundItem {
		gl.AddMessage("No items to pick up here")
	}
}

func (gl *GameLogic) Attack(enemyIndex int) {
	// Проверяем, не усыплен ли игрок
	if CurrentGameState.SleepTurns > 0 {
		CurrentGameState.SleepTurns--
		gl.AddMessage("You are asleep, can't attack")
		return
	}

	if enemyIndex < len(CurrentGameState.Enemies) {
		enemy := CurrentGameState.Enemies[enemyIndex]
		if gl.isInAttackRange(CurrentGameState.Player.Position, enemy.Position) {
			gl.resolveCombat(CurrentGameState.Player, enemy)
			if enemy.Health <= 0 {
				gl.handleEnemyDefeat(enemy)
				CurrentGameState.Enemies = append(
					CurrentGameState.Enemies[:enemyIndex],
					CurrentGameState.Enemies[enemyIndex+1:]...,
				)
			}
		}
	}
}

func (gl *GameLogic) GenerateLevel(levelNum int) {
	// Создаем уровень фиксированного размера как в классическом Rogue
	mapWidth := 80
	mapHeight := 35


	level := &Level{
		Grid:       make([][]Cell, mapWidth),
		Discovered: make([][]Cell, mapWidth),
		Visible:    make([][]bool, mapWidth),
		Rooms:      make([]*Room, 0),
		Corridors:  make([]*Corridor, 0),
	}


	for i := range level.Grid {
		level.Grid[i] = make([]Cell, mapHeight)
		level.Discovered[i] = make([]Cell, mapHeight)
		level.Visible[i] = make([]bool, mapHeight)
		for j := range level.Grid[i] {
			level.Grid[i][j].IsWall = true // Изначально все стены
		}

	}


	// Генерируем комнаты в сетке 3x3
	gl.generateRoomsInGrid(level, levelNum)


	// Соединяем комнаты коридорами
	gl.connectRoomsWithCorridors(level)


	// Заполняем уровень врагами и предметами
	gl.populateLevel(level, levelNum)


	CurrentGameState.Level = level
	CurrentGameState.CurrentLevel = levelNum


	// Размещаем игрока в стартовой комнате
	if level.StartRoom != nil {
		room := level.StartRoom
		CurrentGameState.Player.Position = Position{
			X: room.X + room.Width/2,
			Y: room.Y + room.Height/2,
		}
	}


	gl.updateVisibility()
}

func (gl *GameLogic) generateRoomsInGrid(level *Level, levelNum int) {
	gridSize := 3
	roomCount := 0
	roomPositions := make(map[[2]int]bool) // Для отслеживания позиций комнат

	for gridY := 0; gridY < gridSize; gridY++ {
		for gridX := 0; gridX < gridSize; gridX++ {
			// 70% chance to create a room in this grid cell
			if rand.Intn(100) < 70 {
				roomWidth := 6 + rand.Intn(6)
				roomHeight := 4 + rand.Intn(4)

				// Calculate position within grid cell
				cellWidth := len(level.Grid) / gridSize
				cellHeight := len(level.Grid[0]) / gridSize

				// Ensure we have enough space for the room
				minX := gridX*cellWidth + 2
				minY := gridY*cellHeight + 2
				maxX := (gridX+1)*cellWidth - roomWidth - 2
				maxY := (gridY+1)*cellHeight - roomHeight - 2

				// Check if we have valid coordinates
				if maxX <= minX || maxY <= minY {
					continue
				}
				x := minX + rand.Intn(maxX-minX)
				y := minY + rand.Intn(maxY-minY)

				// Ensure room is within bounds
				if x < 1 {
					x = 1
				}
				if y < 1 {
					y = 1
				}
				if x+roomWidth >= len(level.Grid)-1 {
					roomWidth = len(level.Grid) - x - 2
				}
				if y+roomHeight >= len(level.Grid[0])-1 {
					roomHeight = len(level.Grid[0]) - y - 2
				}

				// Ensure minimum room size
				if roomWidth < 3 || roomHeight < 3 {
					continue
				}

				// Проверяем, не будет ли эта комната создавать двойные коридоры
				// Проверяем соседние клетки сетки на наличие комнат
				hasAdjacentRoom := false
				for _, offset := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
					adjGridX := gridX + offset[0]
					adjGridY := gridY + offset[1]

					if adjGridX >= 0 && adjGridX < gridSize &&
						adjGridY >= 0 && adjGridY < gridSize &&
						roomPositions[[2]int{adjGridX, adjGridY}] {
						hasAdjacentRoom = true
						break
					}
				}

				// Если есть соседняя комната, уменьшаем шанс создания еще одной
				if hasAdjacentRoom && rand.Intn(100) < 50 {
					continue
				}

				room := &Room{
					X:       x,
					Y:       y,
					Width:   roomWidth,
					Height:  roomHeight,
					Enemies: make([]*Enemy, 0),
					Items:   make([]*Item, 0),
				}

				// Create the room
				for rx := x; rx < x+roomWidth; rx++ {
					for ry := y; ry < y+roomHeight; ry++ {
						if rx >= 0 && ry >= 0 && rx < len(level.Grid) && ry < len(level.Grid[0]) {
							level.Grid[rx][ry].IsWall = false
						}
					}
				}

				level.Rooms = append(level.Rooms, room)
				roomPositions[[2]int{gridX, gridY}] = true
				roomCount++
			}
		}
	}

	// Ensure we have at least 3 rooms
	for roomCount < 3 {
		gridX := rand.Intn(gridSize)
		gridY := rand.Intn(gridSize)

		// Пропускаем клетки, где уже есть комнаты
		if roomPositions[[2]int{gridX, gridY}] {
			continue
		}

		roomWidth := 6 + rand.Intn(6)
		roomHeight := 4 + rand.Intn(4)

		cellWidth := len(level.Grid) / gridSize
		cellHeight := len(level.Grid[0]) / gridSize

		// Ensure we have enough space for the room
		minX := gridX*cellWidth + 2
		minY := gridY*cellHeight + 2
		maxX := (gridX+1)*cellWidth - roomWidth - 2
		maxY := (gridY+1)*cellHeight - roomHeight - 2

		// Check if we have valid coordinates
		if maxX <= minX || maxY <= minY {
			continue
		}

		x := minX + rand.Intn(maxX-minX)
		y := minY + rand.Intn(maxY-minY)

		if x < 1 {
			x = 1
		}
		if y < 1 {
			y = 1
		}
		if x+roomWidth >= len(level.Grid)-1 {
			roomWidth = len(level.Grid) - x - 2
		}
		if y+roomHeight >= len(level.Grid[0])-1 {
			roomHeight = len(level.Grid[0]) - y - 2
		}

		// Ensure minimum room size
		if roomWidth < 3 || roomHeight < 3 {
			continue
		}

		room := &Room{
			X:       x,
			Y:       y,
			Width:   roomWidth,
			Height:  roomHeight,
			Enemies: make([]*Enemy, 0),
			Items:   make([]*Item, 0),
		}

		for rx := x; rx < x+roomWidth; rx++ {
			for ry := y; ry < y+roomHeight; ry++ {
				if rx >= 0 && ry >= 0 && rx < len(level.Grid) && ry < len(level.Grid[0]) {
					level.Grid[rx][ry].IsWall = false
				}
			}
		}

		level.Rooms = append(level.Rooms, room)
		roomPositions[[2]int{gridX, gridY}] = true
		roomCount++
	}

	// Set start and exit rooms - выбираем максимально удаленные комнаты
	if len(level.Rooms) > 0 {
		level.StartRoom = level.Rooms[0]

		// Находим самую дальнюю комнату от стартовой для выхода
		maxDistance := 0
		var exitRoom *Room
		startCenterX := level.StartRoom.X + level.StartRoom.Width/2
		startCenterY := level.StartRoom.Y + level.StartRoom.Height/2

		for _, room := range level.Rooms {
			roomCenterX := room.X + room.Width/2
			roomCenterY := room.Y + room.Height/2
			distance := (roomCenterX-startCenterX)*(roomCenterX-startCenterX) +
				(roomCenterY-startCenterY)*(roomCenterY-startCenterY)

			if distance > maxDistance {
				maxDistance = distance
				exitRoom = room
	roomsCreated := 0

	// Гарантированно создаем минимум 3 комнат
	for roomsCreated < 3 {
		for gridY := 0; gridY < gridSize && roomsCreated < 3; gridY++ {
			for gridX := 0; gridX < gridSize && roomsCreated < 3; gridX++ {
				// 70% chance to create a room in this grid cell
				if roomsCreated < 3 || rand.Intn(100) < 70 {
					if gl.createRoomInGridCell(level, gridX, gridY, gridSize) { // Убираем levelNum
						roomsCreated++
					}
				}
			}
		}
	}

	// Устанавливаем стартовую и выходную комнаты
	if len(level.Rooms) > 0 {
		level.StartRoom = level.Rooms[0]

		// Находим самую дальнюю комнату от стартовой для выхода
		maxDistance := 0
		var exitRoom *Room
		startCenterX := level.StartRoom.X + level.StartRoom.Width/2
		startCenterY := level.StartRoom.Y + level.StartRoom.Height/2

		for _, room := range level.Rooms {
			roomCenterX := room.X + room.Width/2
			roomCenterY := room.Y + room.Height/2
			distance := (roomCenterX-startCenterX)*(roomCenterX-startCenterX) +
				(roomCenterY-startCenterY)*(roomCenterY-startCenterY)

			if distance > maxDistance {
				maxDistance = distance
				exitRoom = room
			}
		}

		level.ExitRoom = exitRoom
		level.StartRoom.IsStart = true
		level.ExitRoom.IsExit = true
	}
}

func (gl *GameLogic) createRoomInGridCell(level *Level, gridX, gridY, gridSize int) bool {
	roomWidth := 6 + rand.Intn(6)
	roomHeight := 4 + rand.Intn(4)

	cellWidth := len(level.Grid) / gridSize
	cellHeight := len(level.Grid[0]) / gridSize

	minX := gridX*cellWidth + 2
	minY := gridY*cellHeight + 2
	maxX := (gridX+1)*cellWidth - roomWidth - 2
	maxY := (gridY+1)*cellHeight - roomHeight - 2

	if maxX <= minX || maxY <= minY {
		return false
	}

	x := minX + rand.Intn(maxX-minX)
	y := minY + rand.Intn(maxY-minY)

	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}
	if x+roomWidth >= len(level.Grid)-1 {
		roomWidth = len(level.Grid) - x - 2
	}
	if y+roomHeight >= len(level.Grid[0])-1 {
		roomHeight = len(level.Grid[0]) - y - 2
	}

	// Проверяем, не пересекается ли комната с существующими
	for _, existingRoom := range level.Rooms {
		if x < existingRoom.X+existingRoom.Width+2 &&
			x+roomWidth+2 > existingRoom.X &&
			y < existingRoom.Y+existingRoom.Height+2 &&
			y+roomHeight+2 > existingRoom.Y {
			return false
		}
	}

	// Создаем комнату
	room := &Room{
		X:       x,
		Y:       y,
		Width:   roomWidth,
		Height:  roomHeight,
		GridX:   gridX,
		GridY:   gridY,
		Enemies: make([]*Enemy, 0),
		Items:   make([]*Item, 0),
	}

	// Заполняем комнату полом
	for i := x; i < x+roomWidth; i++ {
		for j := y; j < y+roomHeight; j++ {
			if i >= 0 && i < len(level.Grid) && j >= 0 && j < len(level.Grid[0]) {
				level.Grid[i][j].IsWall = false // может также как и полы залить?
			}
		}
	}

	level.Rooms = append(level.Rooms, room)
	return true
}

func (gl *GameLogic) connectRoomsWithCorridors(level *Level) {
	if len(level.Rooms) <= 1 {
		return
	}

	var edges []Edge
	roomMap := make(map[string]int)

	for i, room := range level.Rooms {
		key := fmt.Sprintf("%d,%d", room.GridX, room.GridY)
		roomMap[key] = i
	}

	for i, room := range level.Rooms {
		rightKey := fmt.Sprintf("%d,%d", room.GridX+1, room.GridY)
		if j, exists := roomMap[rightKey]; exists {
			weight := gl.distanceBetweenRooms(room, level.Rooms[j])
			edges = append(edges, Edge{U: i, V: j, Weight: weight})
		}
		downKey := fmt.Sprintf("%d,%d", room.GridX, room.GridY+1)
		if j, exists := roomMap[downKey]; exists {
			weight := gl.distanceBetweenRooms(room, level.Rooms[j])
			edges = append(edges, Edge{U: i, V: j, Weight: weight})
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Weight < edges[j].Weight
	})

	// из примера DSU (система непересекающихся множеств)
	parent := make([]int, len(level.Rooms))
	for i := range parent {
		parent[i] = i
	}

	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}

	union := func(x, y int) {
		parent[find(x)] = find(y)
	}
	// Минимальное остовеое дерево
	var mstEdges []Edge
	for _, edge := range edges {
		if find(edge.U) != find(edge.V) {
			union(edge.U, edge.V)
			mstEdges = append(mstEdges, edge)
			gl.connectTwoRooms(level, level.Rooms[edge.U], level.Rooms[edge.V])
		}
	}

	for {
		comps := make(map[int][]int)
		for i := range level.Rooms {
			root := find(i)
			comps[root] = append(comps[root], i)
		}
		if len(comps) <= 1 {
			break
		}

		var bestEdge *Edge
		bestDist := math.MaxFloat64
		var roots []int
		for k := range comps {
			roots = append(roots, k)
		}

		for _, a := range comps[roots[0]] {
			for _, b := range comps[roots[1]] {
				dist := gl.distanceBetweenRooms(level.Rooms[a], level.Rooms[b])
				if dist < bestDist {
					bestDist = dist
					bestEdge = &Edge{U: a, V: b, Weight: dist}
				}
			}
		}

		if bestEdge != nil {
			union(bestEdge.U, bestEdge.V)
			mstEdges = append(mstEdges, *bestEdge)
			gl.connectTwoRooms(level, level.Rooms[bestEdge.U], level.Rooms[bestEdge.V])
		}
	}

	extraConnections := int(math.Sqrt(float64(len(level.Rooms))))
	for i := 0; i < extraConnections && i < len(edges); i++ {
		randomEdge := edges[rand.Intn(len(edges))]
		if find(randomEdge.U) != find(randomEdge.V) {
			union(randomEdge.U, randomEdge.V)
			gl.connectTwoRooms(level, level.Rooms[randomEdge.U], level.Rooms[randomEdge.V])
		}
	}
}

// для расчета расстояния между комнами
func (gl *GameLogic) distanceBetweenRooms(room1, room2 *Room) float64 {
	x1 := room1.X + room1.Width/2
	y1 := room1.Y + room1.Height/2
	x2 := room2.X + room2.Width/2
	y2 := room2.Y + room2.Height/2

	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}

// для соединения  комнат
func (gl *GameLogic) connectTwoRooms(level *Level, room1, room2 *Room) {

	x1 := room1.X + room1.Width/2
	y1 := room1.Y + room1.Height/2
	x2 := room2.X + room2.Width/2
	y2 := room2.Y + room2.Height/2

	// L образный коридор
	if rand.Intn(2) == 0 {

		gl.createHorizontalCorridor(level, x1, y1, x2, y1)
		gl.createVerticalCorridor(level, x2, y1, x2, y2)
	} else {

		gl.createVerticalCorridor(level, x1, y1, x1, y2)
		gl.createHorizontalCorridor(level, x1, y2, x2, y2)
	}
}

func (gl *GameLogic) createHorizontalCorridor(level *Level, x1, y1, x2, y2 int) {
	start, end := x1, x2
	if start > end {
		start, end = end, start
	}

	for x := start; x <= end; x++ {
		if x >= 0 && x < len(level.Grid) && y1 >= 0 && y1 < len(level.Grid[0]) {
			level.Grid[x][y1].IsWall = false
			level.Grid[x][y1].IsCorridor = true
		}
	}
}

func (gl *GameLogic) createVerticalCorridor(level *Level, x1, y1, x2, y2 int) {
	start, end := y1, y2
	if start > end {
		start, end = end, start
	}

	for y := start; y <= end; y++ {
		if x1 >= 0 && x1 < len(level.Grid) && y >= 0 && y < len(level.Grid[0]) {
			level.Grid[x1][y].IsWall = false
			level.Grid[x1][y].IsCorridor = true
		}
	}
}

func (gl *GameLogic) populateLevel(level *Level, levelNum int) {
	// Генерируем двери и ключи (Задание 6)
	if levelNum > 1 {
		gl.generateDoorsAndKeys(level, levelNum)
	}

	// Применяем балансировку сложности (Задание 7)
	difficulty := 3 // Средняя сложность по умолчанию
	if CurrentGameState.PlayerProgress != nil {
		difficulty = CurrentGameState.PlayerProgress.DifficultyLevel
	}
	adjustedLevelNum := levelNum + (difficulty - 3)

	for _, room := range level.Rooms {
		if room == level.StartRoom {
			continue
		}

		// Корректируем количество врагов на основе сложности
		baseEnemies := 1 + adjustedLevelNum/2
		numEnemies := baseEnemies + (difficulty - 3)
		if numEnemies > 5 {
			numEnemies = 5
		}
		if numEnemies < 1 {
			numEnemies = 1
		}

		for i := 0; i < numEnemies; i++ {
			if rand.Intn(100) < 70 {
				enemy := gl.createEnemy(adjustedLevelNum, room)
				if enemy != nil {
					room.Enemies = append(room.Enemies, enemy)
					CurrentGameState.Enemies = append(CurrentGameState.Enemies, enemy)
				}
			}
		}

		level.ExitRoom = exitRoom
		level.StartRoom.IsStart = true
		level.ExitRoom.IsExit = true

		// Mark exit position
		exitX := level.ExitRoom.X + level.ExitRoom.Width/2
		exitY := level.ExitRoom.Y + level.ExitRoom.Height/2
		if exitX < len(level.Grid) && exitY < len(level.Grid[0]) {
			level.Grid[exitX][exitY].IsWall = false
		}
	}
}

func (gl *GameLogic) connectRoomsWithCorridors(level *Level) {
	// Connect rooms in the order they were created (minimum spanning tree approach)
	for i := 0; i < len(level.Rooms)-1; i++ {
		room1 := level.Rooms[i]
		room2 := level.Rooms[i+1]

		// Get center points of rooms
		center1X := room1.X + room1.Width/2
		center1Y := room1.Y + room1.Height/2
		center2X := room2.X + room2.Width/2
		center2Y := room2.Y + room2.Height/2

		// Create L-shaped corridor with proper connection
		gl.createLShapedCorridor(level, center1X, center1Y, center2X, center2Y)

		// Add corridor to level structure
		corridor := &Corridor{
			Start: Position{X: center1X, Y: center1Y},
			End:   Position{X: center2X, Y: center2Y},
			Width: 1,
		}
		level.Corridors = append(level.Corridors, corridor)
	}

	// Add some random connections for better maze connectivity
	// but ensure we don't create too many double corridors
	if len(level.Rooms) > 3 {
		// Connect random rooms (about 30% chance for additional connections)
		for i := 0; i < len(level.Rooms); i++ {
			for j := i + 2; j < len(level.Rooms); j++ {
				if rand.Intn(100) < 30 { // 30% chance for additional connection
					room1 := level.Rooms[i]
					room2 := level.Rooms[j]

					center1X := room1.X + room1.Width/2
					center1Y := room1.Y + room1.Height/2
					center2X := room2.X + room2.Width/2
					center2Y := room2.Y + room2.Height/2

					gl.createLShapedCorridor(level, center1X, center1Y, center2X, center2Y)

					corridor := &Corridor{
						Start: Position{X: center1X, Y: center1Y},
						End:   Position{X: center2X, Y: center2Y},
						Width: 1,
					}
					level.Corridors = append(level.Corridors, corridor)
				}
			}
		}
	}

		// Корректируем количество предметов на основе сложности
		baseItems := 1 + adjustedLevelNum/3
		numItems := baseItems - (difficulty - 3)
		if numItems > 3 {
			numItems = 3
		}
		if numItems < 0 {
			numItems = 0
		}

		for i := 0; i < numItems; i++ {
			if rand.Intn(100) < 60 {
				item := gl.createItem(adjustedLevelNum, room, level)
				if item != nil {
					room.Items = append(room.Items, item)
					CurrentGameState.Items = append(CurrentGameState.Items, item)
				}
			}
		}
	}
}

// В функции createEnemy исправим ошибку с enemy
func (gl *GameLogic) createEnemy(levelNum int, room *Room) *Enemy {
	// Сначала определим типы и веса
	enemyTypes := []EnemyType{Zombie, Vampire, Ghost, Ogre, SnakeMage, Mimic}
	weights := []int{35, 15, 15, 15, 10, 10} // Уменьшили веса других, добавили Mimic

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	roll := rand.Intn(totalWeight)
	var chosenType EnemyType

	for i, w := range weights {
		if roll < w {
			chosenType = enemyTypes[i]
			break
		}
		roll -= w
	}

	health := 4 + levelNum*2  // Меньше здоровья
	agility := 3 + levelNum   // Меньше ловкости
	strength := 2 + levelNum  // Меньше силы
	hostility := 5 + levelNum // Меньший радиус агрессии

	switch chosenType {
	case Zombie:
		health += 6
		agility -= 1
		strength += 1
		hostility += 2
	case Vampire:
		health += 3
		agility += 2
		strength += 1
		hostility += 3
	case Ghost:
		health += 1
		agility += 3
		strength -= 1
		hostility -= 2
	case Ogre:
		health += 10
		agility -= 2
		strength += 3
		hostility += 1
	case SnakeMage:
		health += 2
		agility += 2
		strength -= 1
		hostility += 4
	case Mimic:
		health += 8
		agility += 3
		strength += 1
		hostility = 1 // Низкая враждебность
	}

	// Устанавливаем паттерн движения в зависимости от типа
	var pattern MovementPattern
	switch chosenType {
	case Zombie:
		pattern = Random
	case Vampire:
		pattern = Chase
	case Ghost:
		pattern = Teleport
	case Ogre:
		pattern = TwoStep
	case SnakeMage:
		pattern = Diagonal
	case Mimic:
		pattern = Random
	default:
		pattern = Random
	}

	var pos Position
	attempts := 0
	for attempts < 50 {
		pos = Position{
			X: room.X + 1 + rand.Intn(max(1, room.Width-2)),
			Y: room.Y + 1 + rand.Intn(max(1, room.Height-2)),
		}
		if pos.X >= room.X && pos.X < room.X+room.Width &&
			pos.Y >= room.Y && pos.Y < room.Y+room.Height {
			break
		}
		attempts++
	}

	if attempts == 50 {
		return nil
	}

	return &Enemy{
		Type:                 chosenType,
		Health:               health,
		Agility:              agility,
		Strength:             strength,
		Hostility:            hostility,
		Position:             pos,
		IsVisible:            true,
		Pattern:              pattern,
		moveCooldown:         0,
		hasBeenHit:           false,
		attackCooldown:       0,
		teleportCooldown:     0,
		invisibilityCooldown: 0,
		IsDisguised:          chosenType == Mimic, // Mimic начинается замаскированным
	}
}

func (gl *GameLogic) createLShapedCorridor(level *Level, startX, startY, endX, endY int) {
	// Choose turning point - either go horizontal first then vertical,
	// or vertical first then horizontal
	if rand.Intn(2) == 0 {
		// Horizontal then vertical
		turningX := endX

		gl.createHorizontalCorridor(level, startX, turningX, startY)
		gl.createVerticalCorridor(level, startY, endY, turningX)
	} else {
		// Vertical then horizontal
		turningY := endY

		gl.createVerticalCorridor(level, startY, turningY, startX)
		gl.createHorizontalCorridor(level, startX, endX, turningY)
	}
}

func (gl *GameLogic) createHorizontalCorridor(level *Level, startX, endX, y int) {
	minX, maxX := min(startX, endX), max(startX, endX)
	for x := minX; x <= maxX; x++ {
		if x >= 0 && y >= 0 && x < len(level.Grid) && y < len(level.Grid[0]) {
			// Mark as corridor, but don't overwrite rooms
			if level.Grid[x][y].IsWall {
				level.Grid[x][y].IsWall = false
				level.Grid[x][y].IsCorridor = true
			}
		}
	}
}

func (gl *GameLogic) createVerticalCorridor(level *Level, startY, endY, x int) {
	minY, maxY := min(startY, endY), max(startY, endY)
	for y := minY; y <= maxY; y++ {
		if x >= 0 && y >= 0 && x < len(level.Grid) && y < len(level.Grid[0]) {
			// Mark as corridor, but don't overwrite rooms
			if level.Grid[x][y].IsWall {
				level.Grid[x][y].IsWall = false
				level.Grid[x][y].IsCorridor = true
			}
		}
	}
func (gl *GameLogic) createItem(levelNum int, room *Room, levelOpt ...*Level) *Item {
	var level *Level
	if len(levelOpt) > 0 && levelOpt[0] != nil {
		level = levelOpt[0]
	} else {
		level = CurrentGameState.Level
	}

	if level == nil {
		return nil
	}

	itemTypes := []ItemType{Treasure, Food, Potion, Scroll, Weapon}
	weights := []int{30, 25, 15, 15, 15}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	roll := rand.Intn(totalWeight)
	var chosenType ItemType
	var chosenSubType ItemSubType

	for i, w := range weights {
		if roll < w {
			chosenType = itemTypes[i]
			break
		}
		roll -= w
	}

	switch chosenType {
	case Treasure, Food:
		chosenSubType = 0
	case Potion:
		potionTypes := []ItemSubType{HealthPotion, AgilityPotion, StrengthPotion}
		chosenSubType = potionTypes[rand.Intn(len(potionTypes))]
	case Scroll:
		scrollTypes := []ItemSubType{MaxHealthScroll, AgilityScroll, StrengthScroll}
		chosenSubType = scrollTypes[rand.Intn(len(scrollTypes))]
	case Weapon:
		weaponTypes := []ItemSubType{Sword, Axe}
		chosenSubType = weaponTypes[rand.Intn(len(weaponTypes))]
	}

	healthBoost := 5 + levelNum*2
	maxHealthBoost := 3 + levelNum
	agilityBoost := 1 + levelNum
	strengthBoost := 2 + levelNum
	value := 10 + levelNum*5

	var pos Position
	attempts := 0
	found := false

	for attempts < 50 {
		if room != nil {
			pos = Position{
				X: room.X + 1 + rand.Intn(max(1, room.Width-2)),
				Y: room.Y + 1 + rand.Intn(max(1, room.Height-2)),
			}
		} else {
			pos = Position{
				X: rand.Intn(len(level.Grid)),
				Y: rand.Intn(len(level.Grid[0])),
			}
		}

		if pos.X >= 0 && pos.X < len(level.Grid) &&
			pos.Y >= 0 && pos.Y < len(level.Grid[0]) &&
			!level.Grid[pos.X][pos.Y].IsWall {
			found = true
			break
		}
		attempts++
	}

	if !found {
		return nil
	}

	return &Item{
		Type:           chosenType,
		SubType:        chosenSubType,
		HealthBoost:    healthBoost,
		MaxHealthBoost: maxHealthBoost,
		AgilityBoost:   agilityBoost,
		StrengthBoost:  strengthBoost,
		Value:          value,
		Position:       pos,
	}
}

func (gl *GameLogic) SaveGame(filename string) error {
	data, err := json.Marshal(CurrentGameState)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (gl *GameLogic) LoadGame(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &CurrentGameState)
}

func (gl *GameLogic) UpdateGameState() {
	if CurrentGameState.IsPlayerTurn {
		return
	}
	// Обновляем видимость
	gl.updateVisibility()

	// Двигаем врагов КАЖДЫЙ ход
	gl.moveEnemies()

	// Проверяем автоматический выход
	gl.checkForAutomaticExit()

	// Проверяем близость к выходу
	gl.checkExitProximity()

	// Балансируем сложность (Задание 7)
	gl.adjustDifficulty()

	// Проверяем условия игры
	gl.CheckGameConditions()

	// Обновляем камеру
	gl.updateCamera()

	// Обновляем эффекты
	gl.updateEffects()
	CurrentGameState.IsPlayerTurn = true
}

func (gl *GameLogic) updateEffects() {
	// Уменьшаем длительность эффектов
	for i := len(CurrentGameState.Player.ActiveEffects) - 1; i >= 0; i-- {
		effect := CurrentGameState.Player.ActiveEffects[i]
		effect.Duration--

		if effect.Duration <= 0 {
			// Удаляем эффект
			CurrentGameState.Player.ActiveEffects = append(
				CurrentGameState.Player.ActiveEffects[:i],
				CurrentGameState.Player.ActiveEffects[i+1:]...,
			)
		}
	}
}

func (gl *GameLogic) CheckGameConditions() {
	if CurrentGameState.Player.Health <= 0 {
		CurrentGameState.GameOver = true
	}

	// Проверка победы (нахождение на выходе)
	playerPos := CurrentGameState.Player.Position
	if CurrentGameState.Level.ExitRoom != nil {
		exit := CurrentGameState.Level.ExitRoom
		exitCenterX := exit.X + exit.Width/2
		exitCenterY := exit.Y + exit.Height/2

		if playerPos.X == exitCenterX && playerPos.Y == exitCenterY {
			if CurrentGameState.CurrentLevel == 21 {
				CurrentGameState.GameWon = true
			} else {
				gl.GenerateLevel(CurrentGameState.CurrentLevel + 1)
			}
		}
	}
}

func (gl *GameLogic) updateItemEffects() {
	for i := len(CurrentGameState.Player.ActiveEffects) - 1; i >= 0; i-- {
		effect := CurrentGameState.Player.ActiveEffects[i]
		effect.Duration--

		if effect.Duration <= 0 {
			gl.removeEffect(effect)
			CurrentGameState.Player.ActiveEffects = append(
				CurrentGameState.Player.ActiveEffects[:i],
				CurrentGameState.Player.ActiveEffects[i+1:]...,
			)
	// Проверяем смерть игрока
	if CurrentGameState.Player.Health <= 0 {
		CurrentGameState.GameOver = true
		gl.AddMessage("You have died! Game Over.")
		return
	}
}

func (gl *GameLogic) canMoveTo(x, y int) bool {
	if x < 0 || x >= len(CurrentGameState.Level.Grid) ||
		y < 0 || y >= len(CurrentGameState.Level.Grid[0]) {
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

	// Проверяем, нет ли врага на этой клетке
	for _, enemy := range CurrentGameState.Enemies {
		if enemy.Position.X == x && enemy.Position.Y == y {
			return false // Не можем пройти через врага
		}
	}

	// ПРЕДМЕТЫ НЕ БЛОКИРУЮТ ДВИЖЕНИЕ - они автоматически подбираются

	return true
}

func (gl *GameLogic) isInAttackRange(pos1, pos2 Position) bool {
	dx := abs(pos1.X - pos2.X)
	dy := abs(pos1.Y - pos2.Y)
	return dx <= 1 && dy <= 1
}

func (gl *GameLogic) resolveCombat(attacker interface{}, defender interface{}) {
	var attackerStr, attackerAgi int

	// Получаем характеристики атакующего
	switch a := attacker.(type) {
	case *Character:
		attackerStr = a.Strength
		if a.CurrentWeapon != nil {
			attackerStr += a.CurrentWeapon.StrengthBoost
		}
		attackerAgi = a.Agility
	case *Enemy:
		attackerStr = a.Strength
		attackerAgi = a.Agility
	}

	// Получаем характеристики защищающегося
	var defenderAgi int
	switch d := defender.(type) {
	case *Character:
		defenderAgi = d.Agility
	case *Enemy:
		defenderAgi = d.Agility
	}

	// Рассчитываем шанс попадания (более сбалансированный)
	hitChance := 50 + (attackerAgi - defenderAgi)
	if hitChance < 20 {
		hitChance = 20
	}
	if hitChance > 90 {
		hitChance = 90
	}

	if rand.Intn(100) < hitChance {
		damage := attackerStr

		// Настраиваем урон в зависимости от того, кто атакует
		switch d := defender.(type) {
		case *Character:
			// Враги наносят меньше урона игроку
			if _, ok := attacker.(*Enemy); ok {
				damage = max(1, attackerStr/3) // Враги бьют значительно слабее
			}
			d.Health -= damage
			CurrentGameState.GameSession.HitsTaken++
			gl.AddMessage("You take " + strconv.Itoa(damage) + " damage!")

			// спецэффекты врагов
			if e, ok := attacker.(*Enemy); ok {
				gl.applyEnemySpecialEffects(e, d)
			}

		case *Enemy:
			// Игрок наносит полный урон врагам
			if _, ok := attacker.(*Character); ok {
				damage = attackerStr
			} else {
				damage = max(1, attackerStr/2) // Враги друг другу наносят средний урон
			}
			d.Health -= damage
			CurrentGameState.GameSession.HitsLanded++
			gl.AddMessage("You hit enemy for " + strconv.Itoa(damage) + " damage!")

			// особенность вампира: первый удар = дренаж max HP
			if d.Type == Vampire && !d.hasBeenHit {
				d.hasBeenHit = true
				if char, ok := attacker.(*Character); ok {
					char.MaxHealth = max(10, char.MaxHealth-2) // Меньший дренаж
					if char.Health > char.MaxHealth {
						char.Health = char.MaxHealth
					}
					gl.AddMessage("Vampire drained your max health! -2 Max HP")
				}
			}
		}
	} else {
		gl.AddMessage("Attack missed!")
	}
}

// Специальные эффекты для Mimic
func (gl *GameLogic) applyEnemySpecialEffects(enemy *Enemy, player *Character) {
	switch enemy.Type {
	case SnakeMage:
		if rand.Intn(100) < 30 {
			CurrentGameState.SleepTurns = 2 + rand.Intn(3)
			gl.AddMessage("Snake Mage puts you to sleep for " +
				strconv.Itoa(CurrentGameState.SleepTurns) + " turns!")
		}
	case Mimic:
		if rand.Intn(100) < 25 {
			CurrentGameState.SleepTurns = 1
			gl.AddMessage("Mimic's surprise attack stuns you!")
		}
	}
}

func (gl *GameLogic) removeEffect(effect *Effect) {
	switch effect.Type {
	case AgilityEffect:
		CurrentGameState.Player.Agility -= effect.Value
	case StrengthEffect:
		CurrentGameState.Player.Strength -= effect.Value
	case MaxHealthEffect:
		CurrentGameState.Player.MaxHealth -= effect.Value

		if CurrentGameState.Player.Health > CurrentGameState.Player.MaxHealth {
			CurrentGameState.Player.Health = CurrentGameState.Player.MaxHealth
		}

		if CurrentGameState.Player.Health <= 0 {
			CurrentGameState.Player.Health = 1
func (gl *GameLogic) handleEnemyDefeat(enemy *Enemy) {
	CurrentGameState.Player.DefeatedEnemies++
	CurrentGameState.GameSession.EnemiesDefeated++

	switch enemy.Type {
	case Zombie:
		gl.AddMessage("Zombie defeated! +10 gold")
		CurrentGameState.Backpack.Treasure += 10
	case Vampire:
		gl.AddMessage("Vampire defeated! +25 gold")
		CurrentGameState.Backpack.Treasure += 25
	case Ghost:
		gl.AddMessage("Ghost defeated! +20 gold")
		CurrentGameState.Backpack.Treasure += 20
	case Ogre:
		gl.AddMessage("Ogre defeated! +30 gold")
		CurrentGameState.Backpack.Treasure += 30
	case SnakeMage:
		gl.AddMessage("Snake Mage defeated! +40 gold")
		CurrentGameState.Backpack.Treasure += 40
	}

	if rand.Intn(100) < 20 {
		item := gl.createItem(CurrentGameState.CurrentLevel, nil)
		if item != nil {
			item.Position = enemy.Position
			CurrentGameState.Items = append(CurrentGameState.Items, item)
			gl.AddMessage("Enemy dropped an item!")
		}
	}
}

func (gl *GameLogic) applyPotionEffect(potion *Item) {
	var effect *Effect

	switch potion.SubType {
	case HealthPotion:
		CurrentGameState.Player.Health += potion.HealthBoost
		if CurrentGameState.Player.Health > CurrentGameState.Player.MaxHealth {
			CurrentGameState.Player.Health = CurrentGameState.Player.MaxHealth
		}
		return

	case AgilityPotion:
		CurrentGameState.Player.Agility += potion.AgilityBoost
		effect = &Effect{
			Type:        AgilityEffect,
			Duration:    10 + CurrentGameState.CurrentLevel,
			Value:       potion.AgilityBoost,
			Description: "Agility Potion",
		}

	case StrengthPotion:
		CurrentGameState.Player.Strength += potion.StrengthBoost
		effect = &Effect{
			Type:        StrengthEffect,
			Duration:    10 + CurrentGameState.CurrentLevel,
			Value:       potion.StrengthBoost,
			Description: "Strength Potion",
		}
	}

	if effect != nil {
		CurrentGameState.Player.ActiveEffects = append(
			CurrentGameState.Player.ActiveEffects,
			effect,
		)
var enemyTurnCounter int

// Исправляем функцию moveEnemies
func (gl *GameLogic) moveEnemies() {
	enemyTurnCounter++

	for _, enemy := range CurrentGameState.Enemies {
		if enemy.attackCooldown > 0 {
			enemy.attackCooldown--
			continue
		}
		if enemy.teleportCooldown > 0 {
			enemy.teleportCooldown--
		}
		if enemy.invisibilityCooldown > 0 {
			enemy.invisibilityCooldown--
			if enemy.invisibilityCooldown == 0 {
				enemy.IsVisible = true
				gl.AddMessage("Ghost becomes visible again!")
			}
		}

		playerPos := CurrentGameState.Player.Position
		dist := abs(enemy.Position.X-playerPos.X) + abs(enemy.Position.Y-playerPos.Y)

		// Проверяем, находится ли игрок в радиусе враждебности
		if dist <= enemy.Hostility {
			// Игрок в зоне агрессии - преследуем!
			switch enemy.Type {
			case Zombie:
				gl.moveRandom(enemy)
			case Vampire:
				gl.moveChase(enemy)
			case Ghost:
				gl.moveTeleport(enemy)
			case Ogre:
				gl.moveTwoStep(enemy)
			case SnakeMage:
				gl.moveDiagonal(enemy)
			case Mimic: // ДОБАВЛЯЕМ MIMIC
				gl.moveMimic(enemy)
			default:
				gl.moveRandom(enemy)
			}
		} else {
			// Игрок вне зоны агрессии - патрулируем
			switch enemy.Type {
			case Zombie:
				gl.moveRandom(enemy)
			case Vampire:
				gl.moveRandom(enemy)
			case Ghost:
				gl.moveRandom(enemy) // Призрак телепортируется только при преследовании
			case Ogre:
				gl.moveRandom(enemy)
			case SnakeMage:
				gl.moveRandom(enemy)
			case Mimic: // ДОБАВЛЯЕМ MIMIC
				gl.moveMimic(enemy)
			default:
				gl.moveRandom(enemy)
			}
		}

		// Проверяем атаку после движения
		if gl.isInAttackRange(enemy.Position, CurrentGameState.Player.Position) {
			gl.resolveCombat(enemy, CurrentGameState.Player)
			switch enemy.Type {
			case Ogre:
				enemy.attackCooldown = 1
				gl.AddMessage("Ogre is resting after attack")
			}
		}
	}
}

func (gl *GameLogic) moveRandom(enemy *Enemy) {
	directions := []Position{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0},
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}

	shuffled := make([]Position, len(directions))
	copy(shuffled, directions)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for _, dir := range shuffled {
		newX := enemy.Position.X + dir.X
		newY := enemy.Position.Y + dir.Y

		if gl.canMoveTo(newX, newY) {
			enemy.Position.X = newX
			enemy.Position.Y = newY
			break
		}
	}
}

func (gl *GameLogic) moveChase(enemy *Enemy) {
	playerPos := CurrentGameState.Player.Position
	enemyPos := enemy.Position

	dx := 0
	if enemyPos.X < playerPos.X {
		dx = 1
	}
	if enemyPos.X > playerPos.X {
		dx = -1
	}

	dy := 0
	if enemyPos.Y < playerPos.Y {
		dy = 1
	}
	if enemyPos.Y > playerPos.Y {
		dy = -1
	}

	// 70% chance to move directly toward player
	if rand.Intn(100) < 70 {
		// Try horizontal first, then vertical
		if dx != 0 && gl.canMoveTo(enemyPos.X+dx, enemyPos.Y) {
			enemy.Position.X += dx
		} else if dy != 0 && gl.canMoveTo(enemyPos.X, enemyPos.Y+dy) {
			enemy.Position.Y += dy
		}
	} else {
		// Random movement
		gl.moveRandom(enemy)
	}
}

func (gl *GameLogic) moveTeleport(enemy *Enemy) {
	// Призрак телепортируется и становится невидимым
	if enemy.teleportCooldown == 0 {
		// Телепортируемся рядом с игроком
		playerPos := CurrentGameState.Player.Position
		directions := []Position{
			{1, 0}, {-1, 0}, {0, 1}, {0, -1},
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
		}

		for _, dir := range directions {
			newX := playerPos.X + dir.X
			newY := playerPos.Y + dir.Y

			if gl.canMoveTo(newX, newY) {
				enemy.Position.X = newX
				enemy.Position.Y = newY
				enemy.teleportCooldown = 3 // Кулдаун телепорта

				// Становится невидимым
				enemy.IsVisible = false
				enemy.invisibilityCooldown = 2
				gl.AddMessage("Ghost teleports and becomes invisible!")
				break
			}
		}
	} else {
		// Двигаемся случайно во время кулдауна
		gl.moveRandom(enemy)
	}
}

func (gl *GameLogic) moveDiagonal(enemy *Enemy) {
	// Змей-маг двигается по диагонали
	directions := []Position{
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
	}

	// 60% chance to move diagonally toward player
	if rand.Intn(100) < 60 {
		playerPos := CurrentGameState.Player.Position
		enemyPos := enemy.Position

		dx := 1
		if enemyPos.X > playerPos.X {
			dx = -1
		}
		dy := 1
		if enemyPos.Y > playerPos.Y {
			dy = -1
		}

		if gl.canMoveTo(enemyPos.X+dx, enemyPos.Y+dy) {
			enemy.Position.X += dx
			enemy.Position.Y += dy
			return
		}
	}

	// Random diagonal movement
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		newX := enemy.Position.X + dir.X
		newY := enemy.Position.Y + dir.Y

		if gl.canMoveTo(newX, newY) {
			enemy.Position.X = newX
			enemy.Position.Y = newY
			break
		}
	}
}

func (gl *GameLogic) canMoveTo(x, y int) bool {
	if x < 0 || y < 0 ||
		x >= len(CurrentGameState.Level.Grid) ||
		y >= len(CurrentGameState.Level.Grid[0]) {
		return false
	}

	return !CurrentGameState.Level.Grid[x][y].IsWall
func (gl *GameLogic) moveTwoStep(enemy *Enemy) {
	// Огр двигается на две клетки за ход, но с кулдауном
	if enemy.moveCooldown == 0 {
		// Первый шаг
		gl.moveChase(enemy)
		enemy.moveCooldown = 1
	} else {
		// Второй шаг
		gl.moveChase(enemy)
		enemy.moveCooldown = 0
	}
}

func (gl *GameLogic) updateVisibility() {
	playerPos := CurrentGameState.Player.Position
	viewDistance := 8

	for dx := -viewDistance; dx <= viewDistance; dx++ {
		for dy := -viewDistance; dy <= viewDistance; dy++ {
			x, y := playerPos.X+dx, playerPos.Y+dy
			if x >= 0 && y >= 0 && x < len(CurrentGameState.Level.Visible) && y < len(CurrentGameState.Level.Visible[0]) {
				if dx*dx+dy*dy <= viewDistance*viewDistance {
					if gl.hasLineOfSight(playerPos, Position{X: x, Y: y}) {
						CurrentGameState.Level.Visible[x][y] = true
						CurrentGameState.Level.Discovered[x][y] = true
					}

	// Сбрасываем видимость
	for i := range CurrentGameState.Level.Visible {
		for j := range CurrentGameState.Level.Visible[i] {
			CurrentGameState.Level.Visible[i][j] = false
		}
	}

	// Устанавливаем видимость в радиусе 6 клеток
	radius := 6
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			x := playerPos.X + dx
			y := playerPos.Y + dy

			if x >= 0 && x < len(CurrentGameState.Level.Grid) &&
				y >= 0 && y < len(CurrentGameState.Level.Grid[0]) {

				// Проверяем лучами видимость
				if gl.hasLineOfSight(playerPos, Position{X: x, Y: y}) {
					CurrentGameState.Level.Visible[x][y] = true
					CurrentGameState.Level.Discovered[x][y] = CurrentGameState.Level.Grid[x][y]
				}
			}
		}
	}

	// Обновляем видимость врагов
	for _, enemy := range CurrentGameState.Enemies {
		if enemy.invisibilityCooldown == 0 {
			enemy.IsVisible = CurrentGameState.Level.Visible[enemy.Position.X][enemy.Position.Y]
		}
	}

	// Обновляем видимость врагов
	for _, enemy := range CurrentGameState.Enemies {
		if enemy.invisibilityCooldown == 0 {
			enemy.IsVisible = CurrentGameState.Level.Visible[enemy.Position.X][enemy.Position.Y]
		}
	}
}

func (gl *GameLogic) hasLineOfSight(start, end Position) bool {
	// Простой алгоритм Брезенхема для проверки видимости
	dx := abs(end.X - start.X)
	dy := abs(end.Y - start.Y)

	var sx, sy int
	if start.X < end.X {
		sx = 1
	} else {
		sx = -1
	}
	if start.Y < end.Y {
		sy = 1
	} else {
		sy = -1
	}


	err := dx - dy
	x, y := start.X, start.Y


	x, y := start.X, start.Y
	for {
		if x == end.X && y == end.Y {
			break
		}


		if CurrentGameState.Level.Grid[x][y].IsWall {
			return false
		}


		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	return true
}

func (gl *GameLogic) checkForItems() {
	playerPos := CurrentGameState.Player.Position
	for _, item := range CurrentGameState.Items {
		if item.Position.X == playerPos.X && item.Position.Y == playerPos.Y {
			gl.AddMessage("You see an item here (press 'g' to pick up)")
			break
		}
	}
}

func (gl *GameLogic) checkForEnemies() {
	playerPos := CurrentGameState.Player.Position

	for i := len(CurrentGameState.Enemies) - 1; i >= 0; i-- {
		enemy := CurrentGameState.Enemies[i]
		// Бой только при КОНТАКТЕ (на одной клетке)
		if enemy.Position.X == playerPos.X && enemy.Position.Y == playerPos.Y {
			gl.resolveCombat(CurrentGameState.Player, enemy)
			if enemy.Health <= 0 {
				gl.handleEnemyDefeat(enemy)
				CurrentGameState.Enemies = append(
					CurrentGameState.Enemies[:i],
					CurrentGameState.Enemies[i+1:]...,
				)
			}
		}
	}
	playerPos := CurrentGameState.Player.Position

	for i := len(CurrentGameState.Enemies) - 1; i >= 0; i-- {
		enemy := CurrentGameState.Enemies[i]

		// Проверяем соседние клетки (атака на расстоянии 1 клетки)
		if abs(enemy.Position.X-playerPos.X) <= 1 &&
			abs(enemy.Position.Y-playerPos.Y) <= 1 {

			gl.resolveCombat(CurrentGameState.Player, enemy)

			if enemy.Health <= 0 {
				gl.handleEnemyDefeat(enemy)
				// Удаляем врага из списка
				CurrentGameState.Enemies = append(
					CurrentGameState.Enemies[:i],
					CurrentGameState.Enemies[i+1:]...,
				)
			} else {
				// Также враг атакует игрока в ответ
				gl.resolveCombat(enemy, CurrentGameState.Player)
			}
		}
	}
}

func (gl *GameLogic) checkForExit() {
	playerPos := CurrentGameState.Player.Position
	if CurrentGameState.Level.ExitRoom != nil {
		exit := CurrentGameState.Level.ExitRoom
		exitCenterX := exit.X + exit.Width/2
		exitCenterY := exit.Y + exit.Height/2

		// Проверяем только центральную клетку выхода
		if playerPos.X == exitCenterX && playerPos.Y == exitCenterY {
			if CurrentGameState.CurrentLevel == 21 {
				CurrentGameState.GameWon = true
			} else {
				gl.GenerateLevel(CurrentGameState.CurrentLevel + 1)
				gl.AddMessage("Descending to level " + strconv.Itoa(CurrentGameState.CurrentLevel))
			}
		}
	}
}

func (gl *GameLogic) isInAttackRange(pos1, pos2 Position) bool {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return dx*dx+dy*dy <= 2
	playerPos := CurrentGameState.Player.Position
	exitRoom := CurrentGameState.Level.ExitRoom

	if exitRoom != nil &&
		playerPos.X >= exitRoom.X && playerPos.X < exitRoom.X+exitRoom.Width &&
		playerPos.Y >= exitRoom.Y && playerPos.Y < exitRoom.Y+exitRoom.Height {

		gl.AddMessage("You found the exit! (defeat all enemies to win)")
	}
}

func (gl *GameLogic) canAddToBackpack(item *Item) bool {
	if item == nil {
		return false
	}
	if item.Type == Treasure {
		return true
	}
	// КЛЮЧИ МОЖНО ВСЕГДА ПОДБИРАТЬ
	if item.Type == Scroll && (item.SubType == KeyRed || item.SubType == KeyBlue ||
		item.SubType == KeyGreen || item.SubType == KeyYellow) {
		return true
	}

	bp := CurrentGameState.Backpack
	if bp == nil {
		return true
	}

	switch item.Type {
	case Weapon:
		return len(bp.Weapons) < 5
	case Food:
		return len(bp.Food) < 10
	case Potion:
		return len(bp.Potions) < 5
	case Scroll:
		return len(bp.Scrolls) < 5
	default:
		return false
	}
}

func (gl *GameLogic) resolveCombat(attacker interface{}, defender interface{}) {
	if attacker == nil || defender == nil {
		return
	}

	var aStrength, aAgility, dAgility int

	// Определяем характеристики атакующего
	switch a := attacker.(type) {
	case *Character:
		aStrength = a.Strength
		if a.CurrentWeapon != nil {
			aStrength += a.CurrentWeapon.StrengthBoost
		}
		aAgility = a.Agility
	case *Enemy:
		aStrength = a.Strength
		aAgility = a.Agility
	}

	// Определяем характеристики защищающегося
	switch d := defender.(type) {
	case *Enemy:
		dAgility = d.Agility
	case *Character:
		dAgility = d.Agility
	}

	// Расчет шанса попадания
	hitChance := 70 + (aAgility - dAgility)
	if rand.Intn(100) < hitChance {
		// Расчет урона
		damage := aStrength/2 + rand.Intn(aStrength/2+1)

		// ПРИМЕНЕНИЕ УРОНА!
		switch d := defender.(type) {
		case *Enemy:
			d.Health -= damage // ВРАГ получает урон
			CurrentGameState.GameSession.HitsLanded++
		case *Character:
			d.Health -= damage // ИГРОК получает урон
			CurrentGameState.GameSession.HitsTaken++
		}
	}
func (gl *GameLogic) addToBackpack(item *Item) {
	if item == nil {
		return
	}
	if CurrentGameState.Backpack == nil {
		CurrentGameState.Backpack = &Backpack{
			Weapons:  make([]*Item, 0),
			Food:     make([]*Item, 0),
			Potions:  make([]*Item, 0),
			Scrolls:  make([]*Item, 0),
			Treasure: 0,
		}
	}

	switch item.Type {
	case Weapon:
		// Если у игрока нет оружия - автоматически экипируем
		if CurrentGameState.Player.CurrentWeapon == nil {
			CurrentGameState.Player.CurrentWeapon = item
			gl.AddMessage("Equipped weapon: +" + strconv.Itoa(item.StrengthBoost) + " STR")
		} else {
			// Если уже есть оружие - кладем новое в рюкзак
			CurrentGameState.Backpack.Weapons = append(CurrentGameState.Backpack.Weapons, item)
			gl.AddMessage("Picked up weapon: +" + strconv.Itoa(item.StrengthBoost) + " STR (added to backpack)")
		}

	case Food:
		CurrentGameState.Backpack.Food = append(CurrentGameState.Backpack.Food, item)
		gl.AddMessage("Picked up food: +" + strconv.Itoa(item.HealthBoost) + " HP")

	case Potion:
		CurrentGameState.Backpack.Potions = append(CurrentGameState.Backpack.Potions, item)
		gl.AddMessage("Picked up potion")

	case Scroll:
		// ОСОБАЯ ОБРАБОТКА ДЛЯ КЛЮЧЕЙ
		if item.SubType == KeyRed || item.SubType == KeyBlue ||
			item.SubType == KeyGreen || item.SubType == KeyYellow {
			// Ключи не должны автоматически использоваться!
			CurrentGameState.Backpack.Scrolls = append(CurrentGameState.Backpack.Scrolls, item)
			gl.AddMessage("Picked up " + gl.getKeyName(item.SubType))
		} else {
			CurrentGameState.Backpack.Scrolls = append(CurrentGameState.Backpack.Scrolls, item)
			gl.AddMessage("Picked up scroll")
		}

	case Treasure:
		CurrentGameState.Backpack.Treasure += item.Value
		if CurrentGameState.Player != nil {
			CurrentGameState.Player.Treasures += item.Value
		}
		gl.AddMessage("Picked up treasure: +" + strconv.Itoa(item.Value) + " gold")
	}
}

func (gl *GameLogic) applyPotionEffect(potion *Item) {
	switch potion.SubType {
	case HealthPotion:
		CurrentGameState.Player.Health += potion.HealthBoost
		if CurrentGameState.Player.Health > CurrentGameState.Player.MaxHealth {
			CurrentGameState.Player.Health = CurrentGameState.Player.MaxHealth
		}
		gl.AddMessage("Health potion: +" + strconv.Itoa(potion.HealthBoost) + " HP")
	case AgilityPotion:
		CurrentGameState.Player.Agility += potion.AgilityBoost
		gl.AddMessage("Agility potion: +" + strconv.Itoa(potion.AgilityBoost) + " AGI")
	case StrengthPotion:
		CurrentGameState.Player.Strength += potion.StrengthBoost
		gl.AddMessage("Strength potion: +" + strconv.Itoa(potion.StrengthBoost) + " STR")
	}
}

func (gl *GameLogic) populateLevel(level *Level, levelNum int) {
	for _, room := range level.Rooms {
		if room.IsStart {
			continue
		}

		enemyCount := 1 + levelNum/3 + rand.Intn(2)
		for i := 0; i < enemyCount; i++ {
			enemy := gl.generateEnemy(levelNum)
			enemy.Position = gl.getRandomPositionInRoom(room)
			room.Enemies = append(room.Enemies, enemy)
			CurrentGameState.Enemies = append(CurrentGameState.Enemies, enemy)
		}

		itemCount := 1 + levelNum/4 + rand.Intn(2)
		for i := 0; i < itemCount; i++ {
			item := gl.generateItem(levelNum)
			item.Position = gl.getRandomPositionInRoom(room)
			room.Items = append(room.Items, item)
			CurrentGameState.Items = append(CurrentGameState.Items, item)
		}
	}
func (gl *GameLogic) applyScrollEffect(scroll *Item) {
	effect := &Effect{}

	switch scroll.SubType {
	case MaxHealthScroll:
		effect.Type = MaxHealthEffect
		effect.Duration = 10
		effect.Value = scroll.MaxHealthBoost
		effect.Description = "+" + strconv.Itoa(scroll.MaxHealthBoost) + " Max HP"
		CurrentGameState.Player.MaxHealth += scroll.MaxHealthBoost
		CurrentGameState.Player.Health += scroll.MaxHealthBoost
		gl.AddMessage("Max Health Scroll: +" + strconv.Itoa(scroll.MaxHealthBoost) + " Max HP")

	case AgilityScroll:
		effect.Type = AgilityEffect
		effect.Duration = 5
		effect.Value = scroll.AgilityBoost
		effect.Description = "+" + strconv.Itoa(scroll.AgilityBoost) + " AGI"
		CurrentGameState.Player.Agility += scroll.AgilityBoost
		gl.AddMessage("Agility Scroll: +" + strconv.Itoa(scroll.AgilityBoost) + " AGI")

	case StrengthScroll:
		effect.Type = StrengthEffect
		effect.Duration = 5
		effect.Value = scroll.StrengthBoost
		effect.Description = "+" + strconv.Itoa(scroll.StrengthBoost) + " STR"
		CurrentGameState.Player.Strength += scroll.StrengthBoost
		gl.AddMessage("Strength Scroll: +" + strconv.Itoa(scroll.StrengthBoost) + " STR")
	}

	CurrentGameState.Player.ActiveEffects = append(CurrentGameState.Player.ActiveEffects, effect)
}

func (gl *GameLogic) updateCamera() {
	playerPos := CurrentGameState.Player.Position
	cam := CurrentGameState.Camera

	// Центрируем камеру на игроке
	cam.X = playerPos.X - cam.Width/2
	cam.Y = playerPos.Y - cam.Height/2

	// Не даем камере выйти за границы карты
	if cam.X < 0 {
		cam.X = 0
	}
	if cam.Y < 0 {
		cam.Y = 0
	}
	if cam.X+cam.Width >= len(CurrentGameState.Level.Grid) {
		cam.X = len(CurrentGameState.Level.Grid) - cam.Width
	}
	if cam.Y+cam.Height >= len(CurrentGameState.Level.Grid[0]) {
		cam.Y = len(CurrentGameState.Level.Grid[0]) - cam.Height
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func isItemMessage(message string) bool {
	itemKeywords := []string{"Picked up", "potion", "scroll", "weapon", "food", "treasure", "gold"}
	for _, keyword := range itemKeywords {
		if contains(message, keyword) {
			return true
		}
	}
	return false
}

func (gl *GameLogic) TryExitLevel() {
	if CurrentGameState.SleepTurns > 0 {
		CurrentGameState.SleepTurns--
		gl.AddMessage("You are asleep, can't exit")
		return
	}

	playerPos := CurrentGameState.Player.Position
	exitRoom := CurrentGameState.Level.ExitRoom

	if exitRoom != nil {
		exitX := exitRoom.X + exitRoom.Width/2
		exitY := exitRoom.Y + exitRoom.Height/2

		playerAtExit := playerPos.X == exitX && playerPos.Y == exitY

		if playerAtExit {
			// Проверяем, что все враги побеждены
			if len(CurrentGameState.Enemies) == 0 {
				gl.nextLevel()
			} else {
				gl.AddMessage("You must defeat all enemies before exiting! " +
					strconv.Itoa(len(CurrentGameState.Enemies)) + " enemies remaining")
			}
		} else {
			gl.AddMessage("You are not at the exit! Find the '>' symbol")
		}
	}
}

// Новая функция для перехода на следующий уровень
func (gl *GameLogic) nextLevel() {
	gl.AddMessage("Moving to level " + strconv.Itoa(CurrentGameState.CurrentLevel+1) + "!")

	// Сохраняем статистику игрока
	currentHealth := CurrentGameState.Player.Health
	currentMaxHealth := CurrentGameState.Player.MaxHealth
	currentAgility := CurrentGameState.Player.Agility
	currentStrength := CurrentGameState.Player.Strength
	currentTreasures := CurrentGameState.Player.Treasures
	currentDefeatedEnemies := CurrentGameState.Player.DefeatedEnemies
	currentWeapon := CurrentGameState.Player.CurrentWeapon
	currentBackpack := CurrentGameState.Backpack
	currentEffects := CurrentGameState.Player.ActiveEffects

	// Генерируем следующий уровень
	nextLevelNum := CurrentGameState.CurrentLevel + 1
	gl.GenerateLevel(nextLevelNum)

	// Восстанавливаем статистику игрока
	CurrentGameState.Player.Health = currentHealth
	CurrentGameState.Player.MaxHealth = currentMaxHealth
	CurrentGameState.Player.Agility = currentAgility
	CurrentGameState.Player.Strength = currentStrength
	CurrentGameState.Player.Treasures = currentTreasures
	CurrentGameState.Player.DefeatedEnemies = currentDefeatedEnemies
	CurrentGameState.Player.CurrentWeapon = currentWeapon
	CurrentGameState.Backpack = currentBackpack
	CurrentGameState.Player.ActiveEffects = currentEffects

	// Обновляем видимость
	gl.updateVisibility()

	gl.AddMessage("Welcome to level " + strconv.Itoa(nextLevelNum) + "!")
}

func (gl *GameLogic) generateItem(levelNum int) *Item {
	itemTypes := []ItemType{Food, Potion, Scroll, Weapon, Treasure}
	itemType := itemTypes[rand.Intn(len(itemTypes))]

	switch itemType {
	case Food:
		return &Item{
			Type:        Food,
			HealthBoost: 10 + levelNum*2,
			Value:       5 + levelNum,
		}
	case Potion:
		subTypes := []ItemSubType{HealthPotion, AgilityPotion, StrengthPotion}
		subType := subTypes[rand.Intn(len(subTypes))]

		boost := 2 + levelNum
		item := &Item{
			Type:    Potion,
			SubType: subType,
			Value:   10 + levelNum*2,
		}

		switch subType {
		case HealthPotion:
			item.HealthBoost = 15 + levelNum*3
		case AgilityPotion:
			item.AgilityBoost = boost
		case StrengthPotion:
			item.StrengthBoost = boost
		}
		return item

	case Scroll:
		subTypes := []ItemSubType{MaxHealthScroll, AgilityScroll, StrengthScroll}
		subType := subTypes[rand.Intn(len(subTypes))]

		boost := 1 + levelNum/2
		if boost < 1 {
			boost = 1
		}

		item := &Item{
			Type:    Scroll,
			SubType: subType,
			Value:   15 + levelNum*3,
		}

		switch subType {
		case MaxHealthScroll:
			item.MaxHealthBoost = boost
		case AgilityScroll:
			item.AgilityBoost = boost
		case StrengthScroll:
			item.StrengthBoost = boost
		}
		return item

	case Weapon:
		subTypes := []ItemSubType{Sword, Axe}
		subType := subTypes[rand.Intn(len(subTypes))]

		return &Item{
			Type:          Weapon,
			SubType:       subType,
			StrengthBoost: 3 + levelNum,
			Value:         20 + levelNum*4,
		}

	case Treasure:
		return &Item{
			Type:  Treasure,
			Value: 25 + levelNum*5,
		}

	default:
		return &Item{
			Type:        Food,
			HealthBoost: 10,
			Value:       5,
		}
	}
}

func (gl *GameLogic) getRandomPositionInRoom(room *Room) Position {
	// Убедимся, что позиция не в стенах комнаты (минимум 1 клетка от края)
	return Position{
		X: room.X + 1 + rand.Intn(max(1, room.Width-2)),
		Y: room.Y + 1 + rand.Intn(max(1, room.Height-2)),
	}
func (gl *GameLogic) generateDoorsAndKeys(level *Level, levelNum int) {
	colors := []ItemSubType{KeyRed, KeyBlue, KeyGreen, KeyYellow}
	maxAttempts := 10
	attempts := 0

	for attempts < maxAttempts {
		// Очищаем предыдущие двери и ключи
		level.Doors = make([]*Door, 0)

		// Удаляем предыдущие ключи из Items
		newItems := make([]*Item, 0)
		for _, item := range CurrentGameState.Items {
			if item.Type != Scroll ||
				(item.SubType != KeyRed && item.SubType != KeyBlue &&
					item.SubType != KeyGreen && item.SubType != KeyYellow) {
				newItems = append(newItems, item)
			}
		}
		CurrentGameState.Items = newItems

		// Создаем 1-2 двери на уровень
		numDoors := 1 + rand.Intn(2)
		doorsCreated := 0

		for i := 0; i < numDoors && doorsCreated < numDoors; i++ {
			if len(level.Rooms) < 2 {
				continue
			}

			// Выбираем две случайные комнаты
			roomIdx1 := rand.Intn(len(level.Rooms))
			roomIdx2 := (roomIdx1 + 1 + rand.Intn(len(level.Rooms)-1)) % len(level.Rooms)

			room1 := level.Rooms[roomIdx1]
			room2 := level.Rooms[roomIdx2]

			// Находим точку соединения
			doorPos := gl.findConnectionPoint(room1, room2)

			// Проверяем, что позиция валидна
			if doorPos.X < 0 || doorPos.X >= len(level.Grid) ||
				doorPos.Y < 0 || doorPos.Y >= len(level.Grid[0]) ||
				level.Grid[doorPos.X][doorPos.Y].IsWall {
				continue
			}

			color := colors[rand.Intn(len(colors))]

			door := &Door{
				Position: doorPos,
				Color:    color,
				IsLocked: true,
			}

			level.Doors = append(level.Doors, door)

			// Создаем ключ в одной из комнат
			keyRoom := room1
			if rand.Intn(100) < 50 {
				keyRoom = room2
			}

			key := &Item{
				Type:    Scroll,
				SubType: color,
				Position: Position{
					X: keyRoom.X + 1 + rand.Intn(max(1, keyRoom.Width-2)),
					Y: keyRoom.Y + 1 + rand.Intn(max(1, keyRoom.Height-2)),
				},
				Name: gl.getKeyName(color),
			}

			CurrentGameState.Items = append(CurrentGameState.Items, key)
			keyRoom.Items = append(keyRoom.Items, key)

			doorsCreated++
		}

		// Проверяем доступность ключей
		if gl.validateKeyAccessibility(level) {
			gl.AddMessage("Level " + strconv.Itoa(levelNum) +
				" generated with " + strconv.Itoa(len(level.Doors)) + " doors")
			return
		}

		attempts++
	}

	// Если не удалось создать валидную конфигурацию после maxAttempts попыток,
	// создаем уровень без заблокированных дверей
	gl.AddMessage("Warning: Could not generate valid door configuration after " +
		strconv.Itoa(maxAttempts) + " attempts. Creating level without locked doors.")

	for i := range level.Doors {
		level.Doors[i].IsLocked = false
	}
}

func (gl *GameLogic) findConnectionPoint(room1, room2 *Room) Position {
	// Более сложный алгоритм поиска точки соединения
	// Пытаемся найти точку на прямой между центрами комнат

	center1 := Position{
		X: room1.X + room1.Width/2,
		Y: room1.Y + room1.Height/2,
	}
	center2 := Position{
		X: room2.X + room2.Width/2,
		Y: room2.Y + room2.Height/2,
	}

	// Используем алгоритм Брезенхема для поиска подходящей точки
	dx := abs(center2.X - center1.X)
	dy := abs(center2.Y - center1.Y)

	var sx, sy int
	if center1.X < center2.X {
		sx = 1
	} else {
		sx = -1
	}
	if center1.Y < center2.Y {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy

	x, y := center1.X, center1.Y
	for {
		// Проверяем, является ли точка проходимой и не стеной
		if x >= 0 && x < len(CurrentGameState.Level.Grid) &&
			y >= 0 && y < len(CurrentGameState.Level.Grid[0]) &&
			!CurrentGameState.Level.Grid[x][y].IsWall {
			return Position{X: x, Y: y}
		}

		if x == center2.X && y == center2.Y {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	// Если не нашли подходящую точку на прямой, возвращаем середину
	return Position{
		X: (center1.X + center2.X) / 2,
		Y: (center1.Y + center2.Y) / 2,
	}
}

func (gl *GameLogic) getKeyName(color ItemSubType) string {
	switch color {
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

func (gl *GameLogic) UseKey(key *Item) {
	playerPos := CurrentGameState.Player.Position

	// Проверяем соседние клетки на наличие дверей
	directions := []Position{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for _, dir := range directions {
		checkX, checkY := playerPos.X+dir.X, playerPos.Y+dir.Y

		for i, door := range CurrentGameState.Level.Doors {
			if door.Position.X == checkX && door.Position.Y == checkY &&
				door.IsLocked && door.Color == key.SubType {

				// Открываем дверь
				CurrentGameState.Level.Doors[i].IsLocked = false
				gl.AddMessage("You unlocked the " + gl.getKeyName(key.SubType) + " door!")

				// Удаляем ключ из инвентаря
				for j, scroll := range CurrentGameState.Backpack.Scrolls {
					if scroll == key {
						CurrentGameState.Backpack.Scrolls = append(
							CurrentGameState.Backpack.Scrolls[:j],
							CurrentGameState.Backpack.Scrolls[j+1:]...,
						)
						break
					}
				}
				return
			}
		}
	}

	gl.AddMessage("No matching locked door nearby")
	// Ключ НЕ удаляется из инвентаря, если нет подходящей двери
}

func (gl *GameLogic) adjustDifficulty() {
	if CurrentGameState.PlayerProgress == nil {
		return
	}

	progress := CurrentGameState.PlayerProgress
	player := CurrentGameState.Player

	// Анализируем прогресс игрока
	survivalRate := float64(player.Health) / float64(player.MaxHealth)

	// Определяем нужна ли корректировка сложности (каждые 50 ходов)
	if CurrentGameState.GameSession.CellsWalked-progress.LastAdjustmentTime > 50 {
		if survivalRate < 0.3 {
			// Игроку сложно - уменьшаем сложность
			progress.DifficultyLevel = max(1, progress.DifficultyLevel-1)
			gl.AddMessage("The dungeon feels less dangerous...")
		} else if survivalRate > 0.8 {
			// Игроку легко - увеличиваем сложность
			progress.DifficultyLevel = min(5, progress.DifficultyLevel+1)
			gl.AddMessage("The dungeon grows more treacherous...")
		}

		progress.LastAdjustmentTime = CurrentGameState.GameSession.CellsWalked
	}
}

// Функция для движения Mimic
func (gl *GameLogic) moveMimic(enemy *Enemy) {
	playerPos := CurrentGameState.Player.Position
	dist := abs(enemy.Position.X-playerPos.X) + abs(enemy.Position.Y-playerPos.Y)

	// Мимик раскрывается, когда игрок подходит близко
	if enemy.IsDisguised && dist <= 2 {
		enemy.IsDisguised = false
		gl.AddMessage("The item was a Mimic!")
	}

	// Если раскрыт, ведет себя как обычный враг
	if !enemy.IsDisguised {
		gl.moveChase(enemy)
	}
	// Если замаскирован, не двигается
}

// buildAccessibilityGraph строит граф доступности уровня
func (gl *GameLogic) buildAccessibilityGraph(level *Level) *AccessibilityGraph {
	graph := &AccessibilityGraph{Nodes: make([]*GraphNode, 0)}
	nodeMap := make(map[Position]*GraphNode)

	// Создаем узлы для всех проходимых клеток
	for x := 0; x < len(level.Grid); x++ {
		for y := 0; y < len(level.Grid[0]); y++ {
			if !level.Grid[x][y].IsWall {
				pos := Position{X: x, Y: y}
				node := &GraphNode{
					Position: pos,
					IsDoor:   false,
					IsLocked: false,
				}

				// Проверяем, есть ли здесь дверь
				for _, door := range level.Doors {
					if door.Position.X == x && door.Position.Y == y {
						node.IsDoor = true
						node.IsLocked = door.IsLocked
						node.KeyType = door.Color
						break
					}
				}

				graph.Nodes = append(graph.Nodes, node)
				nodeMap[pos] = node
			}
		}
	}

	// Создаем связи между соседними узлами
	for _, node := range graph.Nodes {
		directions := []Position{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

		for _, dir := range directions {
			neighborPos := Position{X: node.Position.X + dir.X, Y: node.Position.Y + dir.Y}
			if neighbor, exists := nodeMap[neighborPos]; exists {
				node.Neighbors = append(node.Neighbors, neighbor)
			}
		}
	}

	return graph
}

// bfsSearch выполняет поиск в ширину для проверки доступности ключей
func (gl *GameLogic) bfsSearch(start *GraphNode, targetKey ItemSubType) bool {
	queue := []*GraphNode{start}
	visited := make(map[Position]bool)
	visited[start.Position] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Если нашли ключ нужного типа - возвращаем true
		for _, item := range CurrentGameState.Items {
			if item.Position.X == current.Position.X &&
				item.Position.Y == current.Position.Y &&
				item.Type == Scroll && // Ключи используют тип Scroll
				item.SubType == targetKey {
				return true
			}
		}

		// Добавляем соседей в очередь
		for _, neighbor := range current.Neighbors {
			if !visited[neighbor.Position] {
				// Если это заблокированная дверь - проверяем, можем ли пройти
				if neighbor.IsDoor && neighbor.IsLocked {
					if neighbor.KeyType == targetKey {
						// Это дверь того же типа, что и нужный ключ - можем пройти
						visited[neighbor.Position] = true
						queue = append(queue, neighbor)
					}
					// Заблокированные двери других типов пропускаем
					continue
				}

				visited[neighbor.Position] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return false
}

// dfsSearch выполняет поиск в глубину для проверки доступности
func (gl *GameLogic) dfsSearch(start *GraphNode, targetKey ItemSubType, visited map[Position]bool) bool {
	if visited == nil {
		visited = make(map[Position]bool)
	}

	visited[start.Position] = true

	// Проверяем, есть ли здесь нужный ключ
	for _, item := range CurrentGameState.Items {
		if item.Position.X == start.Position.X &&
			item.Position.Y == start.Position.Y &&
			item.Type == Scroll &&
			item.SubType == targetKey {
			return true
		}
	}

	// Рекурсивно проверяем соседей
	for _, neighbor := range start.Neighbors {
		if !visited[neighbor.Position] {
			// Обрабатываем заблокированные двери
			if neighbor.IsDoor && neighbor.IsLocked {
				if neighbor.KeyType == targetKey {
					// Можем пройти через дверь того же типа
					if gl.dfsSearch(neighbor, targetKey, visited) {
						return true
					}
				}
				continue
			}

			if gl.dfsSearch(neighbor, targetKey, visited) {
				return true
			}
		}
	}

	return false
}

// validateKeyAccessibility проверяет доступность всех ключей на уровне
func (gl *GameLogic) validateKeyAccessibility(level *Level) bool {
	graph := gl.buildAccessibilityGraph(level)

	// Находим стартовую позицию игрока
	var startNode *GraphNode
	for _, node := range graph.Nodes {
		if node.Position.X == level.StartRoom.X+level.StartRoom.Width/2 &&
			node.Position.Y == level.StartRoom.Y+level.StartRoom.Height/2 {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return false
	}

	// Проверяем доступность каждого ключа
	for _, door := range level.Doors {
		if door.IsLocked {
			// Используем BFS для проверки доступности ключа
			keyAccessible := gl.bfsSearch(startNode, door.Color)

			if !keyAccessible {
				gl.AddMessage("WARNING: Key for " + gl.getKeyName(door.Color) +
					" door is not accessible from start!")
				return false
			}
		}
	}

	return true
}
