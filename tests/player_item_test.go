package tests

import (
	"testing"

	"github.com/4vertak/rogue-go/internal/domain/entity"
)

func TestPlayerPickAndUseFood(t *testing.T) {
	p := entity.DefaultPlayer()
	p.Stats.HP = 10
	food := entity.Item{Type: "Food", Health: 5}
	p.PickItem(food)

	ok := p.UseItem("Food", 0)
	if !ok {
		t.Fatal("expected food to be used")
	}
	if p.Stats.HP != 15 {
		t.Errorf("expected HP=15, got %d", p.Stats.HP)
	}
}

func TestPlayerUseScroll(t *testing.T) {
	p := entity.DefaultPlayer()
	scroll := entity.Item{Type: "Scroll", Strength: 2, MaxHP: 3}
	p.PickItem(scroll)

	oldStr := p.Stats.STR
	oldMaxHP := p.Stats.MaxHP
	p.UseItem("Scroll", 0)

	if p.Stats.STR != oldStr+2 {
		t.Errorf("expected STR=%d, got %d", oldStr+2, p.Stats.STR)
	}
	if p.Stats.MaxHP != oldMaxHP+3 {
		t.Errorf("expected MaxHP=%d, got %d", oldMaxHP+3, p.Stats.MaxHP)
	}
}

func TestPlayerUseWeapon(t *testing.T) {
	p := entity.DefaultPlayer()
	sword := entity.Item{Type: "Weapon", Weapon: &entity.Weapon{Name: "Sword", DamageMin: 2, DamageMax: 4}}
	p.PickItem(sword)

	p.UseItem("Weapon", 0)
	if p.Weapon == nil || p.Weapon.Name != "Sword" {
		t.Errorf("expected Sword equipped, got %+v", p.Weapon)
	}
}

func TestPlayerCollectTreasure(t *testing.T) {
	p := entity.DefaultPlayer()
	treasure := entity.Item{Type: "Treasure", Value: 50}
	p.PickItem(treasure)

	p.UseItem("Treasure", 0)
	if p.Gold != 50 {
		t.Errorf("expected Gold=50, got %d", p.Gold)
	}
}
