package rules

import (
	"math/rand"

	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/utils"
)

func MoveEnemies(state *entity.GameState) {
	for _, enemy := range state.Enemies {
		if enemy.AttackCooldown > 0 {
			enemy.AttackCooldown--
			continue
		}
		if enemy.TeleportCooldown > 0 {
			enemy.TeleportCooldown--
		}
		if enemy.InvisibilityCooldown > 0 {
			enemy.InvisibilityCooldown--
			if enemy.InvisibilityCooldown == 0 {
				enemy.IsVisible = true
				utils.AddMessage(state, "Ghost becomes visible again!")
			}
		}

		playerPos := state.Player.Position
		dist := utils.Abs(enemy.Position.X-playerPos.X) + utils.Abs(enemy.Position.Y-playerPos.Y)
		// Проверяем, находится ли игрок в радиусе враждебности
		if dist <= enemy.Hostility {
			// Игрок в зоне агрессии - преследуем!
			switch enemy.Type {
			case entity.Zombie:
				moveRandom(state, enemy)
			case entity.Vampire:
				moveChase(state, enemy)
			case entity.Ghost:
				moveTeleport(state, enemy)
			case entity.Ogre:
				moveTwoStep(state, enemy)
			case entity.SnakeMage:
				moveDiagonal(state, enemy)
			case entity.Mimic:
				moveMimic(state, enemy)
			default:
				moveRandom(state, enemy)
			}
		} else {
			// Игрок вне зоны агрессии - патрулируем
			switch enemy.Type {
			case entity.Mimic:
				moveMimic(state, enemy)
			default:
				moveRandom(state, enemy) // Все остальные враги патрулируют случайною Убрал лишнее все равно повторяется))
			}
		}

		// Проверяем атаку после движения
		if isInAttackRange(enemy.Position, state.Player.Position) {
			resolveCombat(state, enemy, state.Player)
			if enemy.Type == entity.Ogre {
				enemy.AttackCooldown = 1
				utils.AddMessage(state, "Ogre is resting after attack")
			}
		}
	}
}

func moveRandom(state *entity.GameState, enemy *entity.Enemy) {
	directions := []entity.Position{{X: 0, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: -1, Y: 0}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}}
	shuffled := make([]entity.Position, len(directions))
	copy(shuffled, directions)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	for _, dir := range shuffled {
		newX := enemy.Position.X + dir.X
		newY := enemy.Position.Y + dir.Y
		if canMoveTo(state, newX, newY) {
			enemy.Position.X = newX
			enemy.Position.Y = newY
			break
		}
	}
}

func moveChase(state *entity.GameState, enemy *entity.Enemy) {
	playerPos := state.Player.Position
	enemyPos := enemy.Position
	dx := 0
	if enemyPos.X < playerPos.X {
		dx = 1
	}
	if enemyPos.X > playerPos.X {
		dx = -1
	}
	dy := 0
	if enemyPos.Y < playerPos.Y {
		dy = 1
	}
	if enemyPos.Y > playerPos.Y {
		dy = -1
	}

	if rand.Intn(100) < 70 {
		if dx != 0 && canMoveTo(state, enemyPos.X+dx, enemyPos.Y) {
			enemy.Position.X += dx
		} else if dy != 0 && canMoveTo(state, enemyPos.X, enemyPos.Y+dy) {
			enemy.Position.Y += dy
		}
	} else {
		moveRandom(state, enemy)
	}
}

func moveTeleport(state *entity.GameState, enemy *entity.Enemy) {
	// Призрак телепортируется и становится невидимым
	if enemy.TeleportCooldown == 0 {
		// Телепортируемся рядом с игроком
		playerPos := state.Player.Position
		directions := []entity.Position{{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}}
		for _, dir := range directions {
			newX := playerPos.X + dir.X
			newY := playerPos.Y + dir.Y
			if canMoveTo(state, newX, newY) {
				enemy.Position.X = newX
				enemy.Position.Y = newY
				enemy.TeleportCooldown = 3 // Кулдаун телепорта

				enemy.IsVisible = false // Становится невидимым
				enemy.InvisibilityCooldown = 2
				utils.AddMessage(state, "Ghost teleports and becomes invisible!")
				break
			}
		}
	} else {
		// Двигаемся случайно во время кулдауна
		moveRandom(state, enemy)
	}
}

func moveDiagonal(state *entity.GameState, enemy *entity.Enemy) {
	// Змей-маг двигается по диагонали
	directions := []entity.Position{{X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}}
	if rand.Intn(100) < 60 {
		playerPos := state.Player.Position
		enemyPos := enemy.Position
		dx := 1
		if enemyPos.X > playerPos.X {
			dx = -1
		}
		dy := 1
		if enemyPos.Y > playerPos.Y {
			dy = -1
		}
		if canMoveTo(state, enemyPos.X+dx, enemyPos.Y+dy) {
			enemy.Position.X += dx
			enemy.Position.Y += dy
			return
		}
	}
	rand.Shuffle(len(directions), func(i, j int) { directions[i], directions[j] = directions[j], directions[i] })
	for _, dir := range directions {
		newX := enemy.Position.X + dir.X
		newY := enemy.Position.Y + dir.Y
		if canMoveTo(state, newX, newY) {
			enemy.Position.X = newX
			enemy.Position.Y = newY
			break
		}
	}
}

func moveTwoStep(state *entity.GameState, enemy *entity.Enemy) {
	// Огр двигается на две клетки за ход, но с кулдауном
	if enemy.MoveCooldown == 0 { // Первый шаг
		moveChase(state, enemy)
		enemy.MoveCooldown = 1
	} else { // Второй шаг
		moveChase(state, enemy)
		enemy.MoveCooldown = 0
	}
}

// Функция для движения Mimic
func moveMimic(state *entity.GameState, enemy *entity.Enemy) {
	playerPos := state.Player.Position
	dist := utils.Abs(enemy.Position.X-playerPos.X) + utils.Abs(enemy.Position.Y-playerPos.Y)
	// Мимик раскрывается, когда игрок подходит близко
	if dist <= 2 {
		enemy.IsDisguised = false
		utils.AddMessage(state, "The item was a Mimic!")
	}
	// Если раскрыт, ведет себя как обычный враг
	if !enemy.IsDisguised {

		moveRandom(state, enemy)
	}
	// Если замаскирован, не двигается
}
