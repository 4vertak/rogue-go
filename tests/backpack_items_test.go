package tests

import (
	"testing"

	"github.com/4vertak/rogue-go/internal/domain/entity"
)

func TestAddFoodItem(t *testing.T) {
	b := entity.NewBackpack()
	food := entity.Item{Type: "Food", Health: 5}

	ok := b.AddItem(food)
	if !ok {
		t.Fatal("expected item to be added")
	}
	if len(b.Items["Food"]) != 1 {
		t.Errorf("expected 1 food item, got %d", len(b.Items["Food"]))
	}
}

func TestBackpackLimit(t *testing.T) {
	b := entity.NewBackpack()
	food := entity.Item{Type: "Food", Health: 1}

	for i := 0; i < 12; i++ {
		b.AddItem(food)
	}

	if len(b.Items["Food"]) != 9 {
		t.Errorf("expected 9 items max, got %d", len(b.Items["Food"]))
	}
}

func TestTreasureStacks(t *testing.T) {
	b := entity.NewBackpack()
	gold1 := entity.Item{Type: "Treasure", Value: 10}
	gold2 := entity.Item{Type: "Treasure", Value: 5}

	b.AddItem(gold1)
	b.AddItem(gold2)

	if len(b.Items["Treasure"]) != 1 {
		t.Errorf("expected 1 treasure slot, got %d", len(b.Items["Treasure"]))
	}
	if b.Items["Treasure"][0].Value != 15 {
		t.Errorf("expected total treasure 15, got %d", b.Items["Treasure"][0].Value)
	}
}

func TestUseFood(t *testing.T) {
	b := entity.NewBackpack()
	food := entity.Item{Type: "Food", Health: 5}
	b.AddItem(food)

	item, ok := b.UseItem("Food", 0)
	if !ok {
		t.Fatal("expected to use food")
	}
	if item.Health != 5 {
		t.Errorf("expected food with 5 health, got %+v", item)
	}
	if len(b.Items["Food"]) != 0 {
		t.Errorf("expected food removed after use, got %d left", len(b.Items["Food"]))
	}
}

func TestUseInvalidIndex(t *testing.T) {
	b := entity.NewBackpack()
	_, ok := b.UseItem("Food", 0)
	if ok {
		t.Fatal("expected use to fail with empty inventory")
	}
}
