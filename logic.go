package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
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
}

func (gl *GameLogic) UseItem(itemType ItemType, index int) {
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

	case Scroll:
		if index < len(CurrentGameState.Backpack.Scrolls) {
			scroll := CurrentGameState.Backpack.Scrolls[index]
			gl.applyScrollEffect(scroll)
			CurrentGameState.Backpack.Scrolls = append(
				CurrentGameState.Backpack.Scrolls[:index],
				CurrentGameState.Backpack.Scrolls[index+1:]...,
			)
			CurrentGameState.GameSession.ScrollsUsed++
		}

	case Weapon:
		if index == 0 {
			// Unequip current weapon - кладем в рюкзак, а не на пол
			if CurrentGameState.Player.CurrentWeapon != nil {
				weapon := CurrentGameState.Player.CurrentWeapon
				// Добавляем оружие в рюкзак
				CurrentGameState.Backpack.Weapons = append(CurrentGameState.Backpack.Weapons, weapon)
				CurrentGameState.Player.CurrentWeapon = nil
				gl.AddMessage("Weapon unequipped (moved to backpack)")
			}
		} else if index-1 < len(CurrentGameState.Backpack.Weapons) {
			newWeapon := CurrentGameState.Backpack.Weapons[index-1]

			// If player already has a weapon, put it back in backpack
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
			CurrentGameState.Player.CurrentWeapon = newWeapon

			// Remove the equipped weapon from backpack
			CurrentGameState.Backpack.Weapons = append(
				CurrentGameState.Backpack.Weapons[:index-1],
				CurrentGameState.Backpack.Weapons[index:]...,
			)
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
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func (gl *GameLogic) PickUpItem() {
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
		Discovered: make([][]bool, mapWidth),
		Visible:    make([][]bool, mapWidth),
		Rooms:      make([]*Room, 0),
		Corridors:  make([]*Corridor, 0),
	}

	for i := range level.Grid {
		level.Grid[i] = make([]Cell, mapHeight)
		level.Discovered[i] = make([]bool, mapHeight)
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
}

func (gl *GameLogic) SaveGame(filename string) error {
	data, err := json.MarshalIndent(CurrentGameState, "", "  ")
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
	gl.moveEnemies()
	gl.updateItemEffects()
	gl.CheckGameConditions()
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
	}
}

func (gl *GameLogic) applyScrollEffect(scroll *Item) {
	switch scroll.SubType {
	case MaxHealthScroll:
		CurrentGameState.Player.MaxHealth += scroll.MaxHealthBoost
		CurrentGameState.Player.Health += scroll.MaxHealthBoost
	case AgilityScroll:
		CurrentGameState.Player.Agility += scroll.AgilityBoost
	case StrengthScroll:
		CurrentGameState.Player.Strength += scroll.StrengthBoost
	}
}

func (gl *GameLogic) canMoveTo(x, y int) bool {
	if x < 0 || y < 0 ||
		x >= len(CurrentGameState.Level.Grid) ||
		y >= len(CurrentGameState.Level.Grid[0]) {
		return false
	}

	return !CurrentGameState.Level.Grid[x][y].IsWall
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
				}
			}
		}
	}
}

