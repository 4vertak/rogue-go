package rules

import (
	"math/rand"

	"github.com/4vertak/rogue-go/internal/domain/entity"
)

// бой логика пока минимумальная реализация
func Attack(attacker *entity.Stats, defender *entity.Stats, weapon *entity.Weapon) (bool, int) {

	chance := 50 + (attacker.DEX-defender.DEX)*5
	if chance < 5 {
		chance = 5
	}
	if chance > 95 {
		chance = 95
	}

	if rand.Intn(100) >= chance {
		return false, 0
	}

	damage := attacker.STR
	if weapon != nil {
		damage += rand.Intn(weapon.DamageMax-weapon.DamageMin+1) + weapon.DamageMin
	}

	defender.HP -= damage
	if defender.HP < 0 {
		defender.HP = 0
	}

	return true, damage
}
