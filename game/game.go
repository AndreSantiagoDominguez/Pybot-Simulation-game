package game

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"pybot-simulator/api/sensors"
	"pybot-simulator/api/services"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"pybot-simulator/config"
	"pybot-simulator/entities"
)

type Game struct {
	width  int
	height int

	robot *entities.Robot
	cans  []*entities.Can

	spawnButton    Button
	rechargeButton Button

	canSprite *ebiten.Image

	rng               *rand.Rand
	animationCounter  int
	realTimeCamera    *sensors.RealTimeCamera
	cameraTicks       int
	gpsSensor         *sensors.GPSSensor
	gpsTicks          int
	registerPeriods   *services.RegisterPeriods
	batteryDepleted   bool
}

type Button struct {
	X, Y, Width, Height float64
	Text                string
}

func NewGame(width, height int) *Game {
	g := &Game{
		width:            width,
		height:           height,
		cans:             make([]*entities.Can, 0),
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		animationCounter: 0,
		batteryDepleted:  false,
	}

	// Initialize the real-time camera sensor
	var err error
	g.realTimeCamera, err = sensors.NewRealTimeCamera()
	if err != nil {
		log.Printf("Warning: Failed to initialize real-time camera: %v", err)
	}

	// Initialize the work period service
	g.registerPeriods, err = services.NewRegisterPeriods()
	if err != nil {
		log.Fatalf("Failed to initialize register periods service: %v", err)
	}

	// Initialize the GPS sensor
	g.gpsSensor, err = sensors.NewGPSSensor(g.registerPeriods, width, height)
	if err != nil {
		log.Fatalf("Failed to initialize GPS sensor: %v", err)
	}

	// Create a new work period on start
	log.Println("Creating a new work period...")
	g.createInitialWorkPeriod()

	// Crear robot en el centro (sin sprite todavía)
	g.robot = entities.NewRobot(
		float64(width)/2,
		float64(height)/2,
		nil,
	)
	
	// Establecer límites del robot
	margin := float64(config.GridMargin)
	g.robot.SetBounds(
		margin,
		float64(width)-margin,
		margin,
		float64(height)-margin,
	)
	
	// Cargar sprites
	g.LoadAssets()
	
	// Crear botones
	g.spawnButton = Button{
		X:      20,
		Y:      float64(height - 110),
		Width:  180,
		Height: 40,
		Text:   "Spawn Latas (S)",
	}
	
g.rechargeButton = Button{
		X:      20,
		Y:      float64(height - 60),
		Width:  180,
		Height: 40,
		Text:   "Recargar (R)",
	}
	
	// Spawn inicial
	g.SpawnCans(5)
	
	return g
}

func (g *Game) createInitialWorkPeriod() {
	if err := g.registerPeriods.CreateNewPeriod(); err != nil {
		log.Printf("Warning: Failed to create a new work period: %v", err)
		return
	}
	// As per user's instruction, create a void reading after creating a new period.
	if err := g.registerPeriods.CreateVoidReading(); err != nil {
		log.Printf("Warning: Failed to create void reading: %v", err)
		return
	}
	log.Println("Successfully created new work period and void reading.")
}

func (g *Game) completeAndStartNewPeriod() {
	if err := g.registerPeriods.CompleteLastPeriod(); err != nil {
		log.Printf("Warning: Failed to complete work period: %v", err)
	}
}

func (g *Game) LoadAssets() {
	// Cargar sprites del robot
	g.loadRobotSprites()
	
	// Cargar sprite de batería
	g.loadBatterySprite()
	
	// Cargar sprite de la lata
	var err error
	g.canSprite, _, err = ebitenutil.NewImageFromFile("assets/trash-sprites/trash_types.png")
	if err != nil {
		log.Printf("No se pudo cargar trash_types.png: %v", err)
		// Sprite temporal para la lata
		g.canSprite = ebiten.NewImage(config.CanSize, config.CanSize)
		g.canSprite.Fill(color.RGBA{255, 200, 50, 255})
	}
}