func (gl *GameLogic) hasLineOfSight(start, end Position) bool {
	dx := abs(end.X - start.X)
	dy := abs(end.Y - start.Y)
	sx, sy := 1, 1
	if start.X > end.X {
		sx = -1
	}
	if start.Y > end.Y {
		sy = -1
	}

	err := dx - dy
	x, y := start.X, start.Y

	for {
		if x == end.X && y == end.Y {
			return true
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
}

func (gl *GameLogic) checkForItems() {
	gl.PickUpItem()
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
}

func (gl *GameLogic) generateEnemy(levelNum int) *Enemy {
	enemyTypes := []EnemyType{Zombie, Vampire, Ghost, Ogre, SnakeMage, Mimic}
	enemyType := enemyTypes[5]

	baseHealth := 20 + levelNum*5
	baseAgility := 5 + levelNum
	baseStrength := 5 + levelNum
	baseHostility := 3 + levelNum/2

	switch enemyType {
	case Zombie:
		return &Enemy{
			Type:      Zombie,
			Health:    baseHealth * 3 / 2,
			Agility:   baseAgility / 2,
			Strength:  baseStrength,
			Hostility: baseHostility,
			IsVisible: true,
			Pattern:   Random,
		}
	case Vampire:
		return &Enemy{
			Type:      Vampire,
			Health:    baseHealth * 4 / 3,
			Agility:   baseAgility * 4 / 3,
			Strength:  baseStrength,
			Hostility: baseHostility * 4 / 3,
			IsVisible: true,
			Pattern:   Chase,
		}
	case Ghost:
		return &Enemy{
			Type:      Ghost,
			Health:    baseHealth / 2,
			Agility:   baseAgility * 3 / 2,
			Strength:  baseStrength / 2,
			Hostility: baseHostility / 2,
			IsVisible: false,
			Pattern:   Teleport,
		}
	case Ogre:
		return &Enemy{
			Type:      Ogre,
			Health:    baseHealth * 2,
			Agility:   baseAgility / 3,
			Strength:  baseStrength * 2,
			Hostility: baseHostility,
			IsVisible: true,
			Pattern:   TwoStep,
		}
	case SnakeMage:
		return &Enemy{
			Type:      SnakeMage,
			Health:    baseHealth,
			Agility:   baseAgility * 5 / 3,
			Strength:  baseStrength / 2,
			Hostility: baseHostility * 3 / 2,
			IsVisible: true,
			Pattern:   Diagonal,
		}
	case Mimic:
		return &Enemy{
			Type:      Mimic,
			Health:    baseHealth * 4 / 3,
			Agility:   baseAgility * 4 / 3,
			Strength:  baseStrength / 2,
			Hostility: baseHostility / 2,
			IsVisible: false,
			Pattern:   Imitator,
		}
	default:
		return &Enemy{
			Type:      Zombie,
			Health:    baseHealth,
			Agility:   baseAgility,
			Strength:  baseStrength,
			Hostility: baseHostility,
			IsVisible: true,
			Pattern:   Random,
		}
	}
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
}

func (gl *GameLogic) moveEnemies() {
	for _, enemy := range CurrentGameState.Enemies {
		if gl.shouldChasePlayer(enemy) {
			gl.chasePlayer(enemy)
		} else if enemy.Type != Mimic {
			gl.moveRandomly(enemy)
		}

		playerPos := CurrentGameState.Player.Position
		if enemy.Position.X == playerPos.X && enemy.Position.Y == playerPos.Y {
			gl.resolveCombat(enemy, CurrentGameState.Player)

			if CurrentGameState.Player.Health <= 0 {
				CurrentGameState.GameOver = true
			}
		}
	}
}

func (gl *GameLogic) shouldChasePlayer(enemy *Enemy) bool {
	playerPos := CurrentGameState.Player.Position
	dx := playerPos.X - enemy.Position.X
	dy := playerPos.Y - enemy.Position.Y
	distance := dx*dx + dy*dy
	if enemy.Type == Mimic {
		if abs(enemy.Position.X-playerPos.X)+abs(enemy.Position.Y-playerPos.Y) > 1 {
			return false
		} else {
			enemy.IsVisible = true
			return true
		}

	}

	return distance <= enemy.Hostility*enemy.Hostility
}

func (gl *GameLogic) chasePlayer(enemy *Enemy) {
	playerPos := CurrentGameState.Player.Position

	// Если уже на соседней клетке с игроком - не двигаться дальше
	if abs(enemy.Position.X-playerPos.X) <= 1 && abs(enemy.Position.Y-playerPos.Y) <= 1 {
		return
	}

	dx, dy := 0, 0
	if playerPos.X > enemy.Position.X {
		dx = 1
	} else if playerPos.X < enemy.Position.X {
		dx = -1
	}

	if playerPos.Y > enemy.Position.Y {
		dy = 1
	} else if playerPos.Y < enemy.Position.Y {
		dy = -1
	}

	newX := enemy.Position.X + dx
	newY := enemy.Position.Y + dy

	if gl.canMoveTo(newX, newY) {
		enemy.Position.X = newX
		enemy.Position.Y = newY
	}
}

func (gl *GameLogic) moveRandomly(enemy *Enemy) {
	dx, dy := rand.Intn(3)-1, rand.Intn(3)-1
	newX := enemy.Position.X + dx
	newY := enemy.Position.Y + dy

	if gl.canMoveTo(newX, newY) {
		enemy.Position.X = newX
		enemy.Position.Y = newY
	}
}

func (gl *GameLogic) canAddToBackpack(item *Item) bool {
	switch item.Type {
	case Treasure:
		return true
	case Food:
		return len(CurrentGameState.Backpack.Food) < 9
	case Potion:
		return len(CurrentGameState.Backpack.Potions) < 9
	case Scroll:
		return len(CurrentGameState.Backpack.Scrolls) < 9
	case Weapon:
		return len(CurrentGameState.Backpack.Weapons) < 9
	default:
		return false
	}
}

func (gl *GameLogic) addToBackpack(item *Item) {
	switch item.Type {
	case Treasure:
		CurrentGameState.Backpack.Treasure += item.Value
		CurrentGameState.Player.Treasures += item.Value
		gl.AddMessage("Picked up treasure: +" + strconv.Itoa(item.Value) + " gold")

	case Food:
		CurrentGameState.Backpack.Food = append(CurrentGameState.Backpack.Food, item)
		gl.AddMessage("Picked up food: +" + strconv.Itoa(item.HealthBoost) + " HP")

	case Potion:
		CurrentGameState.Backpack.Potions = append(CurrentGameState.Backpack.Potions, item)
		potionType := ""
		switch item.SubType {
		case HealthPotion:
			potionType = "Health Potion"
		case AgilityPotion:
			potionType = "Agility Potion"
		case StrengthPotion:
			potionType = "Strength Potion"
		}
		gl.AddMessage("Picked up " + potionType)

	case Scroll:
		CurrentGameState.Backpack.Scrolls = append(CurrentGameState.Backpack.Scrolls, item)
		scrollType := ""
		switch item.SubType {
		case MaxHealthScroll:
			scrollType = "Max Health Scroll"
		case AgilityScroll:
			scrollType = "Agility Scroll"
		case StrengthScroll:
			scrollType = "Strength Scroll"
		}
		gl.AddMessage("Picked up " + scrollType)

	case Weapon:
		// Автоматически экипируем первое оружие, если у игрока его нет
		if CurrentGameState.Player.CurrentWeapon == nil {
			CurrentGameState.Player.CurrentWeapon = item
			gl.AddMessage("Equipped weapon: +" + strconv.Itoa(item.StrengthBoost) + " STR")
		} else {
			// Если уже есть оружие, добавляем в рюкзак
			CurrentGameState.Backpack.Weapons = append(CurrentGameState.Backpack.Weapons, item)
			gl.AddMessage("Picked up weapon (in backpack)")
		}
	}
}

func (gl *GameLogic) handleEnemyDefeat(enemy *Enemy) {
	reward := enemy.Health/10 + enemy.Strength + enemy.Agility
	treasure := &Item{
		Type:  Treasure,
		Value: reward,
	}

	// Создаем копию позиции перед удалением врага
	dropPosition := enemy.Position

	// Обнуляем ссылку на врага в комнате
	if CurrentGameState.Level != nil {
		for _, room := range CurrentGameState.Level.Rooms {
			for i, roomEnemy := range room.Enemies {
				if roomEnemy == enemy {
					// Удаляем врага из комнаты
					room.Enemies = append(room.Enemies[:i], room.Enemies[i+1:]...)
					break
				}
			}
		}
	}

	// Помещаем сокровище на карту
	treasure.Position = dropPosition
	CurrentGameState.Items = append(CurrentGameState.Items, treasure)
	CurrentGameState.Player.DefeatedEnemies++
	CurrentGameState.GameSession.EnemiesDefeated++

	// Обнуляем ссылку на врага
	// Это предотвратит дальнейшие операции с "мертвым" врагом
	enemy = nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
