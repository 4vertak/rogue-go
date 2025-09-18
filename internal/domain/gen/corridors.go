package gen

import (
	"math"
	"math/rand"
	"sort"
	"sync" //экперименты с многопочтность)))

	"rogue-go/internal/domain/entity"
)

//позиция комнаты в сетке кораты
type GridPos struct {
	X, Y int
}


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

// будет планировщиком соединений
type CorridorPlaner struct {
	level *entity.Level
	dsu   *DSU
	edges []entity.Edge
}

func NewCorridorPlaner(level *entity.Level) *CorridorPlaner {
	return &CorridorPlaner{
		level: level,
		dsu:   NewDSU(len(level.Rooms)),
		edges: buildEdges(level.Rooms),
	}
}


//строим ребра  -доавил парарллельный расет  ребер между соедями
func buildEdges(rooms []*entity.Room) []entity.Edge {
	roomMap := make(map[GridPos]int)
	for i, room := range rooms {
		roomMap[GridPos{room.GridX, room.GridY}] = i
	}

	results := make(chan entity.Edge, len(rooms) * 2) //2 cоседа справа и снизу
	var wg sync.WaitGroup

	for i, room := range rooms {
		//если справа сосед
		if j, ok := roomMap[GridPos{room.GridX+1, room.GridY}]; ok {
			wg.Add(1)
			go func(a, b int) {
				defer wg.Done()
				d := distanceBetweenRooms(rooms[a], rooms[b]) 
				results <- entity.Edge{U: a, V: b, Weight: d}
			}(i, j)
		}
		// елси сосед снизу
		if j, ok := roomMap[GridPos{ room.GridX, room.GridY+1}]; ok {
			wg.Add(1)
			go func(a, b int) {
				defer wg.Done()
				d := distanceBetweenRooms(rooms[a], rooms[b]) 
				results <- entity.Edge{U: a, V: b, Weight: d}
			}(i, j)
		}
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	var edges []entity.Edge
	for edge := range results {
		edges = append(edges, edge)
	}


	return edges
}



//строим MST
func (p *CorridorPlaner) BuildMST() {

	//сортируем ребра по весу
	sort.Slice(p.edges, func(i, j int) bool {
		return p.edges[i].Weight < p.edges[j].Weight
	})

	for _, edge := range p.edges {
		//DSU Oпределяет, принадлежат ли концы этого ребра к уже существующим компонентам, 
		//и если нет, то добавляет ребро и объединяет компоненты.
		if !p.dsu.Connected(edge.U, edge.U) && !hasCorridorBetween(p.level, p.level.Rooms[edge.U], p.level.Rooms[edge.V]) {
			p.dsu.Union(edge.U, edge.V)
			connectTwoRooms(p.level, p.level.Rooms[edge.U], p.level.Rooms[edge.V])
		}
	}
}


func (p *CorridorPlaner) EnsureConnectivity(){
	for {
		comps := make(map[int][]int)
		for i := range p.level.Rooms {
			comps[p.dsu.Find(i)] = append(comps[p.dsu.Find(i)], i)
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
				dist := distanceBetweenRooms(p.level.Rooms[a], p.level.Rooms[b])
				if dist < bestDist {
					bestDist = dist
					bestEdge = &entity.Edge{U: a, V: b, Weight: dist}
				}
			}
		}

		if bestEdge != nil {
			p.dsu.Union(bestEdge.U, bestEdge.V)
			connectTwoRooms(p.level, p.level.Rooms[bestEdge.U], p.level.Rooms[bestEdge.V])
		}
	}

}

// сдучайые доплнительные корриодоры

func (p *CorridorPlaner) AddExtraCorridor(){
	// рандомно доп корридоры
	extraConnections := int(math.Sqrt(float64(len(p.level.Rooms))))

	if extraConnections == 0 || len(p.edges) == 0 {
		return
	}
	//перемешиваем ребра
	rand.Shuffle(len(p.edges), func(i, j int) { p.edges[i], p.edges[j] = p.edges[j], p.edges[i]})

	for i := 0; i < extraConnections && i < len(p.edges); i++ {
		randomEdge := p.edges[i]
		if !p.dsu.Connected(randomEdge.U, randomEdge.V) && !hasCorridorBetween(p.level, p.level.Rooms[randomEdge.U], p.level.Rooms[randomEdge.V]){
			p.dsu.Union(randomEdge.U, randomEdge.V)
			connectTwoRooms(p.level, p.level.Rooms[randomEdge.U], p.level.Rooms[randomEdge.V])
		}
	}
}



func connectRoomsWithCorridors(level *entity.Level) {
	if len(level.Rooms) <= 1 {
		return
	}
	planner := NewCorridorPlaner(level)
	planner.BuildMST()
	planner.EnsureConnectivity()
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

