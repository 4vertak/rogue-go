package gen

import (
	"math"
	"sort"
	"math/rand"
	"time"

	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/rules"
)

type Config struct{ CellW, CellH, MinRW, MinRH int }

func DefaultConfig() Config { return Config{CellW: 10, CellH: 6, MinRW: 4, MinRH: 3} }

func RNG(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }
func NowRNG() *rand.Rand        { return RNG(time.Now().UnixNano()) }

func BuildLevel(rng *rand.Rand, index int, cfg Config) entity.Level {
	W := cfg.CellW*3 + 1
	H := cfg.CellH*3 + 1
	tiles := make([][]entity.Tile, H)
	for y := 0; y < H; y++ {
		tiles[y] = make([]entity.Tile, W)
		for x := 0; x < W; x++ {
			tiles[y][x] = entity.Wall
		}
	}

	rooms := []entity.Room{}
	generateRoom(rng, tiles, W, H, cfg, &rooms)

	// Соединение соседних комнат
	generatePassage(rng, tiles, rooms)

	// Лестница вниз
	var exit entity.Pos

	if len(rooms) > 0 {
		for {
			startIdxRoom := rng.Intn(len(rooms)) 
			if ! rooms[startIdxRoom].IsGone {
				stair := rooms[startIdxRoom]
				exit = entity.Pos{X: stair.X + stair.W/2, Y: stair.Y + stair.H/2}
				tiles[exit.Y][exit.X] = entity.Exit
				break
			}
		}
	}

	// Монстры
	mobs := []entity.Monster{}
	for _, rm := range rooms {
		if rng.Intn(2) == 0  && !rm.IsGone {
			mobs = append(mobs, entity.Monster{
				Pos:       entity.Pos{X: rm.X + rng.Intn(rm.W), Y: rm.Y +rng.Intn(rm.H)},
				Stats:     entity.Stats{HP: 5 + index, MaxHP: 5 + index, STR: 3, DEX: 3},
				Type:      "zombie",
				Hostility: 5,
				Symbol:    'z',
			})
		}
	}

	// Предметы
	items := []entity.Item{}
	for _, rm := range rooms {
		if rng.Intn(3) == 0 && !rm.IsGone {
			for {
				ix := rm.X + rng.Intn(rm.W)
				iy := rm.Y + rng.Intn(rm.H)
				if  exit.X != ix && exit.Y != iy {
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
		Index: index,
		W:     W,
		H:     H,
		Tiles: tiles,
		Rooms: rooms,
		Exit:  exit,
		Mobs:  mobs,
		Items: items,
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
 // Собираем информацию о комнатах
 type RoomInfo struct {
  Index    int
  GX, GY   int
  Room     entity.Room
 }
 var validRooms []RoomInfo

 for i, room := range rooms {
  if !room.IsGone {
   gx := i % 3
   gy := i / 3
   validRooms = append(validRooms, RoomInfo{
    Index: i,
    GX:    gx,
    GY:    gy,
    Room:  room,
   })
  }
 }

 if len(validRooms) <= 1 {
  return
 }

 // Создаем ребра между соседними комнатами
 type Edge struct {
  U, V    int
  Weight  float64
 }
 var edges []Edge

 for i := 0; i < len(validRooms); i++ {
  for j := i + 1; j < len(validRooms); j++ {
   ri, rj := validRooms[i], validRooms[j]
   dx := rules.Abs(ri.GX - rj.GX)
   dy := rules.Abs(ri.GY - rj.GY)
  
   // Соединяем только соседние комнаты в сетке
   if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
    ciX, ciY := ri.Room.X+ri.Room.W/2, ri.Room.Y+ri.Room.H/2
    cjX, cjY := rj.Room.X+rj.Room.W/2, rj.Room.Y+rj.Room.H/2
    weight := math.Sqrt(float64((ciX-cjX)*(ciX-cjX) + (ciY-cjY)*(ciY-cjY)))
    edges = append(edges, Edge{U: i, V: j, Weight: weight})
   }
  }
 }

 // Алгоритм Крускала
 sort.Slice(edges, func(i, j int) bool {
  return edges[i].Weight < edges[j].Weight
 })

 parent := make([]int, len(validRooms))
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

 var mstEdges []Edge
 for _, e := range edges {
  if find(e.U) != find(e.V) {
   mstEdges = append(mstEdges, e)
   union(e.U, e.V)
  }
 }

 // Добавляем случайные соединения (20% вероятность)
 for _, e := range edges {
  if rng.Float64() < 0.2 {
   // Проверяем, нет ли уже этого соединения в MST
   found := false
   for _, me := range mstEdges {
    if (me.U == e.U && me.V == e.V) || (me.U == e.V && me.V == e.U) {
     found = true
     break
    }
   }
   if !found {
    mstEdges = append(mstEdges, e)
   }
  }
 }

 // Создаем коридоры для выбранных ребер
 for _, e := range mstEdges {
  roomA := validRooms[e.U].Room
  roomB := validRooms[e.V].Room
  connectRooms(rng, tiles, roomA, roomB)
 }
}


func connectRooms(rng *rand.Rand, tiles [][]entity.Tile, a, b entity.Room) {
	if a.W == 0 || b.W == 0 {
		return
	}
	ax, ay := a.X+a.W/2, a.Y+a.H/2
	bx, by := b.X+b.W/2, b.Y+b.H/2
	if rng.Intn(2) == 0 {
		carveH(tiles, ax, bx, ay)
		carveV(tiles, ay, by, bx)
	} else {
		carveV(tiles, ay, by, ax)
		carveH(tiles, ax, bx, by)
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
