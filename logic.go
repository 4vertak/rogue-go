package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
	"strconv"
)

// GameLogic реализует бизнес-логику игры
type GameLogic struct{}

func init() {
	GameActionsImpl = &GameLogic{}
	rand.Seed(time.Now().UnixNano())
}

func (gl *GameLogic) MovePlayer(dx, dy int) {
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

        // Если на клетке есть предметы — автоматически подбираем (если место в рюкзаке)
        picked := false
        for i := len(CurrentGameState.Items) - 1; i >= 0; i-- {
            item := CurrentGameState.Items[i]
            if item.Position.X == newX && item.Position.Y == newY {
                if gl.canAddToBackpack(item) {
                    gl.addToBackpack(item)
                    // удалить из глобального списка предметов
                    CurrentGameState.Items = append(CurrentGameState.Items[:i], CurrentGameState.Items[i+1:]...)
                    // удалить из комнаты, если он там указан
                    if CurrentGameState.Level != nil {
                        for _, room := range CurrentGameState.Level.Rooms {
                            for ri := len(room.Items) - 1; ri >= 0; ri-- {
                                if room.Items[ri] == item || (room.Items[ri].Position == item.Position) {
                                    room.Items = append(room.Items[:ri], room.Items[ri+1:]...)
                                }
                            }
                        }
                    }
                    picked = true
                } else {
                    gl.AddMessage("Item here but backpack is full for this type")
                }
            }
        }
        if picked {
            gl.AddMessage("You picked up available items automatically")
        }

        // После движения проверяем автоматический переход на следующий уровень
        gl.checkForAutomaticExit()
        gl.checkForEnemies() // Бой при движении игрока
        gl.checkForExit()
    }
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
        gl.applyScrollEffect(scroll)
        CurrentGameState.Backpack.Scrolls = append(
            CurrentGameState.Backpack.Scrolls[:index],
            CurrentGameState.Backpack.Scrolls[index+1:]...,
        )
        CurrentGameState.GameSession.ScrollsUsed++
        gl.AddMessage("Used scroll")
    
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
            
            // Если у игрока уже есть оружие - бросаем его на пол
            if CurrentGameState.Player.CurrentWeapon != nil {
                currentWeapon := CurrentGameState.Player.CurrentWeapon
                
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
}

func contains(s, substr string) bool {
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
    
    if x < 1 { x = 1 }
    if y < 1 { y = 1 }
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
        X:      x,
        Y:      y,
        Width:  roomWidth,
        Height: roomHeight,
        Enemies: make([]*Enemy, 0),
        Items:   make([]*Item, 0),
    }
    
    // Заполняем комнату полом
    for i := x; i < x+roomWidth; i++ {
        for j := y; j < y+roomHeight; j++ {
            if i >= 0 && i < len(level.Grid) && j >= 0 && j < len(level.Grid[0]) {
                level.Grid[i][j].IsWall = false
            }
        }
    }
    
    level.Rooms = append(level.Rooms, room)
    return true
}

func (gl *GameLogic) connectRoomsWithCorridors(level *Level) {
    // Соединяем комнаты в порядке их создания
    for i := 0; i < len(level.Rooms)-1; i++ {
        room1 := level.Rooms[i]
        room2 := level.Rooms[i+1]
        
        center1X := room1.X + room1.Width/2
        center1Y := room1.Y + room1.Height/2
        center2X := room2.X + room2.Width/2
        center2Y := room2.Y + room2.Height/2
        
        // Создаем L-образный коридор
        corridorWidth := 1
        corridor := &Corridor{
            Start: Position{X: center1X, Y: center1Y},
            End:   Position{X: center2X, Y: center2Y},
            Width: corridorWidth,
        }
        
        // Горизонтальная часть
        startX, endX := center1X, center2X
        if startX > endX {
            startX, endX = endX, startX
        }
        for x := startX; x <= endX; x++ {
            for w := 0; w < corridorWidth; w++ {
                y := center1Y + w
                if x >= 0 && x < len(level.Grid) && y >= 0 && y < len(level.Grid[0]) {
                    level.Grid[x][y].IsWall = false
                    level.Grid[x][y].IsCorridor = true
                }
            }
        }
        
        // Вертикальная часть
        startY, endY := center1Y, center2Y
        if startY > endY {
            startY, endY = endY, startY
        }
        for y := startY; y <= endY; y++ {
            for w := 0; w < corridorWidth; w++ {
                x := center2X + w
                if x >= 0 && x < len(level.Grid) && y >= 0 && y < len(level.Grid[0]) {
                    level.Grid[x][y].IsWall = false
                    level.Grid[x][y].IsCorridor = true
                }
            }
        }
        
        level.Corridors = append(level.Corridors, corridor)
    }
}

