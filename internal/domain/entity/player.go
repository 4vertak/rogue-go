package entity

type Player struct {
	Character
	Pos      Pos
	Gold     int
	Backpack *Backpack
}

func DefaultPlayer() Player {
	return Player{
		Character: Character{
			Stats: Stats{MaxHP: 20, HP: 20, DEX: 5, STR: 5}},
		Pos: Pos{X: 2, Y: 2}, Gold: 0,
		Backpack: NewBackpack()}
}

// PickItem adds an item to the player's backpack
func (p *Player) PickItem(item Item) bool {
	return p.Backpack.AddItem(item)
}

// UseItem applies the effect of an item from the backpack
func (p *Player) UseItem(itemType string, index int) bool {
	item, ok := p.Backpack.UseItem(itemType, index)
	if !ok {
		return false
	}

	switch item.Type {
	case "Food":
		p.Stats.HP += item.Health
		if p.Stats.HP > p.Stats.MaxHP {
			p.Stats.HP = p.Stats.MaxHP
		}
	case "Elixir":
		p.Stats.DEX += item.Dexterity
		p.Stats.STR += item.Strength
		p.Stats.MaxHP += item.MaxHP
		p.Stats.HP += item.MaxHP
	case "Scroll":
		p.Stats.DEX += item.Dexterity
		p.Stats.STR += item.Strength
		p.Stats.MaxHP += item.MaxHP
		p.Stats.HP += item.MaxHP
	case "Weapon":
		p.Weapon = item.Weapon
	case "Treasure":
		p.Gold += item.Value
	}
	return true
}
