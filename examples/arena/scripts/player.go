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

// NewPlayer creates the animated player in a scene and wires movement.
func NewPlayer(g *engine.Game, s *engine.Scene) *Player {
	p := &Player{
		Object: s.Add(engine.Rect(32, 45, nil).At(400, 300).Tag("player").
			Animation("walk", "sprites/player.png", 2, 8).
			Animation("idle", "sprites/idle.png", 1, 1)),
		HP:    3,
		scene: s,
	}
	p.Play("idle")
	p.OnUpdate(func(dt float64) {
		moving := false
		if g.Key(engine.W) || g.Key(engine.Up) {
			p.Move(0, -260*dt)
			moving = true
		}
		if g.Key(engine.S) || g.Key(engine.Down) {
			p.Move(0, 260*dt)
			moving = true
		}
		if g.Key(engine.A) || g.Key(engine.Left) {
			p.Move(-260*dt, 0)
			moving = true
		}
		if g.Key(engine.D) || g.Key(engine.Right) {
			p.Move(260*dt, 0)
			moving = true
		}
		if moving {
			p.Play("walk")
		} else {
			p.Play("idle")
		}
	})
	return p
}

// TakeDamage applies damage unless the player is in the invulnerability
// window, then opens a new window for one second. Reports whether the
// hit landed.
func (p *Player) TakeDamage(n int) bool {
	if p.invulnerable {
		return false
	}
	p.HP -= n
	p.invulnerable = true
	p.scene.After(1, func() { p.invulnerable = false })
	return true
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
