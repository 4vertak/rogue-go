package entity

type Backpack struct {
	Items map[string][]Item
}

func NewBackpack() *Backpack {
	return &Backpack{Items: make(map[string][]Item)}
}

func (b *Backpack) AddItem(item Item) bool {
	if item.Type == "Treasure" {
		if len(b.Items["Treasure"]) == 0 {
			b.Items["Treasure"] = []Item{item}
		} else {
			b.Items["Treasure"][0].Value += item.Value
		}
		return true
	}

	if len(b.Items[item.Type]) >= 9 {
		return false
	}
	b.Items[item.Type] = append(b.Items[item.Type], item)
	return true
}

func (b *Backpack) UseItem(itemType string, index int) (Item, bool) {
	items, ok := b.Items[itemType]
	if !ok || index < 0 || index >= len(items) {
		return Item{}, false
	}
	item := items[index]
	if itemType != "Treasure" && itemType != "Weapon" {
		b.Items[itemType] = append(items[:index], items[index+1:]...)
	}
	return item, true
}
