package entities

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"pybot-simulator/utils"
)

const (
	PET = iota
	CAN
)

type Can struct {
	Position utils.Vector2D
	Active   bool
	Sprite   *ebiten.Image
	Type     int
	Weight   float64
	WasteID  int64
}

func NewCan(x, y float64, sprite *ebiten.Image) *Can {
	canType := PET
	weight := 10.0
	wasteID := int64(1)

	if rand.Float64() > 0.5 {
		canType = CAN
		weight = 20.0
		wasteID = int64(2)
	}

	return &Can{
		Position: utils.Vector2D{X: x, Y: y},
		Active:   true,
		Sprite:   sprite,
		Type:     canType,
		Weight:   weight,
		WasteID:  wasteID,
	}
}

func (c *Can) GetPosition() utils.Vector2D {
	return c.Position
}

func (c *Can) IsActive() bool {
	return c.Active
}

func (c *Can) Deactivate() {
	c.Active = false
}