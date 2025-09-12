package rules

import "github.com/4vertak/rogue-go/internal/domain/entity"

// TODO: туман войны
func VisibleTiles(level *entity.Level, from entity.Pos, maxDist int, blocks func(p entity.Pos) bool) map[entity.Pos]bool {
	vis := map[entity.Pos]bool{from: true}

	// Перебираем все точки в квадратной области вокруг начальной позиции
	for x := from.X - maxDist; x <= from.X+maxDist; x++ {
		for y := from.Y - maxDist; y <= from.Y+maxDist; y++ {
			p := entity.Pos{X: x, Y: y}
			dx := x - from.X
			dy := y - from.Y
			// Проверяем, находится ли точка в пределах максимальной видимости
			if dx*dx+dy*dy > maxDist*maxDist {
				continue
			}
			// Получаем линию от начальной точки до текущей точки
			line := bresenhamLine(from, p)
			// проверка видимости каждой точки в линии
			for i, point := range line {
				if i == 0 {
					continue
				}
				vis[point] = true
				// Если точка блокирует обзор - прерываем луч
				if blocks(point) {
					break
				}
			}
		}
	}
	return vis
}

// алгорит БрЭзенхэма
func bresenhamLine(from, to entity.Pos) []entity.Pos {
	var points []entity.Pos
	x0, y0 := from.X, from.Y
	x1, y1 := to.X, to.Y

	dx := Abs(x1 - x0)
	dy := Abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		points = append(points, entity.Pos{X: x0, Y: y0})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
	return points
}
