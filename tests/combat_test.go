package tests

import (
	"testing"

	"github.com/4vertak/rogue-go/internal/domain/entity"
	"github.com/4vertak/rogue-go/internal/domain/rules"
)

func TestAttackAlwaysHitsWithHighDex(t *testing.T) {
	attacker := entity.Stats{STR: 5, DEX: 20, HP: 10, MaxHP: 10}
	defender := entity.Stats{STR: 3, DEX: 1, HP: 10, MaxHP: 10}

	hitCount := 0
	for i := 0; i < 50; i++ {
		hit, _ := rules.Attack(&attacker, &defender, nil)
		if hit {
			hitCount++
		}
	}
	if hitCount == 0 {
		t.Errorf("expected some hits, got 0")
	}
}

func TestAttackMissWithLowDex(t *testing.T) {
	attacker := entity.Stats{STR: 5, DEX: 1, HP: 10, MaxHP: 10}
	defender := entity.Stats{STR: 3, DEX: 20, HP: 10, MaxHP: 10}

	hitCount := 0
	for i := 0; i < 50; i++ {
		hit, _ := rules.Attack(&attacker, &defender, nil)
		if hit {
			hitCount++
		}
	}
	if hitCount > 10 {
		t.Errorf("expected mostly misses, got %d hits", hitCount)
	}
}

func TestAttackWithWeapon(t *testing.T) {
	attacker := entity.Stats{STR: 2, DEX: 10, HP: 10, MaxHP: 10}
	defender := entity.Stats{STR: 3, DEX: 1, HP: 20, MaxHP: 20}
	weapon := &entity.Weapon{Name: "Sword", DamageMin: 2, DamageMax: 4}

	hit, dmg := rules.Attack(&attacker, &defender, weapon)
	if hit && dmg < 4 { // 2 STR + min weapon dmg 2
		t.Errorf("expected damage >= 4, got %d", dmg)
	}
}
