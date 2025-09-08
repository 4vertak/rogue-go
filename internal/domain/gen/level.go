package gen

import (
	"math/rand"
	"time"

	"github.com/4vertak/rogue-go/internal/domain/entity"
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
	// Генерация комнат (80% шанс)
	for gy := 0; gy < 3; gy++ {
		for gx := 0; gx < 3; gx++ {
			if rng.Intn(100) >= 80 {
				continue
			}
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

			room := entity.Room{X: rx, Y: ry, W: rw, H: rh}
			rooms = append(rooms, room)

			for y := ry; y < ry+rh && y < H; y++ {
				for x := rx; x < rx+rw && x < W; x++ {
					tiles[y][x] = entity.Floor
				}
			}
		}
	}

	// Соединение соседних комнат
	for gy := 0; gy < 3; gy++ {
		for gx := 0; gx < 3; gx++ {
			idx := gy*3 + gx
			if idx >= len(rooms) {
				continue
			}
			if gx < 2 && idx+1 < len(rooms) {
				connectRooms(rng, tiles, rooms[idx], rooms[idx+1])
			}
			if gy < 2 && idx+3 < len(rooms) {
				connectRooms(rng, tiles, rooms[idx], rooms[idx+3])
			}
		}
	}

	// Лестница вниз
	var exit entity.Pos
	if len(rooms) > 0 {
		stair := rooms[rng.Intn(len(rooms))]
		exit = entity.Pos{X: stair.X + stair.W/2, Y: stair.Y + stair.H/2}
		tiles[exit.Y][exit.X] = entity.Exit
	}

	// Монстры
	mobs := []entity.Monster{}
	for _, rm := range rooms {
		if rng.Intn(2) == 0 {
			mobs = append(mobs, entity.Monster{
				Pos:       entity.Pos{X: rm.X + rm.W/2, Y: rm.Y + rm.H/2},
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
		if rng.Intn(3) == 0 {
			ix := rm.X + rng.Intn(rm.W)
			iy := rm.Y + rng.Intn(rm.H)
			items = append(items, entity.Item{
				Type:   "Food",
				Health: 5,
				Pos:    entity.Pos{X: ix, Y: iy},
			})
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
