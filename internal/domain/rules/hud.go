package rules

import (
	"fmt"
	"rogue-go/internal/domain/entity"
)

type HUDData struct {
	HealthBar  string
	HealthText string
	WeaponText string
	Controls   string
}

func BuildHUD(state *entity.GameState) HUDData {
	var hud HUDData
	if state == nil || state.Player == nil {
		return hud
	}

	player := state.Player
	//  HP
	barWidth := 20
	percent := float64(player.Health) / float64(player.MaxHealth)
	filled := int(percent * float64(barWidth))
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	hud.HealthBar = bar
	hud.HealthText = fmt.Sprintf("%d/%d", player.Health, player.MaxHealth)

	// Weapon
	if player.CurrentWeapon != nil {
		hud.WeaponText = "Weapon: " + player.CurrentWeapon.Name
	} else {
		hud.WeaponText = "Weapon: None"
	}

	hud.Controls = "WASD:Move  Mouse:Look  E:Inventory  Space:Interact  T:2D/3D"

	return hud
}