func (gl *GameLogic) populateLevel(level *Level, levelNum int) {
    for _, room := range level.Rooms {
        // пропускаем стартовую комнату
        if room == level.StartRoom {
            continue
        }

        numEnemies := 1 + rand.Intn(2) + levelNum/2
        if numEnemies > 3 {
            numEnemies = 3
        }
        for i := 0; i < numEnemies; i++ {
            if rand.Intn(100) < 70 {
                enemy := gl.createEnemy(levelNum, room)
                if enemy != nil {
                    room.Enemies = append(room.Enemies, enemy)
                    CurrentGameState.Enemies = append(CurrentGameState.Enemies, enemy)
                }
            }
        }

        numItems := 1 + rand.Intn(2)
        for i := 0; i < numItems; i++ {
            if rand.Intn(100) < 60 {
                item := gl.createItem(levelNum, room, level)
                if item != nil {
                    room.Items = append(room.Items, item)
                    CurrentGameState.Items = append(CurrentGameState.Items, item)
                }
            }
        }
    }
}

func (gl *GameLogic) createEnemy(levelNum int, room *Room) *Enemy {
	enemyTypes := []EnemyType{Zombie, Vampire, Ghost, Ogre, SnakeMage}
	weights := []int{40, 20, 15, 15, 10}

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

	health := 6 + levelNum*3
	agility := 4 + levelNum
	strength := 3 + levelNum
    hostility := 6 + levelNum*2 // Базовая враждебность
    
    switch chosenType {
    case Zombie:
        health += 6
        agility -= 1
        strength += 1
        hostility += 2 // Зомби имеют большую зону агрессии
    case Vampire:
        health += 3
        agility += 2
        strength += 1
        hostility += 3 // Вампиры очень агрессивны
    case Ghost:
        health += 1
        agility += 3
        strength -= 1
        hostility -= 2 // Призраки менее агрессивны
    case Ogre:
        health += 10
        agility -= 2
        strength += 3
        hostility += 1 // Огры умеренно агрессивны
    case SnakeMage:
        health += 2
        agility += 2
        strength -= 1
        hostility += 4 // Змеи-маги очень агрессивны
    }

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
		Type:               chosenType,
		Health:             health,
		Agility:            agility,
		Strength:           strength,
		Hostility:          hostility,
		Position:           pos,
		IsVisible:          true,
		Pattern:            pattern,
		moveCooldown:       0,
		hasBeenHit:         false,
		attackCooldown:     0,
		teleportCooldown:   0,
		invisibilityCooldown: 0,
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
    // Обновляем видимость
    gl.updateVisibility()
    
    // Двигаем врагов КАЖДЫЙ ход
    gl.moveEnemies()
    
    // Проверяем автоматический выход
    gl.checkForAutomaticExit()
    
    // Проверяем близость к выходу (подсказка для игрока)
    gl.checkExitProximity()
    
    // Проверяем условия игры
    gl.CheckGameConditions()
    
    // Обновляем камеру
    gl.updateCamera()
    
    // Обновляем эффекты
    gl.updateEffects()
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
    return !CurrentGameState.Level.Grid[x][y].IsWall
}

func (gl *GameLogic) isInAttackRange(pos1, pos2 Position) bool {
    dx := abs(pos1.X - pos2.X)
    dy := abs(pos1.Y - pos2.Y)
    return dx <= 1 && dy <= 1
}