func (g *Game) loadRobotSprites() {
	sprites := make(map[string]*ebiten.Image)
	
spriteFiles := map[string]string{
		"idle":  "assets/pybot-moves/pybot_idle.png",
		"up":    "assets/pybot-moves/pybot_walk_up.png",
		"left":  "assets/pybot-moves/pybot_walk_left.png",
		"right": "assets/pybot-moves/pybot_walk_right.png",
	}
	
	allLoaded := true
	for name, path := range spriteFiles {
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			log.Printf("Error cargando sprite %s: %v", name, err)
			allLoaded = false
		} else {
			sprites[name] = img
		}
	}
	
	if allLoaded && len(sprites) > 0 {
		g.robot.Sprites = sprites
	} else {
		log.Println("Usando sprites temporales para el robot")
		tempSprite := ebiten.NewImage(config.RobotSize, config.RobotSize)
		tempSprite.Fill(color.RGBA{100, 150, 255, 255})
		sprites["idle"] = tempSprite
		sprites["up"] = tempSprite
		sprites["left"] = tempSprite
		sprites["right"] = tempSprite
		g.robot.Sprites = sprites
	}
}

func (g *Game) loadBatterySprite() {
	img, _, err := ebitenutil.NewImageFromFile("assets/energy-buttons/battery_indicator.png")
	if err != nil {
		log.Printf("No se pudo cargar battery.png: %v", err)
		// Sprite temporal
		g.robot.BatterySprite = ebiten.NewImage(100, 25)
		g.robot.BatterySprite.Fill(color.RGBA{0, 255, 0, 255})
	} else {
		g.robot.BatterySprite = img
	}
}

func (g *Game) SpawnCans(count int) {
	margin := float64(config.GridMargin)
	
	for i := 0; i < count; i++ {
		x := margin + g.rng.Float64()*(float64(g.width)-2*margin)
		y := margin + g.rng.Float64()*(float64(g.height)-2*margin)
		
		can := entities.NewCan(x, y, g.canSprite)
		g.cans = append(g.cans, can)
	}
}

func (g *Game) FindNearestCan() *entities.Can {
	var nearest *entities.Can
	minDistance := math.MaxFloat64
	
	for _, can := range g.cans {
		if !can.Active {
			continue
		}
		
		distance := g.robot.Position.Distance(can.Position)
		if distance < minDistance {
			minDistance = distance
			nearest = can
		}
	}
	
	return nearest
}

func (g *Game) CheckCollisions() {
	robotPos := g.robot.Position
	collectRadius := float64(config.RobotSize/2 + config.CanSize/2)
	
	for _, can := range g.cans {
		if !can.Active {
			continue
		}
		
		distance := robotPos.Distance(can.Position)
		
		if distance < collectRadius {
			can.Deactivate()
			g.robot.CollectCan()
			fmt.Printf("¡Lata recogida! Total: %d\n", g.robot.CansCollected)
		}
	}
}

func (g *Game) GetActiveCansCount() int {
	count := 0
	for _, can := range g.cans {
		if can.Active {
			count++
		}
	}
	return count
}

func (g *Game) Update() error {
	g.animationCounter++
	g.HandleInput()

	// Si el robot no tiene objetivo y tiene batería, buscar la lata más cercana
	if g.robot.Target == nil && g.robot.Battery > 0 {
		nearest := g.FindNearestCan()
		if nearest != nil {
			g.robot.SetTarget(nearest.Position)
		}
	}

	// Handle real-time camera publishing
	if g.robot.Battery > 0 {
		g.cameraTicks++
		// Publish an image every 180 ticks (e.g., every 3 seconds at 60 TPS)
		if g.cameraTicks >= 180 {
			g.cameraTicks = 0
			go g.realTimeCamera.PublishRandomImage()
		}

		// Handle GPS publishing when moving
		if g.robot.Velocity.X != 0 || g.robot.Velocity.Y != 0 {
			g.gpsTicks++
			// Publish GPS data every 60 ticks (e.g., every 1 second at 60 TPS)
			if g.gpsTicks >= 60 {
				g.gpsTicks = 0
				go g.sendGPSData()
			}
		}

	} else {
		// Set flag when battery is depleted
		if !g.batteryDepleted {
			log.Println("Robot battery depleted.")
			g.batteryDepleted = true
		}
	}

	g.robot.Update()
	g.CheckCollisions()
	return nil
}

func (g *Game) sendGPSData() {
	gpsData := g.gpsSensor.GenerateGPSData(g.robot.Position, g.robot.Velocity)
	g.gpsSensor.SendGPSData(gpsData)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
}

func (g *Game) IsPointInButton(btn Button, x, y float64) bool {
	return x >= btn.X && x <= btn.X+btn.Width &&
		y >= btn.Y && y <= btn.Y+btn.Height
}
