package presentation

import "rogue-go/internal/domain/rules"

func (gi *GameInterface) renderMap() {
	gi.gameWindow.Erase()
	gi.gameWindow.Box(0, 0)

	maxWinY, maxWinX := gi.gameWindow.MaxYX()
	gi.game.State.Camera.Width = maxWinX - 2
	gi.game.State.Camera.Height = maxWinY - 2

	data := rules.BuildMapData(
		gi.game.State,
		gi.game.GetItemSymbol,
		gi.game.GetItemColor,
		gi.game.GetEnemySymbol,
		gi.game.GetEnemyColor,
		gi.game.GetKeySymbol,
		gi.game.GetKeyColor,
		gi.game.GetDoorSymbol,
		gi.game.GetDoorColor,
	)

	for _, cell := range data.Cells {
		if cell.Color > 0 {
			gi.gameWindow.ColorOn(cell.Color)
		}
		gi.gameWindow.MovePrint(cell.ScreenY, cell.ScreenX, cell.Ch)
		if cell.Color > 0 {
			gi.gameWindow.ColorOff(cell.Color)
		}
	}

	gi.gameWindow.Refresh()
}
