package gen

import (
	"math/rand"
	"time"

	"github.com/4vertak/rogue-go/internal/domain/entity"
)

type Config struct{ CellW, CellH, MinRW, MinRH int }

func DefaultConfig() Config { return Config{CellW: 10, CellH: 6, MinRW: 4, MinRH: 3} }

func RNG(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }
func nowRNG() *rand.Rand        { return RNG(time.Now().UnixNano()) }

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
	rooms := make([]entity.Room, 0, 9)
	for gy := 0; gy < 3; gy++ {
		for gx := 0; gx < 3; gx++ {
			cx, cy := gx*cfg.CellW, gy*cfg.CellH

			// Исправляем генерацию размеров комнат
			maxRW := cfg.CellW - 2
			if maxRW < cfg.MinRW {
				maxRW = cfg.MinRW
			}
			maxRH := cfg.CellH - 2
			if maxRH < cfg.MinRH {
				maxRH = cfg.MinRH
			}

			rw := cfg.MinRW
			if maxRW > cfg.MinRW {
				rw += rng.Intn(maxRW - cfg.MinRW + 1)
			}
			rh := cfg.MinRH
			if maxRH > cfg.MinRH {
				rh += rng.Intn(maxRH - cfg.MinRH + 1)
			}

			// Исправляем генерацию позиций комнат
			maxX := cfg.CellW - rw - 1
			if maxX < 1 {
				maxX = 1
			}
			maxY := cfg.CellH - rh - 1
			if maxY < 1 {
				maxY = 1
			}

			rx := cx + 1
			if maxX > 0 {
				rx += rng.Intn(maxX)
			}
			ry := cy + 1
			if maxY > 0 {
				ry += rng.Intn(maxY)
			}

			rooms = append(rooms, entity.Room{X: rx, Y: ry, W: rw, H: rh})
			for y := ry; y < ry+rh && y < H; y++ {
				for x := rx; x < rx+rw && x < W; x++ {
					tiles[y][x] = entity.Floor
				}
			}
		}
	}
	// Соединяем комнаты коридорами (право и низ)
	for gy := 0; gy < 3; gy++ {
		for gx := 0; gx < 3; gx++ {
			idx := gy*3 + gx
			if gx < 2 {
				connectRooms(rng, tiles, rooms[idx], rooms[idx+1])
			}
			if gy < 2 {
				connectRooms(rng, tiles, rooms[idx], rooms[idx+3])
			}
		}
	}
	// Выход — в последней комнате
	last := rooms[len(rooms)-1]
	exit := entity.Pos{X: last.X + last.W/2, Y: last.Y + last.H/2}
	tiles[exit.Y][exit.X] = entity.Exit

	// Монстры
	mobs := []entity.Monster{}
	for _, rm := range rooms {
		if rng.Intn(2) == 0 { // шанс
			mobs = append(mobs, entity.Monster{
				Pos:       entity.Pos{X: rm.X + rm.W/2, Y: rm.Y + rm.H/2},
				Stats:     entity.Stats{HP: 5, MaxHP: 5, STR: 3, DEX: 3},
				Type:      "zombie",
				Hostility: 5,
				Symbol:    'z',
			})
		}
	}

	return entity.Level{Index: index, W: W, H: H, Tiles: tiles, Rooms: rooms, Exit: exit, Mobs: mobs}
}

func connectRooms(rng *rand.Rand, tiles [][]entity.Tile, a, b entity.Room) {
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
