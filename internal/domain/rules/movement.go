package rules

import (
	"math"
	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/utils"
	"strconv"
)

func MovePlayer(state *entity.GameState, dx, dy int) {
	// Проверяем, не усыплен ли игрок
	if state.SleepTurns > 0 {
		state.SleepTurns--
		utils.AddMessage(state, "You are asleep for "+strconv.Itoa(state.SleepTurns)+" more turns")
		return
	}

	newX := state.Player.Position.X + dx
	newY := state.Player.Position.Y + dy

	if canMoveTo(state, newX, newY) {
		state.Player.Position.X = newX
		state.Player.Position.Y = newY
		state.GameSession.CellsWalked++
		UpdateVisibility(state)

		// Автоматически подбираем ВСЕ предметы на клетке, КРОМЕ КЛЮЧЕЙ
		picked := false
		for i := len(state.Items) - 1; i >= 0; i-- {
			item := state.Items[i]
			if item.Position.X == newX && item.Position.Y == newY {
				if item.Type == entity.Scroll && (item.SubType == entity.KeyRed || item.SubType == entity.KeyBlue ||
					item.SubType == entity.KeyGreen || item.SubType == entity.KeyYellow) {
					utils.AddMessage(state, "You see a key here (press 'p' to pick up)")
					continue // Пропускаем ключи - они не подбираются автоматически
				}
				if canAddToBackpack(state, item) {
					addToBackpack(state, item) // ТОЛЬКО ОДИН РАЗ вызываем добавление
					state.Items = append(state.Items[:i], state.Items[i+1:]...)
					picked = true
				} else {
					utils.AddMessage(state, "Item here but backpack is full for this type")
				}
			}
		}
		if picked {
			utils.AddMessage(state, "You picked up items automatically")
		}

		// После движения проверяем автоматический переход на следующий уровень
		checkForAutomaticExit(state, nil)
		CheckForEnemies(state) // Бой при движении игрока
		checkForExit(state)
	}
}

func canMoveTo(state *entity.GameState, x, y int) bool {
	if x < 0 || x >= len(state.Level.Grid) || y < 0 || y >= len(state.Level.Grid[0]) {
		return false
	}
	// Проверяем стену
	if state.Level.Grid[x][y].IsWall {
		return false
	}
	// Проверяем заблокированную дверь
	for _, door := range state.Level.Doors {
		if door.Position.X == x && door.Position.Y == y && door.IsLocked {
			return false // Не можем пройти через врага
		}
	}
	// Проверяем, нет ли врага на этой клетке
	for _, enemy := range state.Enemies {
		if enemy.Position.X == x && enemy.Position.Y == y {
			return false
		}
	}

	// ПРЕДМЕТЫ НЕ БЛОКИРУЮТ ДВИЖЕНИЕ - они автоматически подбираются

	return true
}

func PickUpItem(state *entity.GameState) {
	// Проверяем, не усыплен ли игрок
	if state.SleepTurns > 0 {
		state.SleepTurns--
		utils.AddMessage(state, "You are asleep, can't pick up items")
		return
	}

	playerPos := state.Player.Position
	foundItem := false

	for i := len(state.Items) - 1; i >= 0; i-- {
		item := state.Items[i]
		if item.Position.X == playerPos.X && item.Position.Y == playerPos.Y {
			if canAddToBackpack(state, item) {
				addToBackpack(state, item) // ТОЛЬКО ОДИН РАЗ вызываем добавление
				state.Items = append(state.Items[:i], state.Items[i+1:]...)
				foundItem = true
				break // ВЫХОДИМ после первого найденного предмета
			} else {
				utils.AddMessage(state, "Backpack full for this item type!")
			}
		}
	}

	if !foundItem {
		utils.AddMessage(state, "No items to pick up here")
	}
}

func CheckForEnemies(state *entity.GameState) {
	playerPos := state.Player.Position

	for i := len(state.Enemies) - 1; i >= 0; i-- {
		enemy := state.Enemies[i]
		// Проверяем соседние клетки (атака на расстоянии 1 клетки)
		if utils.Abs(enemy.Position.X-playerPos.X) <= 1 && utils.Abs(enemy.Position.Y-playerPos.Y) <= 1 {
			resolveCombat(state, state.Player, enemy)
			// Удаляем врага из списка
			if enemy.Health <= 0 {
				handleEnemyDefeat(state, enemy)
				state.Enemies = append(state.Enemies[:i], state.Enemies[i+1:]...)
			} else {
				// Также враг атакует игрока в ответ
				resolveCombat(state, enemy, state.Player)
			}
		}
	}
}

