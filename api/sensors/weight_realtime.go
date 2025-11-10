package sensors

import (
	"log"
	"pybot-simulator/api/rabbitmq"
	"pybot-simulator/api/services"
)

// WeightSensor handles sending weight and waste data.
type WeightSensor struct {
	publisher       *rabbitmq.RabbitMQPublisher
	registerPeriods *services.RegisterPeriods
}

// NewWeightSensor creates a new weight sensor.
func NewWeightSensor(rp *services.RegisterPeriods) (*WeightSensor, error) {
	publisher, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		log.Printf("Warning: Weight sensor could not connect to RabbitMQ. Weight data will not be sent to broker.")
	}

	return &WeightSensor{
		publisher:       publisher,
		registerPeriods: rp,
	}, nil
}

// RegisterWeight sends the total weight to the API and RabbitMQ.
func (s *WeightSensor) RegisterWeight(totalWeight float64) {
	// Send total weight to /weight endpoint
	go func() {
		if err := s.registerPeriods.RegisterWeigh(totalWeight); err != nil {
			log.Printf("Warning: Failed to send weight data to API: %v", err)
		} else {
			log.Printf("Successfully sent total weight %fg to API.", totalWeight)
		}
	}()

	// Send total weight to RabbitMQ
	go func() {
		payload := map[string]interface{}{
			"prototype_id":  "70755f712d864350abf6df03",
			"weight_g": totalWeight,
		}
		if sent, err := s.publisher.Send(payload, "hx"); err != nil {
			log.Printf("Warning: Failed to send weight data to RabbitMQ: %v", err)
		} else if sent {
			log.Printf("Successfully sent total weight %fg to RabbitMQ.", totalWeight)
		}
	}()
}

// UpdateWasteCount sends a PATCH request to update the count for a given waste type.
func (s *WeightSensor) UpdateWasteCount(wasteID int64) {
	go func() {
		var collectionID int64
		if wasteID == 1 { // PET
			collectionID = s.registerPeriods.GetIdWasteCollectionPET()
		} else if wasteID == 2 { // CAN
			collectionID = s.registerPeriods.GetIdWasteCollectionCANS()
		} else {
			log.Printf("Warning: Unknown wasteID %d for waste count update.", wasteID)
			return
		}

		if collectionID == 0 {
			log.Printf("Warning: No collection ID found for wasteID %d. Cannot update count.", wasteID)
			return
		}

		if err := s.registerPeriods.UpdateWasteCollection(collectionID); err != nil {
			log.Printf("Warning: Failed to update waste collection for wasteID %d (collectionID: %d): %v", wasteID, collectionID, err)
		} else {
			log.Printf("Successfully sent PATCH to update waste collection for wasteID %d.", wasteID)
		}
	}()
}