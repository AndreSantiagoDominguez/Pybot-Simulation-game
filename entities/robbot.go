package entities

import (
	"math"
	
	"github.com/hajimehoshi/ebiten/v2"
	"pybot-simulator/utils"
)

type Robot struct {
	Position       utils.Vector2D
	Velocity       utils.Vector2D
	CansCollected  int
	Sprite         *ebiten.Image
	Sprites        map[string]*ebiten.Image
	BatterySprite  *ebiten.Image
	minX, maxX     float64
	minY, maxY     float64
	
	Battery        *Battery
	
	Target         *utils.Vector2D
	Speed          float64
}

func NewRobot(x, y float64, sprite *ebiten.Image) *Robot {
	return &Robot{
		Position:      utils.Vector2D{X: x, Y: y},
		Velocity:      utils.Vector2D{X: 0, Y: 0},
		CansCollected: 0,
		Sprite:        sprite,
		Sprites:       make(map[string]*ebiten.Image),
		Battery:       NewBattery(), // Nueva entidad Battery
		Speed:         2.0,
		Target:        nil,
	}
}

func (r *Robot) Update() {
	// Consumir batería si está en movimiento
	if r.Velocity.X != 0 || r.Velocity.Y != 0 {
		r.Battery.Drain(1.0 / 60.0) // Dividir por TPS (60)
	}
	
	// Si no hay batería, detener movimiento
	if r.Battery.IsEmpty() {
		r.Velocity.X = 0
		r.Velocity.Y = 0
		r.Target = nil
		return
	}
	
	// Moverse hacia el objetivo si existe
	if r.Target != nil {
		dx := r.Target.X - r.Position.X
		dy := r.Target.Y - r.Position.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		
		// Si llegamos al objetivo, detenerse
		if distance < 5.0 {
			r.Target = nil
			r.Velocity.X = 0
			r.Velocity.Y = 0
		} else {
			// Normalizar y aplicar velocidad
			dx /= distance
			dy /= distance
			r.Velocity.X = dx * r.Speed
			r.Velocity.Y = dy * r.Speed
		}
	}
	
	// Actualizar posición
	newX := r.Position.X + r.Velocity.X
	newY := r.Position.Y + r.Velocity.Y
	
	if newX >= r.minX && newX <= r.maxX {
		r.Position.X = newX
	}
	if newY >= r.minY && newY <= r.maxY {
		r.Position.Y = newY
	}
}

func (r *Robot) SetTarget(target utils.Vector2D) {
	if !r.Battery.IsEmpty() {
		r.Target = &target
	}
}

func (r *Robot) ClearTarget() {
	r.Target = nil
	r.Velocity.X = 0
	r.Velocity.Y = 0
}

func (r *Robot) SetVelocity(vx, vy float64) {
	if !r.Battery.IsEmpty() {
		r.Velocity.X = vx
		r.Velocity.Y = vy
	} else {
		r.Velocity.X = 0
		r.Velocity.Y = 0
	}
}

func (r *Robot) SetBounds(minX, maxX, minY, maxY float64) {
	r.minX = minX
	r.maxX = maxX
	r.minY = minY
	r.maxY = maxY
}

func (r *Robot) CollectCan() {
	r.CansCollected++
	r.ClearTarget() // Buscar siguiente lata
}

func (r *Robot) GetBatteryLevel() int {
	return r.Battery.GetLevel()
}

func (r *Robot) GetPosition() utils.Vector2D {
	return r.Position
}

func (r *Robot) SetPosition(x, y float64) {
	r.Position.X = x
	r.Position.Y = y
}

func (r *Robot) GetCansCollected() int {
	return r.CansCollected
}
