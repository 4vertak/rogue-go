package rules

import (
	"fmt"
	"rogue-go/internal/domain/entity"
)

type InventoryData struct {
	Title string
	Lines []string
}

func BuildInventory(state *entity.GameState) InventoryData {
	var invent InventoryData
	invent.Title = "Inventory:"
	if state == nil || state.Backpack == nil {
		invent.Lines = []string{"W:0 F:0 P:0 S:0"}
		return invent
	}

	w := len(state.Backpack.Weapons)
	f := len(state.Backpack.Food)
	p := len(state.Backpack.Potions)
	s := len(state.Backpack.Scrolls)

	invent.Lines = []string{
		fmt.Sprintf("W:%d F:%d P:%d S:%d", w, f, p, s),
	}
	return invent
}
