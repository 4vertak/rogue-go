package gen

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/utils"
)

// параметры генерации
type Config struct{ CellW, CellH, MinRW, MinRH int }

func GenerateLevel(levelNum int, state *entity.GameState) *entity.Level {
	// Создаем уровень фиксированного размера как в классическом Rogue
	mapWidth := 80
	mapHeight := 35

	level := &entity.Level{
		Grid:       make([][]entity.Cell, mapWidth),
		Discovered: make([][]entity.Cell, mapWidth),
		Visible:    make([][]bool, mapWidth),
		Rooms:      make([]*entity.Room, 0),
		Corridors:  make([]*entity.Corridor, 0),
		// Doors:      make([]*entity.Door, 0),
	}

	for i := range level.Grid {
		level.Grid[i] = make([]entity.Cell, mapHeight)
		level.Discovered[i] = make([]entity.Cell, mapHeight)
		level.Visible[i] = make([]bool, mapHeight)
		for j := range level.Grid[i] {
			level.Grid[i][j].IsWall = true // Изначально все стены
		}
	}

	// Генерируем комнаты в сетке 3x3
	generateRoomsInGrid(level)
	// Соединяем комнаты коридорами
	connectRoomsWithCorridors(level)
	// Заполняем уровень врагами и предметами
	populateLevel(level, levelNum, state)

	// на всякий случай обновим видимость
	// rules.UpdateVisibility(state)

	return level
}

