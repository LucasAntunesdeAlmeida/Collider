package collider

import (
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/LucasAntunesdeAlmeida/collider/internal/physics"
)

// Object is anything that lives in a scene: the player, a wall, a bullet,
// a button. Position is the object's center. Velocity applies every frame.
type Object struct {
	X, Y   float64
	Vx, Vy float64

	// Data is a free field for game state (a card face, hit points, anything).
	Data any

	w, h    float64
	sizeSet bool
	tag     string
	solid   bool
	gravity bool

	// hitW, hitH is the collision box when hitbox is set (Hitbox);
	// otherwise the object collides with its drawn size w, h.
	hitW, hitH float64
	hitbox     bool

	grounded bool
	life     float64
	hasLife  bool
	atEdge   bool

	spritePath string
	img        *ebiten.Image
	fill       Color

	isText   bool
	isButton bool
	visual   bool
	fixed    bool
	layer    int

	// Drawing effects. alpha is stored as transparency (1 - opacity)
	// so the zero value draws fully opaque.
	alpha      float64
	tint       Color
	rotation   float64
	flipX      bool
	flashColor Color
	flashLeft  float64

	textStr   string
	fontPaths []string // Font: the first draws what it can, the rest fill in
	textSize  float64
	textColor Color

	// face measures and draws textStr (textFace); faceText is the text
	// it was chosen for and faceMask the fonts it combines. Font and
	// TextSize clear it.
	face     text.Face
	faceText string
	faceMask uint64

	anims    map[string]*animation
	curAnim  string
	animTime float64
	animOnce bool

	dead  bool
	scene *Scene
	saved *objState

	onUpdate        []func(dt float64)
	onClick         []func()
	onCollision     []func(*Object)
	onCollisionWith map[string][]func(*Object)
}

// objState is the snapshot Restart rolls an object back to.
type objState struct {
	x, y, vx, vy float64
	w, h         float64
	hitW, hitH   float64
	hitbox       bool
	data         any
	textStr      string
	spritePath   string
	img          *ebiten.Image
	life         float64
	hasLife      bool
	layer        int
	alpha        float64
	tint         Color
	rotation     float64
	flipX        bool
}

// Sprite creates an object from an image file. The collider defaults to
// the image bounds; override with Size (drawing too) or Hitbox
// (collisions only). The image itself is loaded from the game's asset
// cache when the object is added to a scene.
func Sprite(path string) *Object {
	return &Object{spritePath: path}
}

// Rect creates a colored rectangle object.
func Rect(w, h float64, c Color) *Object {
	return &Object{w: w, h: h, fill: c}
}

// At places the object's center. Chainable.
func (o *Object) At(x, y float64) *Object {
	o.X, o.Y = x, y
	return o
}

// AtEdge places the object at a random point on a random screen edge
// when it is added to a scene. Chainable.
func (o *Object) AtEdge() *Object {
	o.atEdge = true
	return o
}

// Tag labels the object for tag-based collision rules. Chainable.
func (o *Object) Tag(name string) *Object {
	o.tag = name
	if o.scene != nil {
		o.scene.index.invalidate() // re-tagged mid-frame: Near and Touching see it now
	}
	return o
}

// Solid marks the object as solid geometry: the engine pushes non-solid
// objects out of it instead of letting them pass through. Chainable.
func (o *Object) Solid() *Object {
	o.solid = true
	return o
}

// WithGravity makes the object fall with the scene's gravity. Chainable.
func (o *Object) WithGravity() *Object {
	o.gravity = true
	return o
}

// Size overrides the drawn size, which is also the collider unless
// Hitbox sets one apart. Chainable.
func (o *Object) Size(w, h float64) *Object {
	o.w, o.h = w, h
	o.sizeSet = true
	return o
}

// Hitbox sets the collision box apart from the drawn size: a w x h box
// centered on the object's position, used by collisions, solids,
// Touching and Near, and reported to agents as the object's w and h.
// Drawing, culling and clicks keep the sprite (or Size) bounds, since
// players click what they see. The fair-contact fix: a 42px hero with
// a 26px hitbox is only hurt when a chaser really reaches it. Without
// Hitbox the collider is the drawn size, as before. Chainable, and
// callable any time.
func (o *Object) Hitbox(w, h float64) *Object {
	o.hitW, o.hitH = w, h
	o.hitbox = true
	if o.scene != nil {
		o.scene.index.invalidate() // Near and Touching see the new box now
	}
	return o
}

// Layer sets the draw order: lower layers draw first, higher ones on
// top, and objects on the same layer keep the order they were added
// in. Clicks go to the topmost object. The default layer is 0; any int
// works, negative included. Chainable, and callable any time.
func (o *Object) Layer(n int) *Object {
	o.layer = n
	return o
}

// LifeTime destroys the object automatically after this many seconds.
// Chainable, and callable after Add too (bullets, particles).
func (o *Object) LifeTime(sec float64) *Object {
	o.life = sec
	o.hasLife = true
	return o
}