func canAddToBackpack(state *entity.GameState, item *entity.Item) bool {
	if item == nil {
		return false
	}
	if item.Type == entity.Treasure {
		return true
	}
	// КЛЮЧИ МОЖНО ВСЕГДА ПОДБИРАТЬ
	if item.Type == entity.Scroll && (item.SubType == entity.KeyRed || item.SubType == entity.KeyBlue ||
		item.SubType == entity.KeyGreen || item.SubType == entity.KeyYellow) {
		return true
	}
	bp := state.Backpack
	if bp == nil {
		return true
	}
	switch item.Type {
	case entity.Weapon:
		return len(bp.Weapons) < 5
	case entity.Food:
		return len(bp.Food) < 10
	case entity.Potion:
		return len(bp.Potions) < 5
	case entity.Scroll:
		return len(bp.Scrolls) < 5
	default:
		return false
	}
}

func addToBackpack(state *entity.GameState, item *entity.Item) {
	if item == nil || state.Backpack == nil {
		return
	}

	switch item.Type {
	case entity.Weapon:
		// Если у игрока нет оружия - автоматически экипируем
		if state.Player.CurrentWeapon == nil {
			state.Player.CurrentWeapon = item
			utils.AddMessage(state, "Equipped weapon: +"+strconv.Itoa(item.StrengthBoost)+" STR")
		} else {
			// Если уже есть оружие - кладем новое в рюкзак
			state.Backpack.Weapons = append(state.Backpack.Weapons, item)
			utils.AddMessage(state, "Picked up weapon: +"+strconv.Itoa(item.StrengthBoost)+" STR (added to backpack)")
		}

	case entity.Food:
		state.Backpack.Food = append(state.Backpack.Food, item)
		utils.AddMessage(state, "Picked up food: +"+strconv.Itoa(item.HealthBoost)+" HP")

	case entity.Potion:
		state.Backpack.Potions = append(state.Backpack.Potions, item)
		utils.AddMessage(state, "Picked up potion")

	case entity.Scroll:
		// ОСОБАЯ ОБРАБОТКА ДЛЯ КЛЮЧЕЙ
		if item.SubType == entity.KeyRed || item.SubType == entity.KeyBlue ||
			item.SubType == entity.KeyGreen || item.SubType == entity.KeyYellow {
			state.Backpack.Scrolls = append(state.Backpack.Scrolls, item)
			utils.AddMessage(state, "Picked up "+utils.GetKeyName(item.SubType))
		} else {
			state.Backpack.Scrolls = append(state.Backpack.Scrolls, item)
			utils.AddMessage(state, "Picked up scroll")
		}

	case entity.Treasure:
		state.Backpack.Treasure += item.Value
		if state.Player != nil {
			state.Player.Treasures += item.Value
		}
		utils.AddMessage(state, "Picked up treasure: +"+strconv.Itoa(item.Value)+" gold")
	}
}

func checkForAutomaticExit(state *entity.GameState, nextLevel func()) {
	if state.SleepTurns > 0 {
		return
	}
	playerPos := state.Player.Position
	exitRoom := state.Level.ExitRoom
	if exitRoom == nil {
		return
	}

	// Проверяем конкретную клетку выхода, а не всю комнату
	exitX := exitRoom.X + exitRoom.Width/2
	exitY := exitRoom.Y + exitRoom.Height/2
	playerAtExit := playerPos.X == exitX && playerPos.Y == exitY
	if playerAtExit {
		// Проверяем, что все враги побеждены
		if len(state.Enemies) == 0 {
			// АВТОМАТИЧЕСКИ переходим на следующий уровень
			if nextLevel != nil {
				nextLevel()
			}
		} else {
			utils.AddMessage(state, "You must defeat all enemies before exiting! "+strconv.Itoa(len(state.Enemies))+" enemies remaining")
		}
	}
}

func checkForExit(state *entity.GameState) {
	playerPos := state.Player.Position
	exitRoom := state.Level.ExitRoom
	if exitRoom != nil &&
		playerPos.X >= exitRoom.X && playerPos.X < exitRoom.X+exitRoom.Width &&
		playerPos.Y >= exitRoom.Y && playerPos.Y < exitRoom.Y+exitRoom.Height {
		utils.AddMessage(state, "You found the exit! (defeat all enemies to win)")
	}
}

