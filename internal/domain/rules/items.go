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
			return "Health Potion (HP+" + strconv.Itoa(item.HealthBoost) + ")"
		case entity.AgilityPotion:
			return "Agility Potion (AGI+" + strconv.Itoa(item.AgilityBoost) + ")"
		case entity.StrengthPotion:
			return "Strength Potion (STR+" + strconv.Itoa(item.StrengthBoost) + ")"
		default:
			return "Potion"
		}
	case entity.Scroll:
		switch item.SubType {
		case entity.MaxHealthScroll:
			return "Max Health Scroll (MaxHP+" + strconv.Itoa(item.MaxHealthBoost) + ")"
		case entity.AgilityScroll:
			return "Agility Scroll (AGI+" + strconv.Itoa(item.AgilityBoost) + ")"
		case entity.StrengthScroll:
			return "Strength Scroll (STR+" + strconv.Itoa(item.StrengthBoost) + ")"
		case entity.KeyRed:
			return "Red Key"
		case entity.KeyBlue:
			return "Blue Key"
		case entity.KeyGreen:
			return "Green Key"
		case entity.KeyYellow:
			return "Yellow Key"
		default:
			return "Mystery Scroll"
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
		return 7
	case entity.Food:
		return 14
	case entity.Potion:
		return 8
	case entity.Scroll:
		return 14
	case entity.Treasure:
		return 13
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

	// Добавляем сообщение о использовании
	switch potion.SubType {
	case entity.HealthPotion:
		state.Messages = append(state.Messages, "Used Health Potion: +"+strconv.Itoa(potion.HealthBoost)+" HP")
	case entity.AgilityPotion:
		state.Messages = append(state.Messages, "Used Agility Potion: +"+strconv.Itoa(potion.AgilityBoost)+" AGI")
	case entity.StrengthPotion:
		state.Messages = append(state.Messages, "Used Strength Potion: +"+strconv.Itoa(potion.StrengthBoost)+" STR")
	}
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
		state.Backpack.Scrolls = append(state.Backpack.Scrolls[:index], state.Backpack.Scrolls[index+1:]...)
		state.Messages = append(state.Messages, "Used "+keyName(scroll.SubType))
	} else {
		applyScrollEffect(state, scroll)
		state.Backpack.Scrolls = append(state.Backpack.Scrolls[:index], state.Backpack.Scrolls[index+1:]...)
		state.GameSession.ScrollsUsed++

		// Добавляем сообщение о использовании с правильным названием
		switch scroll.SubType {
		case entity.MaxHealthScroll:
			state.Messages = append(state.Messages, "Used Max Health Scroll: +"+strconv.Itoa(scroll.MaxHealthBoost)+" Max HP")
		case entity.AgilityScroll:
			state.Messages = append(state.Messages, "Used Agility Scroll: +"+strconv.Itoa(scroll.AgilityBoost)+" AGI")
		case entity.StrengthScroll:
			state.Messages = append(state.Messages, "Used Strength Scroll: +"+strconv.Itoa(scroll.StrengthBoost)+" STR")
		default:
			state.Messages = append(state.Messages, "Used scroll")
		}
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
		// Создаем временный эффект
		effect := &entity.Effect{
			Type:        entity.HealthEffect,
			Duration:    0, // Мгновенный эффект
			Value:       potion.HealthBoost,
			Description: "Health restored",
		}
		state.Player.ActiveEffects = append(state.Player.ActiveEffects, effect)

	case entity.AgilityPotion:
		state.Player.Agility += potion.AgilityBoost
		// Создаем временный эффект
		effect := &entity.Effect{
			Type:        entity.AgilityEffect,
			Duration:    10, // Длительность в ходах
			Value:       potion.AgilityBoost,
			Description: "Agility boosted",
		}
		state.Player.ActiveEffects = append(state.Player.ActiveEffects, effect)

	case entity.StrengthPotion:
		state.Player.Strength += potion.StrengthBoost
		// Создаем временный эффект
		effect := &entity.Effect{
			Type:        entity.StrengthEffect,
			Duration:    10, // Длительность в ходах
			Value:       potion.StrengthBoost,
			Description: "Strength boosted",
		}
		state.Player.ActiveEffects = append(state.Player.ActiveEffects, effect)
	}
}

func applyScrollEffect(state *entity.GameState, scroll *entity.Item) {
	switch scroll.SubType {
	case entity.MaxHealthScroll:
		state.Player.MaxHealth += scroll.MaxHealthBoost
		state.Player.Health = state.Player.MaxHealth
		// Создаем временный эффект
		effect := &entity.Effect{
			Type:        entity.MaxHealthEffect,
			Duration:    0, // Постоянный эффект
			Value:       scroll.MaxHealthBoost,
			Description: "Max health increased",
		}
		state.Player.ActiveEffects = append(state.Player.ActiveEffects, effect)

	case entity.AgilityScroll:
		state.Player.Agility += scroll.AgilityBoost
		// Создаем временный эффект
		effect := &entity.Effect{
			Type:        entity.AgilityEffect,
			Duration:    0, // Постоянный эффект
			Value:       scroll.AgilityBoost,
			Description: "Agility permanently increased",
		}
		state.Player.ActiveEffects = append(state.Player.ActiveEffects, effect)

	case entity.StrengthScroll:
		state.Player.Strength += scroll.StrengthBoost
		// Создаем временный эффект
		effect := &entity.Effect{
			Type:        entity.StrengthEffect,
			Duration:    0, // Постоянный эффект
			Value:       scroll.StrengthBoost,
			Description: "Strength permanently increased",
		}
		state.Player.ActiveEffects = append(state.Player.ActiveEffects, effect)
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
		return 10 // Красный
	case entity.KeyBlue:
		return 11 // Синий
	case entity.KeyGreen:
		return 12 // Зеленый
	case entity.KeyYellow:
		return 13 // Желтый
	default:
		return 14 // Белый
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

/*
// Функция для получения символа двери
func GetDoorSymbol(door *entity.Door) string {
    switch door.Color {
    case entity.KeyRed:
        return "R" // Красная дверь
    case entity.KeyBlue:
        return "B" // Синяя дверь
    case entity.KeyGreen:
        return "G" // Зеленая дверь
    case entity.KeyYellow:
        return "Y" // Желтая дверь
    default:
        return "D" // Дверь по умолчанию
    }
}
*/

// Функция для получения символа двери (цифровая версия)
func GetDoorSymbol(door *entity.Door) string {
	switch door.Color {
	case entity.KeyRed:
		return "1" // Красная дверь
	case entity.KeyBlue:
		return "2" // Синяя дверь
	case entity.KeyGreen:
		return "3" // Зеленая дверь
	case entity.KeyYellow:
		return "4" // Желтая дверь
	default:
		return "0" // Дверь по умолчанию
	}
}

// Функция для получения цвета двери
func GetDoorColor(door *entity.Door) int16 {
	switch door.Color {
	case entity.KeyRed:
		return 10 // Красный
	case entity.KeyBlue:
		return 11 // Синий
	case entity.KeyGreen:
		return 12 // Зеленый
	case entity.KeyYellow:
		return 13 // Желтый
	default:
		return 1 // Белый
	}
}
