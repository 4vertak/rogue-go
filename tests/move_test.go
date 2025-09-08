package tests

import (
	"testing"

	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/rules"
)

func makeLevel() *entity.Level {
	tiles := [][]entity.Tile{
		{entity.Floor, entity.Floor, entity.Wall},
		{entity.Floor, entity.Floor, entity.Floor},
	}
	return &entity.Level{W: 3, H: 2, Tiles: tiles}
}

func TestMovePlayerBlockedByWall(t *testing.T) {
	lv := makeLevel()
	p := entity.DefaultPlayer()
	p.Pos = entity.Pos{X: 1, Y: 0} // рядом со стеной

	rules.MovePlayer(&p, 1, 0, lv) // вправо в стену
	if p.Pos.X != 1 || p.Pos.Y != 0 {
		t.Errorf("expected player to stay, got %+v", p.Pos)
	}
}

func TestMovePlayerIntoFloor(t *testing.T) {
	lv := makeLevel()
	p := entity.DefaultPlayer()
	p.Pos = entity.Pos{X: 0, Y: 0}

	rules.MovePlayer(&p, 1, 0, lv) // вправо в пол
	if p.Pos.X != 1 || p.Pos.Y != 0 {
		t.Errorf("expected player to move to (1,0), got %+v", p.Pos)
	}
}

func TestMonsterChasePlayer(t *testing.T) {
	lv := makeLevel()
	p := entity.DefaultPlayer()
	p.Pos = entity.Pos{X: 1, Y: 1}

	m := entity.Monster{
		Pos:       entity.Pos{X: 0, Y: 0},
		Stats:     entity.Stats{HP: 10, MaxHP: 10},
		Type:      "Zombie",
		Hostility: 5,
		Symbol:    'z',
	}

	rules.MoveMonster(&m, p, lv)
	if !(m.Pos.X == 0 && m.Pos.Y == 1) && !(m.Pos.X == 1 && m.Pos.Y == 0) {
		t.Errorf("expected monster to move closer, got %+v", m.Pos)
	}
}