func generateRoomsInGrid(level *entity.Level) {
	gridSize := 3
	
	for gridY := 0; gridY < gridSize; gridY++ {
		for gridX := 0; gridX < gridSize; gridX++ {
			// Гарантируем минимум 3 комнаты, затем 50% шанс для остальных
			if len(level.Rooms) < 3 || rand.Intn(100) <= 50 {
				if createRoomInGridCell(level, gridX, gridY, gridSize) {
					// Комната создана успешно
				}
			}
		}
	}
	
	// Устанавливаем стартовую и выходную комнаты
	if len(level.Rooms) > 0 {
		level.StartRoom = level.Rooms[0]
		// Находим самую дальнюю комнату от стартовой для выхода
		maxDistance := 0
		var exitRoom *entity.Room

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

func createRoomInGridCell(level *entity.Level, gridX, gridY, gridSize int) bool {
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
	room := &entity.Room{
		X:       x,
		Y:       y,
		Width:   roomWidth,
		Height:  roomHeight,
		Enemies: make([]*entity.Enemy, 0),
		Items:   make([]*entity.Item, 0),
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

func connectRoomsWithCorridors(level *entity.Level) {
	if len(level.Rooms) <= 1 {
		return
	}

	var edges []entity.Edge
	roomMap := make(map[string]int)

	// ключ = GridX, GridY (их нужно хранить в entity.Room)
	for i, room := range level.Rooms {
		key := fmt.Sprintf("%d,%d", room.GridX, room.GridY)
		roomMap[key] = i
	}

	for i, room := range level.Rooms {
		rightKey := fmt.Sprintf("%d,%d", room.GridX+1, room.GridY)
		if j, exists := roomMap[rightKey]; exists {
			weight := distanceBetweenRooms(room, level.Rooms[j])
			edges = append(edges, entity.Edge{U: i, V: j, Weight: weight})
		}
		downKey := fmt.Sprintf("%d,%d", room.GridX, room.GridY+1)
		if j, exists := roomMap[downKey]; exists {
			weight := distanceBetweenRooms(room, level.Rooms[j])
			edges = append(edges, entity.Edge{U: i, V: j, Weight: weight})
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Weight < edges[j].Weight
	})

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

	// Минимальное остовное дерево (MST)
	// var mstEdges []entity.Edge
	for _, edge := range edges {
		if find(edge.U) != find(edge.V) {
			union(edge.U, edge.V)
			// mstEdges = append(mstEdges, edge)
			connectTwoRooms(level, level.Rooms[edge.U], level.Rooms[edge.V])
		}
	}

	// Проверка связности
	for {
		comps := make(map[int][]int)
		for i := range level.Rooms {
			root := find(i)
			comps[root] = append(comps[root], i)
		}
		if len(comps) <= 1 {
			break
		}

		var bestEdge *entity.Edge
		bestDist := math.MaxFloat64
		var roots []int
		for k := range comps {
			roots = append(roots, k)
		}

		for _, a := range comps[roots[0]] {
			for _, b := range comps[roots[1]] {
				dist := distanceBetweenRooms(level.Rooms[a], level.Rooms[b])
				if dist < bestDist {
					bestDist = dist
					bestEdge = &entity.Edge{U: a, V: b, Weight: dist}
				}
			}
		}

		if bestEdge != nil {
			union(bestEdge.U, bestEdge.V)
			// mstEdges = append(mstEdges, *bestEdge)
			connectTwoRooms(level, level.Rooms[bestEdge.U], level.Rooms[bestEdge.V])
		}
	}

	// Несколько дополнительных соединений
	extraConnections := int(math.Sqrt(float64(len(level.Rooms))))
	for i := 0; i < extraConnections && i < len(edges); i++ {
		randomEdge := edges[rand.Intn(len(edges))]
		if find(randomEdge.U) != find(randomEdge.V) {
			union(randomEdge.U, randomEdge.V)
			connectTwoRooms(level, level.Rooms[randomEdge.U], level.Rooms[randomEdge.V])
		}
	}
}

func distanceBetweenRooms(room1, room2 *entity.Room) float64 {
	x1 := room1.X + room1.Width/2
	y1 := room1.Y + room1.Height/2
	x2 := room2.X + room2.Width/2
	y2 := room2.Y + room2.Height/2
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}

func connectTwoRooms(level *entity.Level, room1, room2 *entity.Room) {
	x1 := room1.X + room1.Width/2
	y1 := room1.Y + room1.Height/2
	x2 := room2.X + room2.Width/2
	y2 := room2.Y + room2.Height/2

	if rand.Intn(2) == 0 {
		createHorizontalCorridor(level, x1, y1, x2)
		createVerticalCorridor(level, x2, y1, y2)
	} else {
		createVerticalCorridor(level, x1, y1, y2)
		createHorizontalCorridor(level, x1, y2, x2)
	}
}

func createHorizontalCorridor(level *entity.Level, x1, y1, x2 int) {
    start, end := x1, x2
    if start > end {
        start, end = end, start
    }
    for x := start; x <= end; x++ {
        if x >= 0 && x < len(level.Grid) && y1 >= 0 && y1 < len(level.Grid[0]) {
            // ПРОВЕРЯЕМ, что это не комната
            if !isInRoom(level, x, y1) {
                level.Grid[x][y1].IsWall = false
                level.Grid[x][y1].IsCorridor = true
            }
        }
    }
}

func createVerticalCorridor(level *entity.Level, x1, y1, y2 int) {
    start, end := y1, y2
    if start > end {
        start, end = end, start
    }
    for y := start; y <= end; y++ {
        if x1 >= 0 && x1 < len(level.Grid) && y >= 0 && y < len(level.Grid[0]) {
            // ПРОВЕРЯЕМ, что это не комната
            if !isInRoom(level, x1, y) {
                level.Grid[x1][y].IsWall = false
                level.Grid[x1][y].IsCorridor = true
            }
        }
    }
}

// isInRoom проверяет, находится ли позиция внутри любой комнаты
func isInRoom(level *entity.Level, x, y int) bool {
    for _, room := range level.Rooms {
        if x >= room.X && x < room.X+room.Width &&
           y >= room.Y && y < room.Y+room.Height {
            return true
        }
    }
    return false
}

func populateLevel(level *entity.Level, levelNum int, state *entity.GameState) {
	// Генерируем двери и ключи (Задание 6)
	if levelNum > 1 {
		generateDoorsAndKeys(level, levelNum, state)
	}

	// Применяем балансировку сложности (Задание 7)
	difficulty := 3 // Средняя сложность по умолчанию
	if state.PlayerProgress != nil {
		difficulty = state.PlayerProgress.DifficultyLevel
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
				enemy := createEnemy(adjustedLevelNum, room)
				if enemy != nil {
					room.Enemies = append(room.Enemies, enemy)
					state.Enemies = append(state.Enemies, enemy)
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
				item := CreateItem(adjustedLevelNum, room, state, level)
				if item != nil {
					room.Items = append(room.Items, item)
					state.Items = append(state.Items, item)
				}
			}
		}
	}
}

func generateDoorsAndKeys(level *entity.Level, levelNum int, state *entity.GameState) {
    // Двери появляются с 2 уровня и выше
    if levelNum < 2 {
        return
    }

    colors := []entity.ItemSubType{entity.KeyRed, entity.KeyBlue, entity.KeyGreen, entity.KeyYellow}
    
    // Простая логика: создаем 1-2 двери
    numDoors := 1
    if levelNum > 3 {
        numDoors = 2
    }

    attempts := 0
    maxAttempts := 10 // Ограничим попытки генерации

    for i := 0; i < numDoors && attempts < maxAttempts; i++ {
        attempts++
        
        if len(level.Rooms) < 2 {
            continue
        }

        // Выбираем две случайные комнаты
        roomIdx1 := rand.Intn(len(level.Rooms))
        roomIdx2 := rand.Intn(len(level.Rooms))
        for roomIdx2 == roomIdx1 {
            roomIdx2 = rand.Intn(len(level.Rooms))
        }

        room1 := level.Rooms[roomIdx1]
        room2 := level.Rooms[roomIdx2]

        // Ищем позицию для двери в КОРИДОРЕ между комнатами
        doorPos := findDoorPositionInCorridor(room1, room2, level)
        if doorPos.X == -1 && doorPos.Y == -1 {
            continue // Не нашли подходящую позицию
        }

        color := colors[rand.Intn(len(colors))]

        door := &entity.Door{
            Position: doorPos,
            Color:    color,
            IsLocked: true,
        }

        // ВАЖНОЕ ИЗМЕНЕНИЕ: Ключ должен быть в той же комнате, что и игрок
        // или в доступной области, а не случайно в одной из двух комнат
        
        // Определяем, в какой комнате находится игрок (стартовая комната)
        startRoom := level.StartRoom
        
        // Выбираем комнату для ключа - всегда в стартовой комнате или доступной области
        var keyRoom *entity.Room
        if room1 == startRoom || isRoomConnectedToStart(room1, level, startRoom) {
            keyRoom = room1
        } else if room2 == startRoom || isRoomConnectedToStart(room2, level, startRoom) {
            keyRoom = room2
        } else {
            // Если ни одна из комнат не подключена к стартовой, пропускаем эту дверь
            continue
        }

        keyPos := findValidKeyPosition(keyRoom, level, doorPos)
        if keyPos.X == -1 && keyPos.Y == -1 {
            continue // Не нашли подходящую позицию для ключа
        }

        key := &entity.Item{
            Type:     entity.Scroll,
            SubType:  color,
            Position: keyPos,
            Name:     utils.GetKeyName(color),
        }

        // ДОБАВЛЯЕМ ПРОВЕРКУ ДОСТУПНОСТИ
        if !validateKeyAccessibility(level, state, key, door) {
            continue // Пропускаем эту дверь/ключ, если ключ недоступен
        }

        // Если проверка пройдена, добавляем дверь и ключ
        level.Doors = append(level.Doors, door)
        state.Items = append(state.Items, key)
        if keyRoom.Items == nil {
            keyRoom.Items = make([]*entity.Item, 0)
        }
        keyRoom.Items = append(keyRoom.Items, key)

        utils.AddMessage(state, "Generated "+utils.GetKeyName(color)+" door")
    }
}

// isRoomConnectedToStart проверяет, подключена ли комната к стартовой комнате
func isRoomConnectedToStart(room *entity.Room, level *entity.Level, startRoom *entity.Room) bool {
    // Простая проверка: если комната соединена коридором со стартовой комнатой
    for _, corridor := range level.Corridors {
        // Проверяем, соединяет ли этот коридор данную комнату со стартовой
        room1Center := entity.Position{
            X: room.X + room.Width/2,
            Y: room.Y + room.Height/2,
        }
        room2Center := entity.Position{
            X: startRoom.X + startRoom.Width/2,
            Y: startRoom.Y + startRoom.Height/2,
        }
        
        // Упрощенная проверка соединения комнат
        if isCorridorBetween(room1Center, room2Center, corridor) {
            return true
        }
    }
    return false
}

// isCorridorBetween проверяет, соединяет ли коридор две точки
func isCorridorBetween(pos1, pos2 entity.Position, corridor *entity.Corridor) bool {
    // Упрощенная проверка: если коридор проходит между центрами комнат
    minX := min(pos1.X, pos2.X)
    maxX := max(pos1.X, pos2.X)
    minY := min(pos1.Y, pos2.Y)
    maxY := max(pos1.Y, pos2.Y)
    
    return (corridor.Start.X >= minX && corridor.Start.X <= maxX &&
            corridor.Start.Y >= minY && corridor.Start.Y <= maxY &&
            corridor.End.X >= minX && corridor.End.X <= maxX &&
            corridor.End.Y >= minY && corridor.End.Y <= maxY)
}

// findDoorPositionInCorridor ищет позицию для двери в коридоре между двумя комнатами
func findDoorPositionInCorridor(room1, room2 *entity.Room, level *entity.Level) entity.Position {
    // Находим центры комнат
    center1 := entity.Position{
        X: room1.X + room1.Width/2,
        Y: room1.Y + room1.Height/2,
    }
    center2 := entity.Position{
        X: room2.X + room2.Width/2,
        Y: room2.Y + room2.Height/2,
    }

    // Определяем направление коридора
    var checkPositions []entity.Position
    
    // Горизонтальный коридор
    if math.Abs(float64(center1.Y-center2.Y)) < math.Abs(float64(center1.X-center2.X)) {
        // Ищем клетки коридора по горизонтали
        startX := min(center1.X, center2.X)
        endX := max(center1.X, center2.X)
        midY := (center1.Y + center2.Y) / 2
        
        for x := startX; x <= endX; x++ {
            if x >= 0 && x < len(level.Grid) && midY >= 0 && midY < len(level.Grid[0]) {
                // Проверяем, что это коридор и не комната
                if level.Grid[x][midY].IsCorridor && !isInRoom(level, x, midY) {
                    checkPositions = append(checkPositions, entity.Position{X: x, Y: midY})
                }
            }
        }
    } else {
        // Вертикальный коридор
        startY := min(center1.Y, center2.Y)
        endY := max(center1.Y, center2.Y)
        midX := (center1.X + center2.X) / 2
        
        for y := startY; y <= endY; y++ {
            if midX >= 0 && midX < len(level.Grid) && y >= 0 && y < len(level.Grid[0]) {
                // Проверяем, что это коридор и не комната
                if level.Grid[midX][y].IsCorridor && !isInRoom(level, midX, y) {
                    checkPositions = append(checkPositions, entity.Position{X: midX, Y: y})
                }
            }
        }
    }

    // Если нашли подходящие позиции, выбираем случайную
    if len(checkPositions) > 0 {
        return checkPositions[rand.Intn(len(checkPositions))]
    }

    // Если не нашли подходящую позицию в коридоре, возвращаем ошибку
    return entity.Position{X: -1, Y: -1}
}

// Упрощенный поиск позиции для двери
func findSimpleDoorPosition(room1, room2 *entity.Room, level *entity.Level) entity.Position {
    // Находим центр комнат
    center1 := entity.Position{
        X: room1.X + room1.Width/2,
        Y: room1.Y + room1.Height/2,
    }
    center2 := entity.Position{
        X: room2.X + room2.Width/2,
        Y: room2.Y + room2.Height/2,
    }

    // Проверяем несколько точек вдоль линии между центрами
    checkPoints := []float64{0.3, 0.5, 0.7}
    for _, ratio := range checkPoints {
        checkX := center1.X + int(float64(center2.X-center1.X)*ratio)
        checkY := center1.Y + int(float64(center2.Y-center1.Y)*ratio)

        if checkX >= 0 && checkX < len(level.Grid) &&
            checkY >= 0 && checkY < len(level.Grid[0]) &&
            !level.Grid[checkX][checkY].IsWall &&
            !isInRoom(level, checkX, checkY) {
            return entity.Position{X: checkX, Y: checkY}
        }
    }

    // Если не нашли подходящую точку, возвращаем середину
    midX := (center1.X + center2.X) / 2
    midY := (center1.Y + center2.Y) / 2

    if midX >= 0 && midX < len(level.Grid) &&
        midY >= 0 && midY < len(level.Grid[0]) &&
        !level.Grid[midX][midY].IsWall {
        return entity.Position{X: midX, Y: midY}
    }

    return entity.Position{X: -1, Y: -1}
}

// Упрощенный поиск позиции для ключа
func findKeyPosition(room *entity.Room, level *entity.Level) entity.Position {
    maxAttempts := 20
    for i := 0; i < maxAttempts; i++ {
        x := room.X + 1 + rand.Intn(max(1, room.Width-2))
        y := room.Y + 1 + rand.Intn(max(1, room.Height-2))

        if x >= 0 && x < len(level.Grid) &&
            y >= 0 && y < len(level.Grid[0]) &&
            !level.Grid[x][y].IsWall {
            return entity.Position{X: x, Y: y}
        }
    }
    return entity.Position{X: -1, Y: -1}
}

func findValidKeyPosition(room *entity.Room, level *entity.Level, doorPos entity.Position) entity.Position {
    maxAttempts := 50
    attempts := 0
    
    for attempts < maxAttempts {
        pos := entity.Position{
            X: room.X + 1 + rand.Intn(max(1, room.Width-2)),
            Y: room.Y + 1 + rand.Intn(max(1, room.Height-2)),
        }
        
        // Проверяем, что позиция валидна и НЕ рядом с дверью
        // Также убеждаемся, что ключ не заблокирован другими дверями
        if pos.X >= room.X && pos.X < room.X+room.Width && 
           pos.Y >= room.Y && pos.Y < room.Y+room.Height &&
           !isNearDoor(pos, doorPos) && // Не рядом с дверью
           !level.Grid[pos.X][pos.Y].IsWall &&
           isPositionAccessible(pos, level) { // ДОБАВЛЯЕМ ПРОВЕРКУ ДОСТУПНОСТИ
            return pos
        }
        attempts++
    }
    
    // Если не нашли подходящую позицию, возвращаем первую доступную
    for x := room.X + 1; x < room.X+room.Width-1; x++ {
        for y := room.Y + 1; y < room.Y+room.Height-1; y++ {
            pos := entity.Position{X: x, Y: y}
            if !level.Grid[x][y].IsWall && isPositionAccessible(pos, level) {
                return pos
            }
        }
    }
    
    return entity.Position{X: -1, Y: -1}
}

// isPositionAccessible проверяет, доступна ли позиция из стартовой комнаты
func isPositionAccessible(pos entity.Position, level *entity.Level) bool {
    // Упрощенная проверка: позиция должна быть в проходимой клетке
    // и не заблокирована другими дверями
    if pos.X < 0 || pos.X >= len(level.Grid) || 
       pos.Y < 0 || pos.Y >= len(level.Grid[0]) {
        return false
    }
    
    if level.Grid[pos.X][pos.Y].IsWall {
        return false
    }
    
    // Проверяем, не заблокирована ли позиция дверями
    for _, door := range level.Doors {
        if door.Position.X == pos.X && door.Position.Y == pos.Y && door.IsLocked {
            return false
        }
    }
    
    return true
}

func isNearDoor(keyPos, doorPos entity.Position) bool {
    // Проверяем, что ключ не слишком близко к двери (минимум 3 клетки расстояния)
    distance := utils.Abs(keyPos.X-doorPos.X) + utils.Abs(keyPos.Y-doorPos.Y)
    return distance <= 3
}

// validateKeyAccessibility проверяет, доступен ли ключ из стартовой позиции
func validateKeyAccessibility(level *entity.Level, state *entity.GameState, key *entity.Item, door *entity.Door) bool {
    // Простая проверка: ключ должен быть в той же связной компоненте, что и игрок
    // до открытия двери
    
    // Создаем временную копию уровня без этой двери
    tempDoors := make([]*entity.Door, 0)
    for _, d := range level.Doors {
        if d.Position.X != door.Position.X || d.Position.Y != door.Position.Y {
            tempDoors = append(tempDoors, d)
        }
    }
    
    // Проверяем, доступен ли ключ из стартовой позиции
    startPos := entity.Position{
        X: level.StartRoom.X + level.StartRoom.Width/2,
        Y: level.StartRoom.Y + level.StartRoom.Height/2,
    }
    
    return isPositionReachable(startPos, key.Position, level, tempDoors)
}

// isPositionReachable проверяет, можно ли дойти от start до end
func isPositionReachable(start, end entity.Position, level *entity.Level, doors []*entity.Door) bool {
    visited := make(map[entity.Position]bool)
    return dfsReachable(start, end, level, doors, visited)
}

func dfsReachable(current, end entity.Position, level *entity.Level, doors []*entity.Door, visited map[entity.Position]bool) bool {
    if current.X == end.X && current.Y == end.Y {
        return true
    }
    
    if visited[current] {
        return false
    }
    visited[current] = true
    
    // Проверяем соседние клетки
    directions := []entity.Position{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}}
    
    for _, dir := range directions {
        next := entity.Position{X: current.X + dir.X, Y: current.Y + dir.Y}
        
        // Проверяем границы
        if next.X < 0 || next.X >= len(level.Grid) || next.Y < 0 || next.Y >= len(level.Grid[0]) {
            continue
        }
        
        // Проверяем стену
        if level.Grid[next.X][next.Y].IsWall {
            continue
        }
        
        // Проверяем заблокированные двери
        doorBlocked := false
        for _, door := range doors {
            if door.Position.X == next.X && door.Position.Y == next.Y && door.IsLocked {
                doorBlocked = true
                break
            }
        }
        if doorBlocked {
            continue
        }
        
        if dfsReachable(next, end, level, doors, visited) {
            return true
        }
    }
    
    return false
}

// bfsSearch выполняет поиск в ширину для проверки доступности ключей
func bfsSearch(start *entity.GraphNode, targetKey entity.ItemSubType, state *entity.GameState) bool {
	queue := []*entity.GraphNode{start}
	visited := make(map[entity.Position]bool)
	visited[start.Position] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Если нашли ключ нужного типа - возвращаем true
		for _, item := range state.Items {
			if item.Position.X == current.Position.X &&
				item.Position.Y == current.Position.Y &&
				item.Type == entity.Scroll && // Ключи используют тип Scroll
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
func dfsSearch(start *entity.GraphNode, targetKey entity.ItemSubType, visited map[entity.Position]bool, state *entity.GameState) bool {
	if visited == nil {
		visited = make(map[entity.Position]bool)
	}

	visited[start.Position] = true

	// Проверяем, есть ли здесь нужный ключ
	for _, item := range state.Items {
		if item.Position.X == start.Position.X &&
			item.Position.Y == start.Position.Y &&
			item.Type == entity.Scroll &&
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
					if dfsSearch(neighbor, targetKey, visited, state) {
						return true
					}
				}
				continue
			}

			if dfsSearch(neighbor, targetKey, visited, state) {
				return true
			}
		}
	}

	return false
}

// buildAccessibilityGraph строит граф доступности уровня
func buildAccessibilityGraph(level *entity.Level) *entity.AccessibilityGraph {
	graph := &entity.AccessibilityGraph{Nodes: make([]*entity.GraphNode, 0)}
	nodeMap := make(map[entity.Position]*entity.GraphNode)

	// Создаем узлы для всех проходимых клеток
	for x := 0; x < len(level.Grid); x++ {
		for y := 0; y < len(level.Grid[0]); y++ {
			if !level.Grid[x][y].IsWall {
				pos := entity.Position{X: x, Y: y}
				node := &entity.GraphNode{
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
		directions := []entity.Position{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}}

		for _, dir := range directions {
			neighborPos := entity.Position{X: node.Position.X + dir.X, Y: node.Position.Y + dir.Y}
			if neighbor, exists := nodeMap[neighborPos]; exists {
				node.Neighbors = append(node.Neighbors, neighbor)
			}
		}
	}

	return graph
}

// В функции createEnemy исправим ошибку с enemy
func createEnemy(levelNum int, room *entity.Room) *entity.Enemy {
	// Сначала определим типы и веса
	types := []entity.EnemyType{entity.Zombie, entity.Vampire, entity.Ghost, entity.Ogre, entity.SnakeMage, entity.Mimic} // Уменьшили веса других, добавили Mimic
	weights := []int{35, 15, 15, 15, 10, 10}
	total := 0
	for _, w := range weights {
		total += w
	}
	roll := rand.Intn(total)
	var chosen entity.EnemyType
	for i, w := range weights {
		if roll < w {
			chosen = types[i]
			break
		}
		roll -= w
	}
	health := 4 + levelNum*2  // Меньше здоровья
	agility := 3 + levelNum   // Меньше ловкости
	strength := 2 + levelNum  // Меньше силы
	hostility := 5 + levelNum // Меньший радиус агрессии

	switch chosen {
	case entity.Zombie:
		health += 6
		agility -= 1
		strength += 1
		hostility += 2
	case entity.Vampire:
		health += 3
		agility += 2
		strength += 1
		hostility += 3
	case entity.Ghost:
		health += 1
		agility += 3
		strength -= 1
		hostility -= 2
	case entity.Ogre:
		health += 10
		agility -= 2
		strength += 3
		hostility += 1
	case entity.SnakeMage:
		health += 2
		agility += 2
		strength -= 1
		hostility += 4
	case entity.Mimic:
		health += 8
		agility += 3
		strength += 1
		hostility = 1 // Низкая враждебность
	}
	// Устанавливаем паттерн движения в зависимости от типа
	var pattern entity.MovementPattern
	switch chosen {
	case entity.Zombie:
		pattern = entity.Random
	case entity.Vampire:
		pattern = entity.Chase
	case entity.Ghost:
		pattern = entity.Teleport
	case entity.Ogre:
		pattern = entity.TwoStep
	case entity.SnakeMage:
		pattern = entity.Diagonal
	default:
		pattern = entity.Random
	}

	attempts := 0
	var pos entity.Position
	for attempts < 50 {
		pos = entity.Position{X: room.X + 1 + rand.Intn(max(1, room.Width-2)), Y: room.Y + 1 + rand.Intn(max(1, room.Height-2))}
		if pos.X >= room.X && pos.X < room.X+room.Width && pos.Y >= room.Y && pos.Y < room.Y+room.Height {
			break
		}
		attempts++
	}
	if attempts >= 50 {
		return nil
	}
	return &entity.Enemy{
		Type: chosen, Health: health, Agility: agility, Strength: strength, Hostility: hostility,
		Position: pos, IsVisible: true, Pattern: pattern, MoveCooldown: 0, HasBeenHit: false,
		AttackCooldown: 0, TeleportCooldown: 0, InvisibilityCooldown: 0, IsDisguised: true, // Mimic начинается замаскированным
	}
}

func CreateItem(levelNum int, room *entity.Room, state *entity.GameState, levelOpt ...*entity.Level) *entity.Item {
	var level *entity.Level
	if len(levelOpt) > 0 && levelOpt[0] != nil {
		level = levelOpt[0]
	} else {
		level = state.Level
	}

	if level == nil {
		return nil
	}

	itemTypes := []entity.ItemType{entity.Treasure, entity.Food, entity.Potion, entity.Scroll, entity.Weapon}
	weights := []int{30, 25, 15, 15, 15}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	roll := rand.Intn(totalWeight)
	var chosenType entity.ItemType
	var chosenSubType entity.ItemSubType

	for i, w := range weights {
		if roll < w {
			chosenType = itemTypes[i]
			break
		}
		roll -= w
	}

	switch chosenType {
	case entity.Treasure, entity.Food:
		chosenSubType = 0
	case entity.Potion:
		potionTypes := []entity.ItemSubType{entity.HealthPotion, entity.AgilityPotion, entity.StrengthPotion}
		chosenSubType = potionTypes[rand.Intn(len(potionTypes))]
	case entity.Scroll:
		scrollTypes := []entity.ItemSubType{entity.MaxHealthScroll, entity.AgilityScroll, entity.StrengthScroll}
		chosenSubType = scrollTypes[rand.Intn(len(scrollTypes))]
	case entity.Weapon:
		weaponTypes := []entity.ItemSubType{entity.Sword, entity.Axe}
		chosenSubType = weaponTypes[rand.Intn(len(weaponTypes))]
	}

	healthBoost := 5 + levelNum*2
	maxHealthBoost := 3 + levelNum
	agilityBoost := 1 + levelNum
	strengthBoost := 2 + levelNum
	value := 10 + levelNum*5

	var pos entity.Position
	attempts := 0
	found := false

	for attempts < 50 {
		if room != nil {
			pos = entity.Position{
				X: room.X + 1 + rand.Intn(max(1, room.Width-2)),
				Y: room.Y + 1 + rand.Intn(max(1, room.Height-2)),
			}
		} else {
			pos = entity.Position{
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

	return &entity.Item{
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
