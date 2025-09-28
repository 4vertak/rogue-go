package presentation

import (
	"rogue-go/internal/domain/entity"
	"rogue-go/internal/domain/rules"
)

func (gi *GameInterface) renderMinimap() {
	state := gi.game.State
	_, maxX := gi.gameWindow.MaxYX()
	mapSize := 8

	startX := maxX - mapSize - 2
	startY := 1

	mm := rules.BuildMiniMap(
		state,
		mapSize,
		gi.game.GetItemSymbol,
		gi.game.GetItemColor,
		gi.game.GetEnemySymbol,
		gi.game.GetEnemyColor,
		gi.game.GetKeyColor,
	)

	gi.gameWindow.ColorOn(entity.ColorDefault)
	gi.gameWindow.MovePrint(startY, startX, "+")
	for x := 1; x < mapSize+1; x++ {
		gi.gameWindow.MovePrint(startY, startX+x, "-")
	}
	gi.gameWindow.MovePrint(startY, startX+mapSize+1, "+")

	for y := 1; y <= mapSize; y++ {
		gi.gameWindow.MovePrint(startY+y, startX, "|")
		for x := 1; x <= mapSize; x++ {
			var ch string
			var color int16
			for _, cell := range mm.Cells {
				if cell.X == x && cell.Y == y {
					ch = cell.Ch
					color = cell.Color
					break
				}
			}
			if color > 0 {
				gi.gameWindow.ColorOn(color)
			}
			gi.gameWindow.MovePrint(startY+y, startX+x, ch)
			if color > 0 {
				gi.gameWindow.ColorOff(color)
			}
		}
		gi.gameWindow.MovePrint(startY+y, startX+mapSize+1, "|")
	}

	gi.gameWindow.MovePrint(startY+mapSize+1, startX, "+")
	for x := 1; x < mapSize+1; x++ {
		gi.gameWindow.MovePrint(startY+mapSize+1, startX+x, "-")
	}
	gi.gameWindow.MovePrint(startY+mapSize+1, startX+mapSize+1, "+")

	gi.gameWindow.ColorOff(entity.ColorDefault)
}