// resolve loads deferred content and placement that need the game.
// Called by Scene.Add, so sprites are decoded once per game.
func (o *Object) resolve(g *Game) {
	if o.atEdge {
		o.atEdge = false
		w, h := g.Width(), g.Height()
		switch rand.IntN(4) {
		case 0:
			o.X, o.Y = rand.Float64()*w, 0
		case 1:
			o.X, o.Y = rand.Float64()*w, h
		case 2:
			o.X, o.Y = 0, rand.Float64()*h
		case 3:
			o.X, o.Y = w, rand.Float64()*h
		}
	}
	if o.spritePath != "" && o.img == nil {
		o.img = g.assets.Image(o.spritePath)
		if !o.sizeSet {
			b := o.img.Bounds()
			o.w, o.h = float64(b.Dx()), float64(b.Dy())
		}
	}
	for _, a := range o.anims {
		a.load(g, o)
	}
}

// Move shifts the object by a delta. Overlaps with solid objects are
// resolved at the end of the frame.
func (o *Object) Move(dx, dy float64) {
	o.X += dx
	o.Y += dy
}

// MoveToward moves the object dist pixels toward a point, stopping
// exactly on it instead of overshooting.
func (o *Object) MoveToward(x, y, dist float64) {
	dx, dy := x-o.X, y-o.Y
	d := math.Hypot(dx, dy)
	if d <= dist || d == 0 {
		o.X, o.Y = x, y
		return
	}
	o.X += dx / d * dist
	o.Y += dy / d * dist
}

// VelocityToward points the object's velocity at a target with the
// given speed. The bullet helper.
func (o *Object) VelocityToward(x, y, speed float64) {
	dx, dy := x-o.X, y-o.Y
	d := math.Hypot(dx, dy)
	if d == 0 {
		o.Vx, o.Vy = 0, 0
		return
	}
	o.Vx, o.Vy = dx/d*speed, dy/d*speed
}

// Grounded reports whether the object is resting on a solid object.
// The platformer jump check.
func (o *Object) Grounded() bool {
	return o.grounded
}

// SetSprite swaps the object's image at runtime.
func (o *Object) SetSprite(path string) {
	o.spritePath = path
	o.img = nil
	if o.scene != nil {
		o.resolve(o.scene.game)
	}
}

// Destroy removes the object at the end of the frame. Always safe to
// call from inside callbacks.
func (o *Object) Destroy() {
	o.dead = true
}

// OnUpdate runs every frame with the elapsed time in seconds.
func (o *Object) OnUpdate(fn func(dt float64)) {
	o.onUpdate = append(o.onUpdate, fn)
}

// OnClick fires when the object is clicked (a point-vs-bounds collision).
func (o *Object) OnClick(fn func()) {
	o.onClick = append(o.onClick, fn)
}

// OnCollision fires once when the object starts touching another object.
func (o *Object) OnCollision(fn func(other *Object)) {
	o.onCollision = append(o.onCollision, fn)
}

// OnCollisionWith fires once when the object starts touching an object
// carrying this tag.
func (o *Object) OnCollisionWith(tag string, fn func(other *Object)) {
	if o.onCollisionWith == nil {
		o.onCollisionWith = map[string][]func(*Object){}
	}
	o.onCollisionWith[tag] = append(o.onCollisionWith[tag], fn)
}

func (o *Object) fireCollision(other *Object) {
	for _, fn := range o.onCollision {
		fn(other)
	}
	if other.tag != "" {
		for _, fn := range o.onCollisionWith[other.tag] {
			fn(other)
		}
	}
}

func (o *Object) saveState() {
	o.saved = &objState{
		x: o.X, y: o.Y, vx: o.Vx, vy: o.Vy,
		w: o.w, h: o.h,
		hitW: o.hitW, hitH: o.hitH, hitbox: o.hitbox,
		data:       o.Data,
		textStr:    o.textStr,
		spritePath: o.spritePath,
		img:        o.img,
		life:       o.life,
		hasLife:    o.hasLife,
		layer:      o.layer,
		alpha:      o.alpha,
		tint:       o.tint,
		rotation:   o.rotation,
		flipX:      o.flipX,
	}
}

func (o *Object) restoreState() {
	if o.saved == nil {
		return
	}
	s := o.saved
	o.X, o.Y, o.Vx, o.Vy = s.x, s.y, s.vx, s.vy
	o.w, o.h = s.w, s.h
	o.hitW, o.hitH, o.hitbox = s.hitW, s.hitH, s.hitbox
	o.Data = s.data
	o.textStr = s.textStr
	o.spritePath = s.spritePath
	o.img = s.img
	o.life = s.life
	o.hasLife = s.hasLife
	o.layer = s.layer
	o.alpha, o.tint = s.alpha, s.tint
	o.rotation, o.flipX = s.rotation, s.flipX
	o.flashLeft = 0
	o.dead = false
	o.grounded = false
}

// box is the collision box: the Hitbox when one is set, else the drawn
// bounds. Everything physical (collisions, solids, Near, Touching,
// observations) goes through it.
func (o *Object) box() physics.Box {
	if o.hitbox {
		return physics.Box{X: o.X, Y: o.Y, W: o.hitW, H: o.hitH}
	}
	return o.bounds()
}

// bounds is the drawn box (sprite or Size), which clicks hit.
func (o *Object) bounds() physics.Box {
	return physics.Box{X: o.X, Y: o.Y, W: o.w, H: o.h}
}

func (o *Object) contains(x, y float64) bool {
	return physics.Contains(o.bounds(), x, y)
}
