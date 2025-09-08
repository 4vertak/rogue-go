package rules

import (
	"math/rand"

	"github.com/4vertak/rogue-go/internal/domain/entity"
)

// движение игрока
func MovePlayer(p *entity.Player, dx, dy int, level *entity.Level) {
	newX := p.Pos.X + dx
	newY := p.Pos.Y + dy

	if newX < 0 || newY < 0 || newX >= level.W || newY >= level.H {
		return
	}

	if level.Tiles[newY][newX] == entity.Wall {
		return
	}

	p.Pos.X = newX
	p.Pos.Y = newY
}

// паттерн движения моснтра -минимум для наглядности и проверки=)
func MoveMonster(m *entity.Monster, player entity.Player, level *entity.Level) {
	dx, dy := 0, 0

	distX := player.Pos.X - m.Pos.X
	distY := player.Pos.Y - m.Pos.Y
	if Abs(distX)+Abs(distY) <= m.Hostility {
		if Abs(distX) > Abs(distY) {
			dx = Sign(distX)
		} else {
			dy = Sign(distY)
		}
	} else {

		dirs := []entity.Pos{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}}
		d := dirs[rand.Intn(len(dirs))]
		dx, dy = d.X, d.Y
	}

	newX := m.Pos.X + dx
	newY := m.Pos.Y + dy
	if newX < 0 || newY < 0 || newX >= level.W || newY >= level.H {
		return
	}
	if level.Tiles[newY][newX] == entity.Wall {
		return
	}
	m.Pos.X = newX
	m.Pos.Y = newY
}
