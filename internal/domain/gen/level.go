package gen

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/rules"
)

type Config struct{ CellW, CellH, MinRW, MinRH int }

func DefaultConfig() Config { return Config{CellW: 24, CellH: 8, MinRW: 4, MinRH: 3} }

func RNG(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }
func NowRNG() *rand.Rand        { return RNG(time.Now().UnixNano()) }

func BuildLevel(rng *rand.Rand, index int, cfg Config) entity.Level {
	W := cfg.CellW*3 + 1
	H := cfg.CellH*3 + 1

	tiles := make([][]entity.Tile, H)
	for y := range H {
		tiles[y] = make([]entity.Tile, W)
		for x := range W {
			tiles[y][x] = entity.Wall
		}
	}

	explored := make([][]bool, H) //инициализация непросмотренные клетки
	for y := range explored {
		explored[y] = make([]bool, W)
	}

	rooms := []entity.Room{}
	generateRoom(rng, tiles, W, H, cfg, &rooms)

	generatePassage(rng, tiles, rooms)

	var validRooms []entity.Room
	for _, room := range rooms {
		if !room.IsGone {
			validRooms = append(validRooms, room)
		}
	}

	// Лестница вниз
	var exit entity.Pos

	exitRoom := validRooms[rng.Intn(len(validRooms))]
	exit = entity.Pos{X: exitRoom.X + exitRoom.W/2, Y: exitRoom.Y + exitRoom.H/2}
	tiles[exit.Y][exit.X] = entity.Exit

	var playerStartPos entity.Pos
	if len(validRooms) > 1 {
		var startRooms []entity.Room
		for _, room := range validRooms {
			if room.X != exitRoom.X || room.Y != exitRoom.Y {
				startRooms = append(startRooms, room)
			}
		}
		startRoom := startRooms[rng.Intn(len(startRooms))]
		playerStartPos = entity.Pos{
			X: startRoom.X + startRoom.W/2,
			Y: startRoom.Y + startRoom.H/2,
		}
	} else {
		playerStartPos = entity.Pos{
			X: validRooms[0].X + validRooms[0].W/2,
			Y: validRooms[0].Y + validRooms[0].H/2,
		}
	}

	// Монстры
	mobs := []entity.Monster{}
	for _, rm := range validRooms {
		if rng.Intn(2) == 0 {
			mobs = append(mobs, entity.Monster{
				Pos:       entity.Pos{X: rm.X + rng.Intn(rm.W), Y: rm.Y + rng.Intn(rm.H)},
				Stats:     entity.Stats{HP: 5 + index, MaxHP: 5 + index, STR: 3, DEX: 3},
				Type:      "zombie",
				Hostility: 5,
				Symbol:    'z',
			})
		}
	}

	// Предметы
	items := []entity.Item{}
	for _, rm := range validRooms {
		if rng.Intn(3) == 0 {
			for {
				ix := rm.X + rng.Intn(rm.W)
				iy := rm.Y + rng.Intn(rm.H)
				if exit.X != ix && exit.Y != iy {
					items = append(items, entity.Item{
						Type:   "Food",
						Health: 5,
						Pos:    entity.Pos{X: ix, Y: iy},
					})
					break
				}
			}
		}
	}

	return entity.Level{
		Index:          index,
		W:              W,
		H:              H,
		Tiles:          tiles,
		Explored:       explored,
		Rooms:          rooms,
		PlayerStartPos: playerStartPos,
		Exit:           exit,
		Mobs:           mobs,
		Items:          items,
	}
}

func generateRoom(rng *rand.Rand, tiles [][]entity.Tile, W int, H int, cfg Config, rooms *[]entity.Room) {

	for gy := 0; gy < 3; gy++ {
		for gx := 0; gx < 3; gx++ {
			cx, cy := gx*cfg.CellW, gy*cfg.CellH

			maxRW := max(cfg.MinRW, cfg.CellW-2)
			maxRH := max(cfg.MinRH, cfg.CellH-2)

			rw := cfg.MinRW
			if maxRW > cfg.MinRW {
				rw += rng.Intn(maxRW - cfg.MinRW + 1)
			}
			rh := cfg.MinRH
			if maxRH > cfg.MinRH {
				rh += rng.Intn(maxRH - cfg.MinRH + 1)
			}

			maxX := cfg.CellW - rw - 1
			if maxX < 1 {
				maxX = 1
			}
			maxY := cfg.CellH - rh - 1
			if maxY < 1 {
				maxY = 1
			}

			rx := cx + 1 + rng.Intn(maxX)
			ry := cy + 1 + rng.Intn(maxY)

			isGone := rng.Intn(100) >= 80

			room := entity.Room{X: rx, Y: ry, W: rw, H: rh, IsGone: isGone}
			*rooms = append(*rooms, room)

			if !isGone {
				for y := ry; y < ry+rh && y < H; y++ {
					for x := rx; x < rx+rw && x < W; x++ {
						tiles[y][x] = entity.Floor
					}
				}
			}
		}
	}
}

