package rules

import "github.com/4vertak/rogue-go/internal/domain/entity"

func TryMove(p *entity.Player, dx, dy int, lvl *entity.Level) {
	nx, ny := p.Pos.X+dx, p.Pos.Y+dy
	if nx < 0 || ny < 0 || ny >= lvl.H || nx >= lvl.W {
		return
	}
	t := lvl.Tiles[ny][nx]
	if t == entity.Floor || t == entity.Exit {
		p.Pos.X, p.Pos.Y = nx, ny
	}
}
