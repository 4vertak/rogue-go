package presentation

import (
	"rogue-go/internal/domain/entity"

	"github.com/rthornton128/goncurses"
)

func (gi *GameInterface) handleInput() {
	// Если игра окончена, обрабатываем только выход
	if gi.game.State.GameOver || gi.game.State.GameWon {
		char := gi.stdscr.GetChar()
		if char == 'q' || char == 'Q' || char == goncurses.KEY_ESC {
			gi.game.State.IsRunning = false
		}
		return
	}

	state := gi.game.State

	char := gi.stdscr.GetChar()

	if state.RenderMode == entity.Render3D {
		gi.handle3DInput(char)
		return
	}

	// Оригинальная обработка для 2D режима
	switch char {
	case goncurses.KEY_UP, 'w', 'W':
		gi.game.MovePlayer(0, -1)
	case goncurses.KEY_DOWN, 's', 'S':
		gi.game.MovePlayer(0, 1)
	case goncurses.KEY_LEFT, 'a', 'A':
		gi.game.MovePlayer(-1, 0)
	case goncurses.KEY_RIGHT, 'd', 'D':
		gi.game.MovePlayer(1, 0)
	case 'h', 'H':
		gi.showUseItemMenu(entity.Weapon)
	case 'j', 'J':
		gi.showUseItemMenu(entity.Food)
	case 'k', 'K':
		gi.showUseItemMenu(entity.Potion)
	case 'e', 'E':
		gi.showUseItemMenu(entity.Scroll)
	case 'i', 'I':
		gi.showInventory()
	case 'p', 'P':
		gi.showPickupMenu()
	case 'q', 'Q':
		gi.showQuitMenu()
	case 'v', 'V':
		gi.showSaveMenu()
	case 'o', 'O':
		gi.showLoadMenu()
	case 't', 'T': // Переключение в 3D режим
		state.RenderMode = entity.Render3D
		// Инициализируем направление взгляда, если не инициализировано
		if state.PlayerDir.X == 0 && state.PlayerDir.Y == 0 {
			state.PlayerDir = entity.PositionF{X: -1, Y: 0}
			state.PlayerPlane = entity.PositionF{X: 0, Y: 0.66}
		}
		gi.game.AddMessage("Switched to 3D mode")
	}
}

func (gi *GameInterface) handle3DInput(char goncurses.Key) {
	state := gi.game.State
	moveSpeed := 1
	rotSpeed := 0.05

	switch char {
	case 'w', 'W':
		gi.game.MoveIn3D(moveSpeed)
	case 's', 'S':
		gi.game.MoveIn3D(-moveSpeed)
	case 'a', 'A':
		// Страф влево
		gi.game.Strafe(-moveSpeed)

	case 'd', 'D':
		// Страф вправо
		gi.game.Strafe(moveSpeed)

	case goncurses.KEY_LEFT:
		gi.game.Rotate(-rotSpeed)

	case goncurses.KEY_RIGHT:
		gi.game.Rotate(rotSpeed)

	case 't', 'T':
		state.RenderMode = entity.Render2D
		gi.game.AddMessage("Switched to 2D mode")

	case 'm', 'M':
		gi.showMinimap = !gi.showMinimap
		if gi.showMinimap {
			gi.game.AddMessage("Mini-map enabled")
		} else {
			gi.game.AddMessage("Mini-map disabled")
		}

	// Действия с инвентарем - добавляем обработку оригинальных клавиш
	case 'h', 'H':
		gi.showUseItemMenu(entity.Weapon)
	case 'j', 'J':
		gi.showUseItemMenu(entity.Food)
	case 'k', 'K':
		gi.showUseItemMenu(entity.Potion)
	case 'e', 'E':
		gi.showUseItemMenu(entity.Scroll)
	case 'i', 'I':
		gi.showInventory()
	case 'p', 'P':
		gi.game.PickUpItem()

	// Быстрое поворачивание (новые клавиши)
	case 'q', 'Q':
		state.PlayerAngle -= 45 // Быстрый поворот налево
		if state.PlayerAngle < 0 {
			state.PlayerAngle += 360
		}
	case 'r', 'R': // Вместо 'e' для быстрого поворота направо
		state.PlayerAngle += 45
		if state.PlayerAngle >= 360 {
			state.PlayerAngle -= 360
		}

	case ' ', '\n', '\r': // Пробел или Enter
		// Взаимодействие с ближайшим объектом
		gi.interactWithNearestObject()

	// Бег
	case goncurses.KEY_UP:
		gi.game.MoveIn3D(moveSpeed) // Бег вперед
	case goncurses.KEY_DOWN:
		gi.game.MoveIn3D(-moveSpeed) // Бег назад

	// Выход из игры
	case goncurses.KEY_ESC:
		gi.showQuitMenu()
	}
}

func (gi *GameInterface) interactWithNearestObject() {
	state := gi.game.State
	playerPos := state.Player.Position

	// Проверяем ближайшие клетки
	directions := []entity.Position{
		{X: 0, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: -1, Y: 0},
		{X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1},
	}

	for _, dir := range directions {
		checkX := playerPos.X + dir.X
		checkY := playerPos.Y + dir.Y

		// Проверяем предметы
		for _, item := range state.Items { // УБРАТЬ i
			if item.Position.X == checkX && item.Position.Y == checkY {
				gi.game.PickUpItem()
				return
			}
		}

		// Проверяем врагов
		for i, enemy := range state.Enemies {
			if enemy.Position.X == checkX && enemy.Position.Y == checkY && enemy.IsVisible {
				gi.game.Attack(i)
				return
			}
		}
	}

	gi.game.AddMessage("Nothing to interact with nearby")
}
