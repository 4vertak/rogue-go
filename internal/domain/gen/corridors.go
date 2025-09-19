package gen

import (
	"math"
	"math/rand"
	"sort"
	"sync" //экперименты с многопочтность)))

	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/utils"
)

// позиция комнаты в сетке кораты
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

// Поэтому для определения, принадлежат ли два элемента к одному и тому же множеству,
// для каждого элемента нужно найти корень соответствующего дерева (поднимаясь вверх пока это возможно) и сравнить эти корни.
func (d *DSU) Connected(x, y int) bool {
	return d.Find(x) == d.Find(y)
}

func (d *DSU) Find(x int) int {
	if d.parent[x] != x {
		d.parent[x] = d.Find(d.parent[x])
	}
	return d.parent[x]
}

// Объединяет два множества, к которым принадлежат элементы x и y, в одно новое множество.
func (d *DSU) Union(x, y int) {
	d.parent[d.Find(x)] = d.Find(y)
}

type Exit struct {
	X, Y int
	Side string
}

type CorridorContext struct {
	exits map[*entity.Room][]Exit
}

func NewCorridorContext() *CorridorContext {
	return &CorridorContext{
		exits: make(map[*entity.Room][]Exit),
	}
}

// Локальная версия randomExit
func (ctx *CorridorContext) randomExit(room *entity.Room, preferredSide string) (int, int, bool) {
	// Проверка — есть ли уже выход на этой стороне
	sides := []string{"left", "right", "top", "bottom"}

	for i, side := range sides {
		if side == preferredSide {
			sides[0], sides[i] = sides[i], sides[0]
			break
		}
	}

	for _, side := range sides {
		already := false
		for _, e := range ctx.exits[room] {
			if e.Side == side {
				already = true
				break
			}
		}
		if already {
			continue
		}

		var x, y int
		switch side {
		case "left":
			x = room.X
			y = rand.Intn(room.Height-2) + room.Y + 1
		case "right":
			x = room.X + room.Width - 1
			y = rand.Intn(room.Height-2) + room.Y + 1
		case "top":
			y = room.Y
			x = rand.Intn(room.Width-2) + room.X + 1
		case "bottom":
			y = room.Y + room.Height - 1
			x = rand.Intn(room.Width-2) + room.X + 1
		}

		exit := Exit{X: x, Y: y, Side: side}
		ctx.exits[room] = append(ctx.exits[room], exit)
		return x, y, true
	}

	return 0, 0, false
}

// будет планировщиком соединений
type CorridorPlaner struct {
	level *entity.Level
	dsu   *DSU
	edges []entity.Edge
	ctx   *CorridorContext
}

func NewCorridorPlaner(level *entity.Level) *CorridorPlaner {
	return &CorridorPlaner{
		level: level,
		dsu:   NewDSU(len(level.Rooms)),
		edges: buildEdges(level.Rooms),
		ctx:   NewCorridorContext(),
	}
}