func (gl *GameLogic) resolveCombat(attacker interface{}, defender interface{}) {
	var attackerStr, attackerAgi, defenderAgi int

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

	switch d := defender.(type) {
	case *Character:
		defenderAgi = d.Agility
	case *Enemy:
		defenderAgi = d.Agility
	}

	hitChance := 65 + (attackerAgi-defenderAgi)*2
	if hitChance < 20 {
		hitChance = 20
	}
	if hitChance > 95 {
		hitChance = 95
	}

	if rand.Intn(100) < hitChance {
		damage := attackerStr
		switch d := defender.(type) {
		case *Character:
			if _, ok := attacker.(*Enemy); ok {
				damage = attackerStr / 2 // враги бьют слабее
			}
			d.Health -= damage
			CurrentGameState.GameSession.HitsTaken++
			gl.AddMessage("You take " + strconv.Itoa(damage) + " damage!")

			// спецэффекты врагов
			if e, ok := attacker.(*Enemy); ok {
				gl.applyEnemySpecialEffects(e, d)
			}

		case *Enemy:
			if _, ok := attacker.(*Character); ok {
				// игрок наносит полный урон
			} else {
				damage = attackerStr / 2
			}
			d.Health -= damage
			CurrentGameState.GameSession.HitsLanded++
			gl.AddMessage("You hit enemy for " + strconv.Itoa(damage) + " damage!")

			// особенность вампира: первый удар = дренаж max HP
			if d.Type == Vampire && !d.hasBeenHit {
				d.hasBeenHit = true
				if char, ok := attacker.(*Character); ok {
					char.MaxHealth -= 3
					if char.Health > char.MaxHealth {
						char.Health = char.MaxHealth
					}
					gl.AddMessage("Vampire drained your max health! -3 Max HP")
				}
			}
		}
	} else {
		gl.AddMessage("Attack missed!")
	}
}

func (gl *GameLogic) applyEnemySpecialEffects(enemy *Enemy, player *Character) {
    switch enemy.Type {
    case SnakeMage:
        // Змей-маг усыпляет игрока
        if rand.Intn(100) < 30 { // 30% chance
            CurrentGameState.SleepTurns = 2 + rand.Intn(3)
            gl.AddMessage("Snake Mage puts you to sleep for " + 
                strconv.Itoa(CurrentGameState.SleepTurns) + " turns!")
        }
    }
}

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

var enemyTurnCounter int

func (gl *GameLogic) moveEnemies() {
    enemyTurnCounter++
    
    // Монстры двигаются КАЖДЫЙ ход, но с разной скоростью в зависимости от типа
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
            switch enemy.Pattern {
            case Random:
                gl.moveRandom(enemy)
            case Chase:
                gl.moveChase(enemy)
            case Teleport:
                gl.moveTeleport(enemy)
            case Diagonal:
                gl.moveDiagonal(enemy)
            case TwoStep:
                gl.moveTwoStep(enemy)
            }
        } else {
            // Игрок вне зоны агрессии - патрулируем
            switch enemy.Pattern {
            case Random:
                gl.moveRandom(enemy)
            case Teleport:
                gl.moveRandom(enemy) // Призрак телепортируется только при преследовании
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
    if enemyPos.X < playerPos.X { dx = 1 }
    if enemyPos.X > playerPos.X { dx = -1 }
    
    dy := 0
    if enemyPos.Y < playerPos.Y { dy = 1 }
    if enemyPos.Y > playerPos.Y { dy = -1 }
    
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
        if enemyPos.X > playerPos.X { dx = -1 }
        dy := 1
        if enemyPos.Y > playerPos.Y { dy = -1 }
        
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
                    CurrentGameState.Level.Discovered[x][y] = true
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
}

func (gl *GameLogic) hasLineOfSight(start, end Position) bool {
    // Простой алгоритм Брезенхема для проверки видимости
    dx := abs(end.X - start.X)
    dy := abs(end.Y - start.Y)
    
    var sx, sy int
    if start.X < end.X { sx = 1 } else { sx = -1 }
    if start.Y < end.Y { sy = 1 } else { sy = -1 }
    
    err := dx - dy
    
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
        
        if enemy.Position.X == playerPos.X && enemy.Position.Y == playerPos.Y {
            gl.resolveCombat(CurrentGameState.Player, enemy)
            
            if enemy.Health <= 0 {
                gl.handleEnemyDefeat(enemy)
                // Удаляем врага из списка
                CurrentGameState.Enemies = append(
                    CurrentGameState.Enemies[:i],
                    CurrentGameState.Enemies[i+1:]...,
                )
                break // Выходим после первого убитого врага
            }
        }
    }
}

func (gl *GameLogic) checkForExit() {
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

func (gl *GameLogic) addToBackpack(item *Item) {
    if item == nil {
        return
    }
    if CurrentGameState.Backpack == nil {
        CurrentGameState.Backpack = &Backpack{
            Weapons: make([]*Item, 0),
            Food:    make([]*Item, 0),
            Potions: make([]*Item, 0),
            Scrolls: make([]*Item, 0),
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
        CurrentGameState.Backpack.Scrolls = append(CurrentGameState.Backpack.Scrolls, item)
        gl.AddMessage("Picked up scroll")
        
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
    if cam.X < 0 { cam.X = 0 }
    if cam.Y < 0 { cam.Y = 0 }
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