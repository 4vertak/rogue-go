package rules

import (
    "math"
    "rogue-go/internal/domain/entity"
    "rogue-go/internal/domain/utils"
)

// GetNearestInteractable возвращает ближайший объект для взаимодействия
func GetNearestInteractable(state *entity.GameState) (interface{}, float64) {
    playerPos := state.Player.Position
    var nearest interface{}
    minDist := 2.0 // Максимальная дистанция взаимодействия
    
    // Проверяем врагов
    for _, enemy := range state.Enemies {
        if enemy.IsVisible {
            dist := math.Sqrt(math.Pow(float64(enemy.Position.X-playerPos.X), 2) + 
                           math.Pow(float64(enemy.Position.Y-playerPos.Y), 2))
            if dist < minDist {
                minDist = dist
                nearest = enemy
            }
        }
    }
    
    // Проверяем предметы (кроме ключей - они не автоподбираются)
    for _, item := range state.Items {
        if item.Type == entity.Scroll && (item.SubType == entity.KeyRed || 
           item.SubType == entity.KeyBlue || item.SubType == entity.KeyGreen || 
           item.SubType == entity.KeyYellow) {
            continue // Ключи не автоподбираются
        }
        
        dist := math.Sqrt(math.Pow(float64(item.Position.X-playerPos.X), 2) + 
                       math.Pow(float64(item.Position.Y-playerPos.Y), 2))
        if dist < minDist {
            minDist = dist
            nearest = item
        }
    }
    
    // Проверяем двери
    for _, door := range state.Level.Doors {
        dist := math.Sqrt(math.Pow(float64(door.Position.X-playerPos.X), 2) + 
                       math.Pow(float64(door.Position.Y-playerPos.Y), 2))
        if dist < minDist {
            minDist = dist
            nearest = door
        }
    }
    
    return nearest, minDist
}

// InteractWithNearest взаимодействует с ближайшим объектом
func InteractWithNearest(state *entity.GameState) {
    target, dist := GetNearestInteractable(state)
    
    if dist > 1.5 { // Дистанция взаимодействия
        utils.AddMessage(state, "Nothing to interact with nearby")
        return
    }
    
    switch t := target.(type) {
    case *entity.Enemy:
        // Находим индекс врага
        for i, enemy := range state.Enemies {
            if enemy == t {
                Attack(state, i)
                break
            }
        }
    case *entity.Item:
        // Для ключей - специальная обработка
        if t.Type == entity.Scroll && (t.SubType == entity.KeyRed || 
           t.SubType == entity.KeyBlue || t.SubType == entity.KeyGreen || 
           t.SubType == entity.KeyYellow) {
            // Подбираем ключ вручную
            if canAddToBackpack(state, t) {
                addToBackpack(state, t)
                // Удаляем предмет из уровня
                for i, item := range state.Items {
                    if item == t {
                        state.Items = append(state.Items[:i], state.Items[i+1:]...)
                        break
                    }
                }
                utils.AddMessage(state, "Picked up "+utils.GetKeyName(t.SubType))
            }
        }
    case *entity.Door:
        InteractWithDoor3D(state, t)
    default:
        utils.AddMessage(state, "Cannot interact with this object")
    }
}

// InteractWithDoor3D взаимодействует с конкретной дверью
func InteractWithDoor3D(state *entity.GameState, door *entity.Door) {
    if door.IsLocked {
        // Проверяем, есть ли ключ в инвентаре
        for _, item := range state.Backpack.Scrolls {
            if item.Type == entity.Scroll && item.SubType == door.Color {
                door.IsLocked = false
                utils.AddMessage(state, "Unlocked door with "+utils.GetKeyName(door.Color))
                // Удаляем использованный ключ
                for i, scroll := range state.Backpack.Scrolls {
                    if scroll == item {
                        state.Backpack.Scrolls = append(state.Backpack.Scrolls[:i], 
                                                      state.Backpack.Scrolls[i+1:]...)
                        break
                    }
                }
                return
            }
        }
        utils.AddMessage(state, "Door is locked - need "+utils.GetKeyName(door.Color))
    } else {
        utils.AddMessage(state, "Door is already unlocked")
    }
}