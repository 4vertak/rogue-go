package presentation

import (
	"rogue-go/internal/domain/entity"
)

func (gi *GameInterface) createWindows() {
	rows, cols := gi.stdscr.MaxYX()

	// Главное игровое окно (центр экрана)
	gameRows := rows - 6 // Оставляем место для HUD в 3D режиме
	gameCols := cols - 31

	if gameRows < 10 {
		gameRows = 10
	}
	if gameCols < 20 {
		gameCols = 20
	}

	gi.gameWindow = gi.stdscr.Sub(gameRows, gameCols, 0, 0)
	gi.gameWindow.Keypad(true)

	// Окно статуса (справа)
	gi.statusWindow = gi.stdscr.Sub(gameRows, 30, 0, gameCols)
	gi.statusWindow.Keypad(true)

	// Окно сообщений (снизу слева)
	gi.messageWindow = gi.stdscr.Sub(3, gameCols, gameRows, 0)
	gi.messageWindow.Keypad(true)

	// Окно инвентаря (снизу справа)
	gi.inventoryWindow = gi.stdscr.Sub(3, 30, gameRows, gameCols)
	gi.inventoryWindow.Keypad(true)

	// Камера
	if gi.game.State.Camera != nil {
		maxY, maxX := gi.gameWindow.MaxYX()
		gi.game.State.Camera.Width = maxX - 2
		gi.game.State.Camera.Height = maxY - 2
	}
}

func (gi *GameInterface) renderInitialScreen() {
	gi.stdscr.Erase()
	gi.stdscr.Print("=== ROGUE GAME ===\n")
	gi.stdscr.Print("Press any key to start...")
	gi.stdscr.Refresh()

	gi.stdscr.Timeout(-1)
	gi.stdscr.GetChar()
	gi.stdscr.Timeout(100)
}

func (gi *GameInterface) renderGame() {
	state := gi.game.State
	if state.RenderMode == entity.Render3D {
		gi.render3D()
	} else {
		gi.renderMap()
	}

	// Всегда рендерим остальные окна независимо от режима
	gi.renderStatus()
	gi.renderMessageBar()
	gi.renderInventory()
	gi.stdscr.Refresh()
}