func generatePassage(rng *rand.Rand, tiles [][]entity.Tile, rooms []entity.Room) {
	type RoomInfo struct {
		Index  int
		GX, GY int
		Room   *entity.Room
	}

	var allRooms []RoomInfo
	for i := range rooms {
		if rooms[i].IsGone {
			continue
		}
		gx := i % 3
		gy := i / 3
		allRooms = append(allRooms, RoomInfo{
			Index: i, GX: gx, GY: gy, Room: &rooms[i],
		})
	}

	if len(allRooms) <= 1 {
		return
	}

	type Edge struct {
		U, V   int
		Weight float64
	}

	// рёбра между соседними комнатами
	var edges []Edge
	for i := 0; i < len(allRooms); i++ {
		for j := i + 1; j < len(allRooms); j++ {
			ri, rj := allRooms[i], allRooms[j]
			dx := rules.Abs(ri.GX - rj.GX)
			dy := rules.Abs(ri.GY - rj.GY)
			if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
				ciX, ciY := ri.Room.X+ri.Room.W/2, ri.Room.Y+ri.Room.H/2
				cjX, cjY := rj.Room.X+rj.Room.W/2, rj.Room.Y+rj.Room.H/2
				weight := math.Sqrt(float64((ciX-cjX)*(ciX-cjX) + (ciY-cjY)*(ciY-cjY)))
				edges = append(edges, Edge{U: i, V: j, Weight: weight})
			}
		}
	}

	// сортировка рёбер по dtce
	sort.Slice(edges, func(i, j int) bool { return edges[i].Weight < edges[j].Weight })

	// DSU
	parent := make([]int, len(allRooms))
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
	union := func(x, y int) { parent[find(x)] = find(y) }

	var mstEdges []Edge
	for _, e := range edges {
		if find(e.U) != find(e.V) {
			mstEdges = append(mstEdges, e)
			union(e.U, e.V)
		}
	}

	// проверяем связность и  соединяем оставшиеся комнаты
	for {
		comps := make(map[int][]int)
		for i := range allRooms {
			root := find(i)
			comps[root] = append(comps[root], i)
		}
		if len(comps) <= 1 {
			break
		}

		// находим ближайшие комнаты
		var bestEdge *Edge
		bestDist := math.MaxFloat64
		var roots []int
		for k := range comps {
			roots = append(roots, k)
		}
		for _, a := range comps[roots[0]] {
			for _, b := range comps[roots[1]] {
				ax, ay := allRooms[a].Room.X+allRooms[a].Room.W/2, allRooms[a].Room.Y+allRooms[a].Room.H/2
				bx, by := allRooms[b].Room.X+allRooms[b].Room.W/2, allRooms[b].Room.Y+allRooms[b].Room.H/2
				dist := math.Sqrt(float64((ax-bx)*(ax-bx) + (ay-by)*(ay-by)))
				if dist < bestDist {
					bestDist = dist
					bestEdge = &Edge{U: a, V: b, Weight: dist}
				}
			}
		}

		if bestEdge != nil {
			mstEdges = append(mstEdges, *bestEdge)
			union(bestEdge.U, bestEdge.V)
		}
	}

	for _, e := range mstEdges {
		connectRooms(rng, tiles, *allRooms[e.U].Room, *allRooms[e.V].Room)
	}
}

func randomDoor(rng *rand.Rand, room entity.Room) (int, int) {
	if room.W < 6 || room.H < 6 {
		return room.X + room.W/2, room.Y + room.H/2
	}

	side := rng.Intn(4)
	var doorX, doorY int

	minX := room.X + 2
	maxX := room.X + room.W - 3
	minY := room.Y + 2
	maxY := room.Y + room.H - 3

	switch side {
	case 0: // верх
		doorX = minX + rng.Intn(max(1, maxX-minX+1))
		doorY = room.Y
	case 1: // низ
		doorX = minX + rng.Intn(max(1, maxX-minX+1))
		doorY = room.Y + room.H - 1
	case 2: // левая
		doorY = minY + rng.Intn(max(1, maxY-minY+1))
		doorX = room.X
	case 3: // правая
		doorY = minY + rng.Intn(max(1, maxY-minY+1))
		doorX = room.X + room.W - 1
	}
	return doorX, doorY
}

func connectRooms(rng *rand.Rand, tiles [][]entity.Tile, a, b entity.Room) {
	if a.W == 0 || b.W == 0 {
		return
	}

	aDoorX, aDoorY := randomDoor(rng, a)
	bDoorX, bDoorY := randomDoor(rng, b)

	tiles[aDoorY][aDoorX] = entity.Door
	tiles[bDoorY][bDoorX] = entity.Door

	if rng.Intn(2) == 0 {
		carveH(tiles, aDoorX, bDoorX, aDoorY)
		carveV(tiles, aDoorY, bDoorY, bDoorX)
	} else {
		carveV(tiles, aDoorY, bDoorY, aDoorX)
		carveH(tiles, aDoorX, bDoorX, bDoorY)
	}
}

func carveH(tiles [][]entity.Tile, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		tiles[y][x] = entity.Floor
	}
}
func carveV(tiles [][]entity.Tile, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		tiles[y][x] = entity.Floor
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
