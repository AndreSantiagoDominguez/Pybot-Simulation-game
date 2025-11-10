package game

import (
	"fmt"
	"image"
	"image/color"

	"pybot-simulator/config"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 50, 255})
	g.DrawPlayArea(screen)
	g.DrawCans(screen)
	g.DrawRobot(screen)
	g.DrawBattery(screen)
	g.DrawButtons(screen)
	g.DrawInfo(screen)
}

func (g *Game) DrawPlayArea(screen *ebiten.Image) {
	margin := float64(config.GridMargin)
	areaColor := color.RGBA{80, 80, 100, 255}
	
	x := margin
	y := margin
	w := float64(g.width) - 2*margin
	h := float64(g.height) - 2*margin
	
	ebitenutil.DrawRect(screen, x, y, w, 2, areaColor)
	ebitenutil.DrawRect(screen, x, y+h-2, w, 2, areaColor)
	ebitenutil.DrawRect(screen, x, y, 2, h, areaColor)
	ebitenutil.DrawRect(screen, x+w-2, y, 2, h, areaColor)
}

func (g *Game) DrawRobot(screen *ebiten.Image) {
	pos := g.robot.Position
	vel := g.robot.Velocity
	
	// Seleccionar el sprite correcto según el movimiento
	var spriteName string
	
	if vel.X == 0 && vel.Y == 0 {
		spriteName = "idle"
	} else {
		// Priorizar dirección horizontal sobre vertical
		if vel.X < -0.1 {
			spriteName = "left"
		} else if vel.X > 0.1 {
			spriteName = "right"
		} else if vel.Y < 0 {
			spriteName = "up"
		} else {
			spriteName = "idle"
		}
	}
	
	currentSprite := g.robot.Sprites[spriteName]
	
	// Si el sprite no existe, usar idle como fallback
	if currentSprite == nil {
		currentSprite = g.robot.Sprites["idle"]
	}
	
	// Si aún no hay sprite, no dibujar nada
	if currentSprite == nil {
		return
	}
	
	// Calcular el frame actual de la animación (4 frames por sprite)
	frameWidth := 75.0 // 300px / 4 frames = 75px por frame
	frameHeight := 75.0
	
	// Animar solo cuando se está moviendo
	frameIndex := 0
	if vel.X != 0 || vel.Y != 0 {
		// Ciclar entre frames 0, 1, 2, 3 basado en el tiempo
		frameIndex = (g.animationCounter / 8) % 4 // Cambia de frame cada 8 ticks
	}
	
	// Crear subimagen del frame actual
	sx := float64(frameIndex) * frameWidth
	frameRect := image.Rect(int(sx), 0, int(sx+frameWidth), int(frameHeight))
	frameImg := currentSprite.SubImage(frameRect).(*ebiten.Image)
	
	// Dibujar el frame centrado en la posición del robot
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-frameWidth/2, -frameHeight/2)
	op.GeoM.Translate(pos.X, pos.Y)
	
	screen.DrawImage(frameImg, op)
}

