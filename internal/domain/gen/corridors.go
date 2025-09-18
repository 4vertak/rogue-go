package gen

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"rogue-go/internal/domain/entity"
)




// теперь пороб


//структура для  нахождания компонентов свзяности и для построения MST

//DSU представляет собой набор корневых деревьев (лес). 
//Каждое дерево соответствует определённому множеству. 
//При реализации DSU необходимо лишь подниматься вверх по деревьям,
// поэтому достаточно хранить для каждой вершины только номер её прямого предка.


type DSU struct {
	parent []int
}

//Для описания множества используется номер вершины, являющейся корнем соответствующего дерева.

func NewDSU(n int) *DSU {
	p := make([]int, n)

	for i := range p {
		p[i] = i
	}
	return &DSU{parent: p}
}


//Поэтому для определения, принадлежат ли два элемента к одному и тому же множеству, 
//для каждого элемента нужно найти корень соответствующего дерева (поднимаясь вверх пока это возможно) и сравнить эти корни.
func (d *DSU) Connected(x, y int) bool {
	return d.Find(x) == d.Find(y)
}

func (d *DSU) Find(x int) int {
	if d.parent[x] != x {
		d.parent[x] = d.Find(d.parent[x])
	}
	return d.parent[x]
}


//Объединяет два множества, к которым принадлежат элементы x и y, в одно новое множество.
func (d *DSU) Union(x, y int){
	d.parent[d.Find(x)] = d.Find(y)
} 


//MST, алгоритм Крускала - берет список ребр и возвращает те что попали в остов
func buildMST(edges []entity.Edge, n int) []entity.Edge {

	//сортируем ребра по весу
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Weight < edges[j].Weight
	})

	//DSU используется для отслеживания уже объединенных компонент графа.
	dsu := NewDSU(n)

	var mst []entity.Edge

	for _, edge := range edges {
		//DSU Oпределяет, принадлежат ли концы этого ребра к уже существующим компонентам, 
		//и если нет, то добавляет ребро и объединяет компоненты.
		if !dsu.Connected(edge.U, edge.U) {
			dsu.Union(edge.U, edge.V)
			mst = append(mst, edge)
		}
	}

	return mst
}

//строим ребра 
func buildEdges(rooms []*entity.Room) []entity.Edge {
	roomMap := make(map[string]int)

	for i, room := range rooms {
		key := fmt.Sprintf("%d,%d", room.GridX, room.GridY)
		roomMap[key] = i
	}
	var edges []entity.Edge

	for i, room := range rooms {
		//если справа сосед
		if j, ok := roomMap[fmt.Sprintf("%d,%d", room.GridX+1, room.GridY)]; ok {
			edges = append(edges, entity.Edge{U: i, V: j, Weight: distanceBetweenRooms(room, rooms[j])})
		}
		// елси сосед снизу
		if j, exists := roomMap[fmt.Sprintf("%d,%d", room.GridX, room.GridY+1)]; exists {
			edges = append(edges, entity.Edge{U: i, V: j, Weight: distanceBetweenRooms(room, rooms[j])})
		}
	}

	return edges
}

func ensuringConnectivity(level *entity.Level, dsu *DSU){
	for {
		comps := make(map[int][]int)
		for i := range level.Rooms {
			root := dsu.Find(i)
			comps[root] = append(comps[root], i)
		}
		if len(comps) <= 1 {
			break
		}
		
		//соединяем два компонента  ближайшей парой
		var bestEdge *entity.Edge
		bestDist := math.MaxFloat64
		
		roots := make([]int, 0, len(comps))
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
			dsu.Union(bestEdge.U, bestEdge.V)
			connectTwoRooms(level, level.Rooms[bestEdge.U], level.Rooms[bestEdge.V])
		}
	}

}



func connectRoomsWithCorridors(level *entity.Level) {
	if len(level.Rooms) <= 1 {
		return
	}

	edges := buildEdges(level.Rooms)
	mstEdges := buildMST(edges, len(level.Rooms))
	dsu := NewDSU(len(level.Rooms))


   	 // В строим MST
   	 for _, edge := range mstEdges {
            if !hasCorridorBetween(level, level.Rooms[edge.U], level.Rooms[edge.V]) {
                dsu.Union(edge.U, edge.V)
                connectTwoRooms(level, level.Rooms[edge.U], level.Rooms[edge.V])
            }
        }


	// Обспечиваем связность - елис нет связи то соединяем комнаты
	ensuringConnectivity(level, dsu)

	// рандомно доп корридоры
	extraConnections := int(math.Sqrt(float64(len(level.Rooms))))
	for i := 0; i < extraConnections && i < len(edges); i++ {
		randomEdge := edges[rand.Intn(len(edges))]
		if !dsu.Connected(randomEdge.U, randomEdge.V) {
			dsu.Union(randomEdge.U, randomEdge.V)
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