// строим ребра  -доавил парарллельный расет  ребер между соедями
func buildEdges(rooms []*entity.Room) []entity.Edge {
	roomMap := make(map[GridPos]int)
	for i, room := range rooms {
		roomMap[GridPos{room.GridX, room.GridY}] = i
	}

	results := make(chan entity.Edge, len(rooms)*2) //2 cоседа справа и снизу
	var wg sync.WaitGroup

	for i, room := range rooms {
		//если справа сосед
		if j, ok := roomMap[GridPos{room.GridX + 1, room.GridY}]; ok {
			wg.Add(1)
			go func(a, b int) {
				defer wg.Done()
				d := distanceBetweenRooms(rooms[a], rooms[b])
				results <- entity.Edge{U: a, V: b, Weight: d}
			}(i, j)
		}
		// елси сосед снизу
		if j, ok := roomMap[GridPos{room.GridX, room.GridY + 1}]; ok {
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

// строим MST
func (p *CorridorPlaner) BuildMST() {

	//сортируем ребра по весу
	sort.Slice(p.edges, func(i, j int) bool {
		return p.edges[i].Weight < p.edges[j].Weight
	})

	for _, edge := range p.edges {
		//DSU Oпределяет, принадлежат ли концы этого ребра к уже существующим компонентам,
		//и если нет, то добавляет ребро и объединяет компоненты.
		if !p.dsu.Connected(edge.U, edge.V) && !hasCorridorBetween(p.level, edge.U, edge.V) {
			p.dsu.Union(edge.U, edge.V)
			connectTwoRooms(p.level, p.ctx, edge.U, edge.V)
		}
	}
}

func (p *CorridorPlaner) EnsureConnectivity() {
	for {
		comps := make(map[int][]int)
		for i := range p.level.Rooms {
			root := p.dsu.Find(i)
			comps[root] = append(comps[root], i)
		}

		if len(comps) <= 1 {
			break
		}

		roots := make([]int, 0, len(comps))
		for k := range comps {
			roots = append(roots, k)
		}

		for i := 0; i < len(roots)-1; i++ {
			compA := comps[roots[i]]
			bestDist := math.MaxFloat64
			var bestU, bestV int

			for _, u := range compA {
				for _, v := range comps[roots[i+1]] {
					dist := distanceBetweenRooms(p.level.Rooms[u], p.level.Rooms[v])
					if dist < bestDist {
						bestDist = dist
						bestU, bestV = u, v
					}
				}
			}

			if !p.dsu.Connected(bestU, bestV) {
				p.dsu.Union(bestU, bestV)
				connectTwoRooms(p.level, p.ctx, bestU, bestV)
			}
		}
	}
}

// сдучайые доплнительные корриодоры
func (p *CorridorPlaner) AddExtraCorridor() {
	extraConnections := int(math.Sqrt(float64(len(p.level.Rooms))))

	if extraConnections == 0 || len(p.edges) == 0 {
		return
	}

	rand.Shuffle(len(p.edges), func(i, j int) { p.edges[i], p.edges[j] = p.edges[j], p.edges[i] })

	added := 0

	for _, e := range p.edges {
		if hasCorridorBetween(p.level, e.U, e.V) {
			continue
		}
		connectTwoRooms(p.level, p.ctx, e.U, e.V)
		added++
		if added >= extraConnections {
			break
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

func connectTwoRooms(level *entity.Level, ctx *CorridorContext, u, v int) {
	room1 := level.Rooms[u]
	room2 := level.Rooms[v]

	if hasCorridorBetween(level, u, v) {
		return
	}

	dx := (room2.X + room2.Width/2) - (room1.X + room1.Width/2)
	dy := (room2.Y + room2.Height/2) - (room1.Y + room1.Height/2)

	var x1, y1, x2, y2 int
	var ok1, ok2 bool

	if utils.Abs(dx) > utils.Abs(dy) {
		if dx > 0 {
			x1, y1, ok1 = ctx.randomExit(room1, "right")
			x2, y2, ok2 = ctx.randomExit(room2, "left")
		} else {
			x1, y1, ok1 = ctx.randomExit(room1, "left")
			x2, y2, ok2 = ctx.randomExit(room2, "right")
		}
	} else {
		if dy > 0 {
			x1, y1, ok1 = ctx.randomExit(room1, "bottom")
			x2, y2, ok2 = ctx.randomExit(room2, "top")
		} else {
			x1, y1, ok1 = ctx.randomExit(room1, "top")
			x2, y2, ok2 = ctx.randomExit(room2, "bottom")
		}
	}

	if !ok1 || !ok2 {
		return
	}

	if rand.Intn(2) == 0 {
		createHorizontalCorridor(level, x1, y1, x2)
		createVerticalCorridor(level, x2, y1, y2)
	} else {
		createVerticalCorridor(level, x1, y1, y2)
		createHorizontalCorridor(level, x1, y2, x2)
	}

	corridor := &entity.Corridor{
		Start: entity.Position{X: x1, Y: y1},
		End:   entity.Position{X: x2, Y: y2},
		U:     u,
		V:     v,
		Width: 1,
	}
	level.Corridors = append(level.Corridors, corridor)
}

// Добавить функцию проверки существующего коридора -СДЕЛАНО
func hasCorridorBetween(level *entity.Level, u, v int) bool {
	for _, corridor := range level.Corridors {
		if (corridor.U == u && corridor.V == v) || (corridor.U == v && corridor.V == u) {
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
