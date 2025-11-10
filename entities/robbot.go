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
	TotalWeight    float64
	Sprite         *ebiten.Image
	Sprites        map[string]*ebiten.Image
	BatterySprite  *ebiten.Image
	minX, maxX     float64
	minY, maxY     float64
	
	// Sistema de batería
	Battery        float64 // 0.0 a 120.0 (2 minutos en segundos)
	MaxBattery     float64 // 120.0
	BatteryDrain   float64 // Consumo por segundo
	
	// Sistema de navegación automática
	Target         *utils.Vector2D
	Speed          float64
}

func NewRobot(x, y float64, sprite *ebiten.Image) *Robot {
	return &Robot{
		Position:      utils.Vector2D{X: x, Y: y},
		Velocity:      utils.Vector2D{X: 0, Y: 0},
		CansCollected: 0,
		TotalWeight:   0.0,
		Sprite:        sprite,
		Sprites:       make(map[string]*ebiten.Image),
		Battery:       120.0, // Inicia con batería completa (2 min)
		MaxBattery:    120.0,
		BatteryDrain:  1.0, // 1 segundo por segundo
		Speed:         2.0,
		Target:        nil,
	}
}

func (r *Robot) Update() {
	// Consumir batería si está en movimiento
	if r.Velocity.X != 0 || r.Velocity.Y != 0 {
		r.Battery -= r.BatteryDrain / 60.0 // Dividir por TPS (60)
		if r.Battery < 0 {
			r.Battery = 0
		}
	}
	
	// Si no hay batería, detener movimiento
	if r.Battery <= 0 {
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
	if r.Battery > 0 {
		r.Target = &target
	}
}

func (r *Robot) ClearTarget() {
	r.Target = nil
	r.Velocity.X = 0
	r.Velocity.Y = 0
}

func (r *Robot) SetVelocity(vx, vy float64) {
	if r.Battery > 0 {
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

func (r *Robot) CollectCan(can *Can) {
	r.CansCollected++
	r.TotalWeight += can.Weight
	r.ClearTarget() // Buscar siguiente lata
}

func (r *Robot) Recharge() {
	r.Battery = r.MaxBattery
}

func (r *Robot) GetBatteryLevel() int {
	// Retorna nivel de batería de 0 a 4 (para los 4 frames que tenemos en el sprite)
	percentage := r.Battery / r.MaxBattery
	if percentage > 0.75 {
		return 4
	} else if percentage > 0.50 {
		return 3
	} else if percentage > 0.25 {
		return 2
	} else if percentage > 0 {
		return 1
	}
	return 0
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