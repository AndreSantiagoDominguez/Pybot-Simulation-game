package sensors

import (
	"log"
	"math/rand"
	"pybot-simulator/api/rabbitmq"
	"pybot-simulator/api/services"
	"pybot-simulator/utils"
	"time"
)

const (
	// Origin coordinates (Mexico City Zocalo)
	originLat = 19.4326
	originLon = -99.1332

	// Scale factor: 1 pixel = 0.00001 degrees
	coordScale = 0.00001

	// Fixed altitude with some noise
	baseAltitude = 2240.0 // meters
)

// GPSSensor handles the generation and sending of GPS data.
type GPSSensor struct {
	publisher       *rabbitmq.RabbitMQPublisher
	registerPeriods *services.RegisterPeriods
	originX         float64
	originY         float64
}

// NewGPSSensor creates a new GPS sensor.
func NewGPSSensor(rp *services.RegisterPeriods, screenWidth, screenHeight int) (*GPSSensor, error) {
	publisher, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		// The publisher already logs a warning, but we can add context.
		log.Printf("Warning: GPS sensor could not connect to RabbitMQ. GPS data will not be sent to broker.")
	}

	return &GPSSensor{
		publisher:       publisher,
		registerPeriods: rp,
		originX:         float64(screenWidth / 2),
		originY:         float64(screenHeight / 2),
	}, nil
}

// GenerateGPSData creates a GPS data map from the robot's state.
func (s *GPSSensor) GenerateGPSData(position, velocity utils.Vector2D) map[string]interface{} {
	// Map simulation coordinates to GPS coordinates
	lat := originLat + (position.Y-s.originY)*coordScale
	lon := originLon + (position.X-s.originX)*coordScale

	// Calculate speed from velocity magnitude (and convert to km/h for realism)
	// Assuming 1 pixel/tick = 0.1 m/s
	speedMs := velocity.Magnitude() * 0.1
	speedKmh := speedMs * 3.6

	// Add some noise to altitude
	alt := baseAltitude + (rand.Float64()*2 - 1) // +/- 1 meter

	// Get current time for the payload
	now := time.Now().UTC()

	data := map[string]interface{}{
		"prototype_id":  "a99fd25c7e4a4e2cb5b7a1d1",
		"lat": lat,
		"lon": lon,
		"alt": alt,
		"spd": speedKmh,
		"date": now.Format("2006-01-02"),
		"UTC":  now.Format("15:04:05.999"),
	}
	return data
}

// SendGPSData sends the generated data to the API and RabbitMQ.
func (s *GPSSensor) SendGPSData(data map[string]interface{}) {
	// Send to API
	if err := s.registerPeriods.RegisterGPS(data); err != nil {
		log.Printf("Warning: Failed to send GPS data to API: %v", err)
	} else {
		log.Println("Successfully sent GPS data to API.")
	}

	// Send to RabbitMQ
	if sent, err := s.publisher.Send(data, "neo"); err != nil {
		log.Printf("Warning: Failed to send GPS data to RabbitMQ: %v", err)
	} else if sent {
		log.Println("Successfully sent GPS data to RabbitMQ.")
	}
}
