package utils

import "math"

type Vector2D struct {
	X, Y float64
}

func (v Vector2D) Distance(other Vector2D) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v Vector2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}