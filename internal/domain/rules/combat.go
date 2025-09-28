package rules

import (
	"math/rand"
	"strconv"

	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/gen"
	"rogue-go/internal/domain/utils"
)

func isInAttackRange(pos1, pos2 entity.Position) bool {
	dx := utils.Abs(pos1.X - pos2.X)
	dy := utils.Abs(pos1.Y - pos2.Y)
	return dx <= 1 && dy <= 1
}

func Attack(state *entity.GameState, enemyIndex int) {
	// Проверяем, не усыплен ли игрок
	if state.SleepTurns > 0 {
		state.SleepTurns--
		utils.AddMessage(state, "You are asleep, can't attack")
		return
	}
	if enemyIndex < len(state.Enemies) {
		enemy := state.Enemies[enemyIndex]
		if isInAttackRange(state.Player.Position, enemy.Position) {
			resolveCombat(state, state.Player, enemy)
			if enemy.Health <= 0 {
				handleEnemyDefeat(state, enemy)
				state.Enemies = append(state.Enemies[:enemyIndex], state.Enemies[enemyIndex+1:]...)
			}
		}
	}
}

func resolveCombat(state *entity.GameState, attacker interface{}, defender interface{}) {
	var attackerStr, attackerAgi int

	// Получаем характеристики атакующего
	switch a := attacker.(type) {
	case *entity.Character:
		attackerStr = a.Strength
		if a.CurrentWeapon != nil {
			attackerStr += a.CurrentWeapon.StrengthBoost
		}
		attackerAgi = a.Agility
	case *entity.Enemy:
		attackerStr = a.Strength
		attackerAgi = a.Agility
	}

	// Получаем характеристики защищающегося
	var defenderAgi int
	switch d := defender.(type) {
	case *entity.Character:
		defenderAgi = d.Agility
	case *entity.Enemy:
		defenderAgi = d.Agility
	}

	hitChance := 50 + (attackerAgi - defenderAgi)
	if hitChance < 20 {
		hitChance = 20
	}
	if hitChance > 90 {
		hitChance = 90
	}

	// Рассчитываем шанс попадания (более сбалансированный)
	if rand.Intn(100) < hitChance {
		damage := attackerStr

		// Настраиваем урон в зависимости от того, кто атакует
		switch d := defender.(type) {
		case *entity.Character:
			// Враги наносят меньше урона игроку
			if _, ok := attacker.(*entity.Enemy); ok {
				damage = max(1, attackerStr/3) // Враги бьют значительно слабее
			}
			d.Health -= damage
			state.GameSession.HitsTaken++
			utils.AddMessage(state, "You take "+strconv.Itoa(damage)+" damage!")

			// спецэффекты врагов
			if e, ok := attacker.(*entity.Enemy); ok {
				applyEnemySpecialEffects(state, e)
			}
		case *entity.Enemy:
			// Игрок наносит полный урон врагам
			if _, ok := attacker.(*entity.Character); ok {
				damage = attackerStr
			} else {
				damage = max(1, attackerStr/2) // Враги друг другу наносят средний урон
			}
			d.Health -= damage
			state.GameSession.HitsLanded++
			utils.AddMessage(state, "You hit enemy for "+strconv.Itoa(damage)+" damage!")

			// особенность вампира: первый удар = дренаж max HP
			if d.Type == entity.Vampire && !d.HasBeenHit {
				d.HasBeenHit = true
				if char, ok := attacker.(*entity.Character); ok {
					char.MaxHealth = max(10, char.MaxHealth-2) // Меньший дренаж
					if char.Health > char.MaxHealth {
						char.Health = char.MaxHealth
					}
					utils.AddMessage(state, "Vampire drained your max health! -2 Max HP")
				}
			}
		}
	} else {
		utils.AddMessage(state, "Attack missed!")
	}
}

// Специальные эффекты для Mimic
func applyEnemySpecialEffects(state *entity.GameState, enemy *entity.Enemy) {
	switch enemy.Type {
	case entity.SnakeMage:
		if rand.Intn(100) < 30 {
			state.SleepTurns = 2 + rand.Intn(3)
			utils.AddMessage(state, "Snake Mage puts you to sleep for "+strconv.Itoa(state.SleepTurns)+" turns!")
		}
	case entity.Mimic:
		if rand.Intn(100) < 25 {
			state.SleepTurns = 1
			utils.AddMessage(state, "Mimic's surprise attack stuns you!")
		}
	}
}

func handleEnemyDefeat(state *entity.GameState, enemy *entity.Enemy) {
	state.Player.DefeatedEnemies++
	state.GameSession.EnemiesDefeated++

	switch enemy.Type {
	case entity.Zombie:
		utils.AddMessage(state, "Zombie defeated! +10 gold")
		state.Backpack.Treasure += 10
	case entity.Vampire:
		utils.AddMessage(state, "Vampire defeated! +25 gold")
		state.Backpack.Treasure += 25
	case entity.Ghost:
		utils.AddMessage(state, "Ghost defeated! +20 gold")
		state.Backpack.Treasure += 20
	case entity.Ogre:
		utils.AddMessage(state, "Ogre defeated! +30 gold")
		state.Backpack.Treasure += 30
	case entity.SnakeMage:
		utils.AddMessage(state, "Snake Mage defeated! +40 gold")
		state.Backpack.Treasure += 40
	case entity.Mimic:
		utils.AddMessage(state, "Mimic defeated! +15 gold")
		state.Backpack.Treasure += 15
	}

	if rand.Intn(100) < 20 {
		item := gen.CreateItem(state.CurrentLevel, nil, state, state.Level)
		if item != nil {
			item.Position = enemy.Position
			state.Items = append(state.Items, item)
			utils.AddMessage(state, "Enemy dropped an item!")
		}
	}
}
