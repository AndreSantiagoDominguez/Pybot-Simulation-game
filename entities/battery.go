package entities

type Battery struct {
	Current    float64 // Carga actual en segundos
	Max        float64 // Capacidad máxima en segundos
	DrainRate  float64 // Segundos consumidos por segundo de movimiento
	ChargeRate float64 // Segundos cargados por segundo
	IsCharging bool
}

func NewBattery() *Battery {
	return &Battery{
		Current:    120.0, // 2 minutos iniciales
		Max:        120.0,
		DrainRate:  1.0,   // 1 seg/seg de movimiento
		ChargeRate: 10.0,  // 10 seg/seg de carga
		IsCharging: false,
	}
}

func (b *Battery) Drain(deltaTime float64) {
	if b.IsCharging {
		return
	}
	
	b.Current -= b.DrainRate * deltaTime
	if b.Current < 0 {
		b.Current = 0
	}
}

func (b *Battery) Charge(deltaTime float64) {
	b.IsCharging = true
	b.Current += b.ChargeRate * deltaTime
	if b.Current > b.Max {
		b.Current = b.Max
	}
}

func (b *Battery) StopCharging() {
	b.IsCharging = false
}

func (b *Battery) Recharge() {
	b.Current = b.Max
}

func (b *Battery) IsEmpty() bool {
	return b.Current <= 0
}

func (b *Battery) GetPercentage() float64 {
	return b.Current / b.Max
}

func (b *Battery) GetLevel() int {
	// Retorna nivel de 0 a 3 para los 4 frames
	// 0 = Lleno, 3 = Vacío
	percentage := b.GetPercentage()
	
	if percentage > 0.75 {
		return 0 // Frame 0: Batería LLENA
	} else if percentage > 0.50 {
		return 1 // Frame 1: Batería ALTA
	} else if percentage > 0.25 {
		return 2 // Frame 2: Batería MEDIA
	} else {
		return 3 // Frame 3: Batería BAJA
	}
}
