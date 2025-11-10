package sensors

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"pybot-simulator/api/rabbitmq"
	"pybot-simulator/api/services"
	"time"
)

//==================================================================
// 2. HX711 READER (Sensor de Peso)
//==================================================================

// HX711Device (Interfaz para simular el hardware)
type HX711Device interface {
	GetRawData(times int) ([]float64, error)
	PowerDown() error
	PowerUp() error
}

// MockHX711 (Implementación simulada de la interfaz)
type MockHX711 struct{ offset float64 }

func NewMockHX711(dataPin, clockPin int, offset float64) *MockHX711 {
	log.Printf("[MockHX711] Inicializado (pines %d, %d)\n", dataPin, clockPin)
	return &MockHX711{offset: offset}
}
func (m *MockHX711) GetRawData(times int) ([]float64, error) {
	raws := make([]float64, times)
	for i := range raws {
		// Simula un peso aleatorio entre -50g y +150g
		simulatedWeightGrams := (rand.Float64() * 200) - 50
		// Simula el valor 'raw' basado en el offset
		raws[i] = m.offset + (simulatedWeightGrams * 0.0388) + (rand.Float64()*10 - 5) // Añade ruido
	}
	return raws, nil
}
func (m *MockHX711) PowerDown() error { return nil }
func (m *MockHX711) PowerUp() error   { return nil }

// HX711Reader (La "clase" que lee el sensor)
type HX711Reader struct {
	prototypeID     string
	mqtt            *rabbitmq.RabbitMQPublisher
	hx              HX711Device
	serviceRegister *services.RegisterPeriods
	handler         *WasteHandler
	offset          float64
	scale           float64
}

// NewHX711Reader es el constructor
func NewHX711Reader(serviceRegister *services.RegisterPeriods, h *WasteHandler) (*HX711Reader, error) {
	mqtt, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		return nil, fmt.Errorf("HX711: falló al inicializar RabbitMQ: %w", err)
	}

	offset := 14664.59
	// Usamos el Mock, pasando los pines y el offset simulado
	hx := NewMockHX711(20, 21, offset)

	return &HX711Reader{
		prototypeID:     os.Getenv("ID_PROTOTYPE"),
		mqtt:            mqtt,
		hx:              hx,
		serviceRegister: serviceRegister,
		handler:         h,
		offset:          offset,
		scale:           0.0388,
	}, nil
}

// runWeightCycle (Lógica de un ciclo para manejar errores)
func (r *HX711Reader) runWeightCycle() error {
	raws, err := r.hx.GetRawData(20)
	if err != nil { return fmt.Errorf("error al leer HX711: %w", err) }
	if len(raws) == 0 { return fmt.Errorf("no se recibieron datos crudos") }

	var sum float64
	for _, val := range raws { sum += val }
	rawAvg := sum / float64(len(raws))

	diff := rawAvg - r.offset
	weight := diff / r.scale
	data := map[string]interface{}{
		"prototype_id": r.prototypeID,
		"weight_g":     weight,
	}

	log.Printf("[HX711] %+v\n", data)
	
	// Publica en MQTT
	if err := r.mqtt.Send(data, "hx"); err != nil {
		log.Printf("[HX711] Error al enviar por MQTT: %v\n", err)
	}

	if weight >= 0 {
		// Inserta en API
		if err := r.serviceRegister.RegisterWeigh(weight); err != nil {
			log.Printf("[HX711] Error al registrar peso en API: %v\n", err)
		}
		// Pasa al handler
		r.handler.ProcessWeight(weight)
	}

	// Ciclo de energía
	r.hx.PowerDown()
	time.Sleep(5 * time.Second)
	r.hx.PowerUp()

	return nil
}

// Start (La función que se manda a llamar)
func (r *HX711Reader) Start() {
	log.Println("[HX711] Iniciando loop...")
	// while True:
	for {
		// try:
		if err := r.runWeightCycle(); err != nil {
			// except Exception as e:
			log.Printf("[HX711] Error en el ciclo: %v\n", err)
			time.Sleep(1 * time.Second)
		}
	}
}

// Close (Para limpiar conexiones)
func (r *HX711Reader) Close() {
	r.mqtt.Close()
}