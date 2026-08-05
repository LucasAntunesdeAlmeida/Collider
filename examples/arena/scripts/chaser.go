package scripts

import (
	"math/rand/v2"

	engine "github.com/LucasAntunesdeAlmeida/collider"
)

// Chaser is an engine object extended with its own hunting speed.
type Chaser struct {
	*engine.Object
	speed float64
}

// NewChaser spawns an animated chaser on a random screen edge, hunting
// the player at a speed that varies per individual.
func NewChaser(s *engine.Scene, target *Player) *Chaser {
	c := &Chaser{
		Object: s.Add(engine.Rect(32, 45, nil).AtEdge().Tag("chaser").
			Animation("walk", "sprites/chaser.png", 2, 5)),
		speed: 60 + rand.Float64()*60,
	}
	c.Play("walk")
	c.OnUpdate(func(dt float64) {
		c.MoveToward(target.X, target.Y, c.speed*dt)
	})
	return c
}
