package rules

import (
	"strconv"

	"rogue-go/internal/domain/entity"
)

func UseItem(state *entity.GameState, itemType entity.ItemType, index int) {
	if state.SleepTurns > 0 {
		state.SleepTurns--
		state.Messages = append(state.Messages, "You are asleep, can't use items")
		return
	}
	if state.Backpack == nil {
		state.Messages = append(state.Messages, "No backpack available")
		return
	}

	switch itemType {
	case entity.Food:
		useFood(state, index)

	case entity.Potion:
		usePotion(state, index)

	case entity.Scroll:
		useScroll(state, index)

	case entity.Weapon:
		useWeapon(state, index)
	}
}

func GetItemName(item *entity.Item) string {
	switch item.Type {
	case entity.Weapon:
		return "Weapon (Str+" + strconv.Itoa(item.StrengthBoost) + ")"
	case entity.Food:
		return "Food (HP+" + strconv.Itoa(item.HealthBoost) + ")"
	case entity.Potion:
		switch item.SubType {
		case entity.HealthPotion:
			return "Health Potion"
		case entity.AgilityPotion:
			return "Agility Potion"
		case entity.StrengthPotion:
			return "Strength Potion"
		default:
			return "Potion"
		}
	case entity.Scroll:
		switch item.SubType {
		case entity.MaxHealthScroll:
			return "Max Health Scroll"
		case entity.AgilityScroll:
			return "Agility Scroll"
		case entity.StrengthScroll:
			return "Strength Scroll"
		case entity.KeyRed:
			return "Red Key"
		case entity.KeyBlue:
			return "Blue Key"
		case entity.KeyGreen:
			return "Green Key"
		case entity.KeyYellow:
			return "Yellow Key"
		default:
			return "Scroll"
		}
	case entity.Treasure:
		return "Treasure (" + strconv.Itoa(item.Value) + " gold)"
	default:
		return "Unknown item"
	}
}

func GetItemSymbol(item *entity.Item) string {
	switch item.Type {
	case entity.Weapon:
		return "!"
	case entity.Food:
		return "%"
	case entity.Potion:
		return "&"
	case entity.Scroll:
		return "?"
	case entity.Treasure:
		return "$"
	default:
		return "?"
	}
}

func GetItemColor(item *entity.Item) int16 {
	switch item.Type {
	case entity.Weapon:
		return 4
	case entity.Food:
		return 2
	case entity.Potion:
		return 5
	case entity.Scroll:
		return 3
	case entity.Treasure:
		return 6
	default:
		return 7
	}
}

func useFood(state *entity.GameState, index int) {
	if index < 0 || index >= len(state.Backpack.Food) {
		state.Messages = append(state.Messages, "Invalid food selection")
		return
	}
	food := state.Backpack.Food[index]
	state.Player.Health += food.HealthBoost
	if state.Player.Health > state.Player.MaxHealth {
		state.Player.Health = state.Player.MaxHealth
	}
	state.Backpack.Food = append(state.Backpack.Food[:index], state.Backpack.Food[index+1:]...)
	state.GameSession.FoodConsumed++
	state.Messages = append(state.Messages, "Ate food: +"+strconv.Itoa(food.HealthBoost)+" HP")
}

func usePotion(state *entity.GameState, index int) {
	if index < 0 || index >= len(state.Backpack.Potions) {
		state.Messages = append(state.Messages, "Invalid potion selection")
		return
	}
	potion := state.Backpack.Potions[index]
	applyPotionEffect(state, potion)
	state.Backpack.Potions = append(state.Backpack.Potions[:index], state.Backpack.Potions[index+1:]...)
	state.GameSession.PotionsUsed++
	state.Messages = append(state.Messages, "Used potion")
}

func useScroll(state *entity.GameState, index int) {
	if index < 0 || index >= len(state.Backpack.Scrolls) {
		state.Messages = append(state.Messages, "Invalid scroll selection")
		return
	}
	scroll := state.Backpack.Scrolls[index]

	if scroll.SubType == entity.KeyRed || scroll.SubType == entity.KeyBlue ||
		scroll.SubType == entity.KeyGreen || scroll.SubType == entity.KeyYellow {
		useKey(state, scroll)
	} else {
		applyScrollEffect(state, scroll)
		state.Backpack.Scrolls = append(state.Backpack.Scrolls[:index], state.Backpack.Scrolls[index+1:]...)
		state.GameSession.ScrollsUsed++
		state.Messages = append(state.Messages, "Used scroll")
	}
}

