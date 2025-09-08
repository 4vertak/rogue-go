package tests

import (
	"testing"

	"github.com/4vertak/rogue-go/internal/domain/entity"
)

func TestDefaultPlayer(t *testing.T) {
	p := entity.DefaultPlayer()

	if p.Stats.HP != 20 || p.Stats.MaxHP != 20 {
		t.Errorf("expected player HP=20/20, got %d/%d", p.Stats.HP, p.Stats.MaxHP)
	}
	if p.Stats.DEX != 5 || p.Stats.STR != 5 {
		t.Errorf("expected DEX=5, STR=5, got %d/%d", p.Stats.DEX, p.Stats.STR)
	}
	if p.Pos.X != 2 || p.Pos.Y != 2 {
		t.Errorf("expected start pos (2,2), got (%d,%d)", p.Pos.X, p.Pos.Y)
	}
}

func TestWeaponAssignment(t *testing.T) {
	p := entity.DefaultPlayer()
	sword := &entity.Weapon{Name: "Sword", DamageMin: 2, DamageMax: 4, ToHitBonus: 1}
	p.Weapon = sword

	if p.Weapon == nil || p.Weapon.Name != "Sword" {
		t.Errorf("expected player to have weapon Sword, got %+v", p.Weapon)
	}
}

func TestMonsterCreation(t *testing.T) {
	m := entity.Monster{
		Pos:       entity.Pos{X: 1, Y: 1},
		Stats:     entity.Stats{MaxHP: 10, HP: 10, DEX: 3, STR: 4},
		Type:      "Zombie",
		Hostility: 5,
		Symbol:    'z',
	}

	if m.Type != "Zombie" || m.Symbol != 'z' {
		t.Errorf("monster type mismatch: %+v", m)
	}
	if m.Stats.HP > m.Stats.MaxHP {
		t.Errorf("monster HP cannot exceed MaxHP")
	}
}
