// Package scripts holds this game's custom types. This is the pattern
// for extending Collider: embed *engine.Object in your own struct, add
// fields and methods, and the result is both a full engine object (Move,
// OnCollisionWith, velocity...) and your own domain type.
package scripts

import (
	engine "github.com/LucasAntunesdeAlmeida/collider"
)

// Player is an engine object extended with hit points and damage rules.
type Player struct {
	*engine.Object
	HP int

	scene        *engine.Scene
	invulnerable bool
}

// NewPlayer creates the player in a scene and wires its movement.
func NewPlayer(g *engine.Game, s *engine.Scene) *Player {
	p := &Player{
		Object: s.Add(engine.Rect(36, 36, engine.Blue).At(400, 300)),
		HP:     3,
		scene:  s,
	}
	p.OnUpdate(func(dt float64) {
		if g.Key(engine.W) {
			p.Move(0, -260*dt)
		}
		if g.Key(engine.S) {
			p.Move(0, 260*dt)
		}
		if g.Key(engine.A) {
			p.Move(-260*dt, 0)
		}
		if g.Key(engine.D) {
			p.Move(260*dt, 0)
		}
	})
	return p
}

// TakeDamage applies damage unless the player is in the invulnerability
// window, then opens a new window for one second.
func (p *Player) TakeDamage(n int) {
	if p.invulnerable {
		return
	}
	p.HP -= n
	p.invulnerable = true
	p.scene.After(1, func() { p.invulnerable = false })
}

// Knockback shoves the player away from a point of impact.
func (p *Player) Knockback(fromX, fromY float64) {
	p.MoveToward(fromX, fromY, -60)
}

// Reset restores the player's stats for a fresh round. Position is
// handled by the engine's Restart snapshot; stats are the game's job.
func (p *Player) Reset() {
	p.HP = 3
	p.invulnerable = false
}
