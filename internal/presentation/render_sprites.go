package presentation

import (
	"math"
	"rogue-go/internal/domain/entity"
	"rogue-go/internal/presentation/sprites"
	"sort"
)

func (gi *GameInterface) drawSpritesOptimized(enemies []*entity.Enemy, items []*entity.Item, zBuffer []float64, width, height int) {
	player := gi.game.State.Player
	posX := float64(player.Position.X) + 0.5
	posY := float64(player.Position.Y) + 0.5
	dirX := gi.game.State.PlayerDir.X
	dirY := gi.game.State.PlayerDir.Y
	planeX := gi.game.State.PlayerPlane.X
	planeY := gi.game.State.PlayerPlane.Y

	sprites := gi.prepareSprites(enemies, items, posX, posY)

	for _, sprite := range sprites {
		spriteX := sprite.x - posX
		spriteY := sprite.y - posY

		invDet := 1.0 / (planeX*dirY - dirX*planeY)
		transformX := invDet * (dirY*spriteX - dirX*spriteY)
		transformY := invDet * (-planeY*spriteX + planeX*spriteY)

		if transformY <= 0 {
			continue
		}

		spriteScreenX := int(float64(width) / 2 * (1 + transformX/transformY))
		spriteHeight := int(float64(height) / transformY)
		spriteWidth := spriteHeight

		drawStartY := -spriteHeight/2 + height/2
		if drawStartY < 0 {
			drawStartY = 0
		}
		drawEndY := spriteHeight/2 + height/2
		if drawEndY >= height {
			drawEndY = height - 1
		}

		drawStartX := -spriteWidth/2 + spriteScreenX
		if drawStartX < 0 {
			drawStartX = 0
		}
		drawEndX := spriteWidth/2 + spriteScreenX
		if drawEndX >= width {
			drawEndX = width - 1
		}

		for stripe := drawStartX; stripe < drawEndX; stripe++ {
			if stripe < 0 || stripe >= width || transformY > zBuffer[stripe] {
				continue
			}

			texX := int(256*(stripe-(-spriteWidth/2+spriteScreenX))*len(sprite.data[0])/spriteWidth) / 256

			for y := drawStartY; y < drawEndY; y++ {
				d := (y - (-spriteHeight/2 + height/2)) * 256 / spriteHeight
				texY := ((d * len(sprite.data)) / 256)

				if texX >= 0 && texX < len(sprite.data[0]) && texY >= 0 && texY < len(sprite.data) {
					color := sprite.data[texY][texX]
					if color != ' ' {
						gi.gameWindow.ColorOn(sprite.color)
						gi.gameWindow.MovePrint(y+1, stripe+1, string(color))
						gi.gameWindow.ColorOff(sprite.color)
					}
				}
			}
		}
	}
}

type spriteData struct {
	x, y  float64
	data  [][]rune
	color int16
	dist  float64
}

func (gi *GameInterface) prepareSprites(enemies []*entity.Enemy, items []*entity.Item, posX, posY float64) []spriteData {
	var sprites []spriteData

	// Враги
	for _, enemy := range enemies {
		if !enemy.IsVisible {
			continue
		}
		dist := math.Pow(float64(enemy.Position.X)+0.5-posX, 2) +
			math.Pow(float64(enemy.Position.Y)+0.5-posY, 2)
		sprite := spriteData{
			x:     float64(enemy.Position.X) + 0.5,
			y:     float64(enemy.Position.Y) + 0.5,
			data:  gi.GetEnemySprite(enemy),
			color: gi.game.GetEnemyColor(enemy),
			dist:  dist,
		}
		sprites = append(sprites, sprite)
	}

	// Предметы
	for _, item := range items {
		dist := math.Pow(float64(item.Position.X)+0.5-posX, 2) +
			math.Pow(float64(item.Position.Y)+0.5-posY, 2)
		sprite := spriteData{
			x:     float64(item.Position.X) + 0.5,
			y:     float64(item.Position.Y) + 0.5,
			data:  gi.GetItemSprite(item),
			color: gi.game.GetItemColor(item),
			dist:  dist,
		}
		sprites = append(sprites, sprite)
	}
	// лестинца
	sort.Slice(sprites, func(i, j int) bool {
		return sprites[i].dist > sprites[j].dist
	})

	return sprites
}

func (gi *GameInterface) getVisibleEnemies() []*entity.Enemy {
	var visible []*entity.Enemy
	for _, enemy := range gi.game.State.Enemies {
		if enemy.IsVisible && gi.game.State.Level.Visible[enemy.Position.X][enemy.Position.Y] {
			visible = append(visible, enemy)
		}
	}
	return visible
}

func (gi *GameInterface) getVisibleItems() []*entity.Item {
	var visible []*entity.Item
	for _, item := range gi.game.State.Items {
		if gi.game.State.Level.Visible[item.Position.X][item.Position.Y] {
			visible = append(visible, item)
		}
	}
	return visible
}

func maxDoorSpriteWidth() int {
	maxW := 0
	for _, row := range sprites.DoorSprite {
		if len(row) > maxW {
			maxW = len(row)
		}
	}
	return maxW
}

func (gi *GameInterface) getDoorSymbolSafe(x, y int) string {
	if y < 0 || y >= len(sprites.DoorSprite) {
		return " "
	}
	row := sprites.DoorSprite[y]
	if x < 0 || x >= len(row) {
		return " "
	}
	return string(row[x])
}