func MoveIn3D(state *entity.GameState, distance int) {
	moveSpeed := distance //* 0.1 // Увеличим скорость движения
	player := state.Player
	dir := state.PlayerDir

	// Правильно вычисляем новую позицию с учетом направления взгляда
	newX := player.Position.X + int(math.Round(dir.X*float64(moveSpeed)))
	newY := player.Position.Y + int(math.Round(dir.Y*float64(moveSpeed)))

	// Проверяем возможность движения
	if canMoveTo3D(state, int(newX), int(newY)) {
		player.Position.X = int(newX)
		player.Position.Y = int(newY)
		UpdateVisibility(state)
		utils.AddMessage(state, "Moved forward")

		// Автоподбор предметов и проверка врагов
		checkForAutoPickup(state)
		CheckForEnemies(state)
	} else {
		utils.AddMessage(state, "Cannot move - obstacle ahead")
	}
}

func canMoveTo3D(state *entity.GameState, x, y int) bool {
	// Проверяем границы карты
	if x < 0 || x >= len(state.Level.Grid) ||
		y < 0 || y >= len(state.Level.Grid[0]) {
		utils.AddMessage(state, "Cannot move - edge of map")
		return false
	}

	// Проверяем стену
	if state.Level.Grid[x][y].IsWall {
		return false
	}

	// Проверяем заблокированную дверь
	for _, door := range state.Level.Doors {
		if door.Position.X == x && door.Position.Y == y && door.IsLocked {
			return false
		}
	}

	// Проверяем врагов
	for _, enemy := range state.Enemies {
		if enemy.Position.X == x && enemy.Position.Y == y {
			utils.AddMessage(state, "Enemy blocks the way!")
			return false
		}
	}

	return true
}

// Исправим функцию strafe для бокового движения
func Strafe(state *entity.GameState, distance int) {
	moveSpeed := distance //* 0.1
	player := state.Player
	dir := state.PlayerDir

	// Перпендикулярное направление для страфа
	perpX := -dir.Y
	perpY := dir.X

	newX := player.Position.X + int(math.Round(perpX*float64(moveSpeed)))
	newY := player.Position.Y + int(math.Round(perpY*float64(moveSpeed)))

	if canMoveTo3D(state, int(newX), int(newY)) {
		player.Position.X = int(newX)
		player.Position.Y = int(newY)
		utils.AddMessage(state, "Strafed")

		checkForAutoPickup(state)
		CheckForEnemies(state)
	}

}

func Rotate(state *entity.GameState, angle float64) {
	oldDirX := state.PlayerDir.X
	state.PlayerDir.X = state.PlayerDir.X*math.Cos(angle) - state.PlayerDir.Y*math.Sin(angle)
	state.PlayerDir.Y = oldDirX*math.Sin(angle) + state.PlayerDir.Y*math.Cos(angle)

	oldPlaneX := state.PlayerPlane.X
	state.PlayerPlane.X = state.PlayerPlane.X*math.Cos(angle) - state.PlayerPlane.Y*math.Sin(angle)
	state.PlayerPlane.Y = oldPlaneX*math.Sin(angle) + state.PlayerPlane.Y*math.Cos(angle)
}

func InteractWithDoor(state *entity.GameState) {
	// Проверяем дверь перед игроком
	angleRad := state.PlayerAngle * math.Pi / 180
	checkX := state.Player.Position.X + int(1.5*math.Cos(angleRad))
	checkY := state.Player.Position.Y + int(1.5*math.Sin(angleRad))

	// Проходим по всем дверям без использования индекса
	doorFound := false
	for _, door := range state.Level.Doors {
		if door.Position.X == checkX && door.Position.Y == checkY {
			doorFound = true
			if door.IsLocked {
				utils.AddMessage(state, "The door is locked")
			} else {
				// Открываем/закрываем дверь
				utils.AddMessage(state, "You open the door")
			}
			break
		}
	}

	if !doorFound {
		utils.AddMessage(state, "No door nearby")
	}
}

// Новая функция для автоподбора в 3D
func checkForAutoPickup(state *entity.GameState) {
	playerPos := state.Player.Position
	for i := len(state.Items) - 1; i >= 0; i-- {
		item := state.Items[i]
		if item.Position.X == playerPos.X && item.Position.Y == playerPos.Y {
			if canAddToBackpack(state, item) {
				addToBackpack(state, item)
				state.Items = append(state.Items[:i], state.Items[i+1:]...)
				utils.AddMessage(state, "Picked up item automatically")
			}
		}
	}
}