func (g *Game) DrawBattery(screen *ebiten.Image) {
	// Posición de la batería (esquina superior derecha)
	batteryX := float64(g.width) - 110.0
	batteryY := 10.0
	
	batteryLevel := g.robot.GetBatteryLevel()
	
	if g.robot.BatterySprite == nil {
		// Dibujar batería simple si no hay sprite
		colors := []color.RGBA{
			{50, 255, 50, 255},   // Verde lleno (>75%)
			{150, 255, 50, 255},  // Verde medio (50-75%)
			{255, 200, 50, 255},  // Amarillo (25-50%)
			{255, 50, 50, 255},   // Rojo (<25%)
		}
		
		barColor := colors[batteryLevel]
		width := 100.0
		height := 30.0
		
		// Fondo
		ebitenutil.DrawRect(screen, batteryX, batteryY, width, height, color.RGBA{50, 50, 50, 255})
		
		// Barra de batería
		percentage := g.robot.Battery.Current / g.robot.Battery.Max
		ebitenutil.DrawRect(screen, batteryX+2, batteryY+2, (width-4)*percentage, height-4, barColor)
		
		// Borde
		ebitenutil.DrawRect(screen, batteryX, batteryY, width, 2, color.RGBA{200, 200, 200, 255})
		ebitenutil.DrawRect(screen, batteryX, batteryY+height-2, width, 2, color.RGBA{200, 200, 200, 255})
		ebitenutil.DrawRect(screen, batteryX, batteryY, 2, height, color.RGBA{200, 200, 200, 255})
		ebitenutil.DrawRect(screen, batteryX+width-2, batteryY, 2, height, color.RGBA{200, 200, 200, 255})
	} else {
		// Dibujar sprite de batería (300x60px dividido en 4 frames de 75x60px)
		// Frame 0 (izq) = Lleno | Frame 3 (der) = Vacío
		frameWidth := 75.0  // 300px ÷ 4 frames = 75px
		frameHeight := 60.0
		
		// Obtener frame según nivel (0=lleno, 3=vacío)
		frameIndex := batteryLevel
		
		// Calcular posición X del frame en el sprite sheet
		sx := float64(frameIndex) * frameWidth
		frameRect := image.Rect(int(sx), 0, int(sx+frameWidth), int(frameHeight))
		frameImg := g.robot.BatterySprite.SubImage(frameRect).(*ebiten.Image)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(batteryX, batteryY)
		screen.DrawImage(frameImg, op)
	}
}

func (g *Game) DrawCans(screen *ebiten.Image) {
	for _, can := range g.cans {
		if !can.Active {
			continue
		}
		
		pos := can.Position
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-config.CanSize/2, -config.CanSize/2)
		op.GeoM.Translate(pos.X, pos.Y)
		
		if can.Sprite != nil {
			screen.DrawImage(can.Sprite, op)
		}
	}
}

func (g *Game) DrawButtons(screen *ebiten.Image) {
	g.DrawButton(screen, g.spawnButton)
	g.DrawButton(screen, g.rechargeButton)
}

func (g *Game) DrawButton(screen *ebiten.Image, btn Button) {
	ebitenutil.DrawRect(screen, btn.X, btn.Y, btn.Width, btn.Height,
		color.RGBA{70, 130, 180, 255})
	
	ebitenutil.DrawRect(screen, btn.X, btn.Y, btn.Width, 2, color.RGBA{200, 200, 200, 255})
	ebitenutil.DrawRect(screen, btn.X, btn.Y+btn.Height-2, btn.Width, 2, color.RGBA{200, 200, 200, 255})
	ebitenutil.DrawRect(screen, btn.X, btn.Y, 2, btn.Height, color.RGBA{200, 200, 200, 255})
	ebitenutil.DrawRect(screen, btn.X+btn.Width-2, btn.Y, 2, btn.Height, color.RGBA{200, 200, 200, 255})
	
	ebitenutil.DebugPrintAt(screen, btn.Text, int(btn.X+15), int(btn.Y+13))
}

func (g *Game) DrawInfo(screen *ebiten.Image) {
	activeCans := g.GetActiveCansCount()
	
	info := fmt.Sprintf("Recolectadas: %d | Activas: %d",
		g.robot.CansCollected, activeCans)
	ebitenutil.DebugPrintAt(screen, info, 10, 10)
	
	status := "Estado: Buscando latas"
	if g.robot.Battery.IsCharging {
		status = "Estado: CARGANDO"
	} else if g.robot.Battery.IsEmpty() {
		status = "Estado: SIN BATERÍA"
	} else if g.robot.Target != nil {
		status = "Estado: Recolectando"
	} else if activeCans == 0 {
		status = "Estado: Esperando latas"
	}
	ebitenutil.DebugPrintAt(screen, status, 10, 30)
	
	controls := "Controles: S = Spawn latas | R (mantener) = Recargar"
	ebitenutil.DebugPrintAt(screen, controls, 10, 50)
}