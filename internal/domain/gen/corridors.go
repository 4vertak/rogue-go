package gen

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"rogue-go/internal/domain/entity"
)



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

    // В цикле MST добавляем проверку
    for _, edge := range edges {
        if find(edge.U) != find(edge.V) {
            // Проверяем, что комнаты еще не соединены
            room1 := level.Rooms[edge.U]  // Это *entity.Room
            room2 := level.Rooms[edge.V]  // Это *entity.Room
            
            if !hasCorridorBetween(level, room1, room2) {
                union(edge.U, edge.V)
                connectTwoRooms(level, room1, room2)  // Теперь передаем *entity.Room
            }
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

    // Проверяем, есть ли уже коридор между этими комнатами
    if hasCorridorBetween(level, room1, room2) {
        return
    }

    // Выбираем только один тип соединения вместо двух
    if rand.Intn(2) == 0 {
        createHorizontalCorridor(level, x1, y1, x2)
        createVerticalCorridor(level, x2, y1, y2)
    } else {
        createVerticalCorridor(level, x1, y1, y2)
        createHorizontalCorridor(level, x1, y2, x2)
    }

    // Добавляем коридор в список
    corridor := &entity.Corridor{
        Start: entity.Position{X: x1, Y: y1},
        End:   entity.Position{X: x2, Y: y2},
    }
    level.Corridors = append(level.Corridors, corridor)
}

// Добавить функцию проверки существующего коридора
func hasCorridorBetween(level *entity.Level, room1, room2 *entity.Room) bool {
    for _, corridor := range level.Corridors {
        room1Center := entity.Position{
            X: room1.X + room1.Width/2,
            Y: room1.Y + room1.Height/2,
        }
        room2Center := entity.Position{
            X: room2.X + room2.Width/2,
            Y: room2.Y + room2.Height/2,
        }
        
        if (corridor.Start == room1Center && corridor.End == room2Center) ||
           (corridor.Start == room2Center && corridor.End == room1Center) {
            return true
        }
    }
    return false
}

func createHorizontalCorridor(level *entity.Level, x1, y1, x2 int) {
    start, end := x1, x2
    if start > end {
        start, end = end, start
    }
    
    // Создаем коридор шириной 1 клетка
    for x := start; x <= end; x++ {
        if x >= 0 && x < len(level.Grid) && y1 >= 0 && y1 < len(level.Grid[0]) {
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
    
    // Создаем коридор шириной 1 клетка
    for y := start; y <= end; y++ {
        if x1 >= 0 && x1 < len(level.Grid) && y >= 0 && y < len(level.Grid[0]) {
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

