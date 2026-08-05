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

// NewChaser spawns a chaser on a random screen edge, hunting the player
// at a speed that varies per individual.
func NewChaser(s *engine.Scene, target *Player) *Chaser {
	c := &Chaser{
		Object: s.Add(engine.Rect(28, 28, engine.Red).AtEdge().Tag("chaser")),
		speed:  60 + rand.Float64()*60,
	}
	c.OnUpdate(func(dt float64) {
		c.MoveToward(target.X, target.Y, c.speed*dt)
	})
	return c
}