func useWeapon(state *entity.GameState, index int) {
	if index == 0 {
		// Unequip current weapon — бросаем на пол
		if state.Player.CurrentWeapon != nil {
			weapon := state.Player.CurrentWeapon
			if dropItemNearPlayer(state, weapon) {
				state.Messages = append(state.Messages, "Weapon dropped on the floor")
			}
			state.Player.CurrentWeapon = nil
		} else {
			state.Messages = append(state.Messages, "No weapon equipped")
		}
		return
	}

	if index-1 < 0 || index-1 >= len(state.Backpack.Weapons) {
		state.Messages = append(state.Messages, "Invalid weapon selection")
		return
	}

	newWeapon := state.Backpack.Weapons[index-1]

	if state.Player.CurrentWeapon != nil {
		if dropItemNearPlayer(state, state.Player.CurrentWeapon) {
			state.Messages = append(state.Messages, "Old weapon dropped on the floor")
		}
	}

	state.Player.CurrentWeapon = newWeapon
	state.Backpack.Weapons = append(state.Backpack.Weapons[:index-1], state.Backpack.Weapons[index:]...)
	state.Messages = append(state.Messages, "Equipped weapon: +"+strconv.Itoa(newWeapon.StrengthBoost)+" STR")
}

func applyPotionEffect(state *entity.GameState, potion *entity.Item) {
	switch potion.SubType {
	case entity.HealthPotion:
		state.Player.Health += potion.HealthBoost
		if state.Player.Health > state.Player.MaxHealth {
			state.Player.Health = state.Player.MaxHealth
		}
	case entity.AgilityPotion:
		state.Player.Agility += potion.AgilityBoost
	case entity.StrengthPotion:
		state.Player.Strength += potion.StrengthBoost
	}
}

func applyScrollEffect(state *entity.GameState, scroll *entity.Item) {
	switch scroll.SubType {
	case entity.MaxHealthScroll:
		state.Player.MaxHealth += scroll.MaxHealthBoost
		state.Player.Health = state.Player.MaxHealth
	case entity.AgilityScroll:
		state.Player.Agility += scroll.AgilityBoost
	case entity.StrengthScroll:
		state.Player.Strength += scroll.StrengthBoost
	}
}

func useKey(state *entity.GameState, key *entity.Item) {
	for _, door := range state.Level.Doors {
		if door.IsLocked && door.Color == key.SubType {
			door.IsLocked = false
			state.Messages = append(state.Messages, "Unlocked a "+keyName(key.SubType)+" door")
			return
		}
	}
	state.Messages = append(state.Messages, "No matching door to unlock")
}

func keyName(color entity.ItemSubType) string {
	switch color {
	case entity.KeyRed:
		return "red"
	case entity.KeyBlue:
		return "blue"
	case entity.KeyGreen:
		return "green"
	case entity.KeyYellow:
		return "yellow"
	}
	return "unknown"
}

// Функция для получения символа ключа
func GetKeySymbol(keyType entity.ItemSubType) string {
	switch keyType {
	case entity.KeyRed:
		return "R"
	case entity.KeyBlue:
		return "B"
	case entity.KeyGreen:
		return "G"
	case entity.KeyYellow:
		return "Y"
	default:
		return "K"
	}
}

// Функция для получения цвета ключа
func GetKeyColor(keyType entity.ItemSubType) int16 {
	switch keyType {
	case entity.KeyRed:
		return 1 // Красный
	case entity.KeyBlue:
		return 4 // Синий
	case entity.KeyGreen:
		return 2 // Зеленый
	case entity.KeyYellow:
		return 3 // Желтый
	default:
		return 7 // Белый
	}
}

func dropItemNearPlayer(state *entity.GameState, item *entity.Item) bool {
	playerPos := state.Player.Position
	directions := []entity.Position{
		{X: 1, Y: 0}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: -1},
		{X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1},
	}
	for _, dir := range directions {
		x := playerPos.X + dir.X
		y := playerPos.Y + dir.Y
		if canMoveTo(state, x, y) {
			item.Position = entity.Position{X: x, Y: y}
			state.Items = append(state.Items, item)
			return true
		}
	}
	return false
}
