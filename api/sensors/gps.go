package sensors

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"pybot-simulator/api/rabbitmq"
	"pybot-simulator/api/services"
	"strconv"
	"time"
)

//==================================================================
// 4. GPS READER (Sensor de Posicionamiento)
//==================================================================

// MockGPSDevice simula la combinación de 'serial.Serial' y 'pynmea2'
// Devuelve los datos *como si pynmea2 ya los hubiera parseado*
type MockGPSDevice struct{}

func NewMockGPSDevice(port string, baud int) (*MockGPSDevice, error) {
	if port != "/dev/serial0" {
		log.Printf("[MockGPS] Puerto %s no encontrado\n", port)
		return nil, fmt.Errorf("mock port not found")
	}
	log.Printf("[MockGPS] Puerto %s abierto a %d baud\n", port, baud)
	return &MockGPSDevice{}, nil
}

// Read simula la lectura de una línea NMEA y su parseo
// Devuelve: (tipo de línea, datos parseados, error)
func (m *MockGPSDevice) Read() (string, map[string]interface{}, error) {
	time.Sleep(100 * time.Millisecond) // Simula espera de I/O
	parsedData := make(map[string]interface{})
	r := rand.Float64()

	if r < 0.45 {
		// Simula $GPRMC (lat, lon, spd, date, time)
		parsedData["lat"] = 22.76 + (rand.Float64() * 0.01) // Coordenadas simuladas
		parsedData["lon"] = -102.58 + (rand.Float64() * 0.01)
		// 'spd_over_grnd' en nudos
		parsedData["spd_over_grnd"] = 20.0 + rand.Float64()*5
		if rand.Float64() > 0.1 { // 90% de las veces tenemos fecha/hora
			parsedData["datestamp"] = time.Now() // time.Time object
			parsedData["timestamp"] = time.Now() // time.Time object
		}
		return "$GPRMC", parsedData, nil
	} else if r < 0.9 {
		// Simula $GPGGA (alt, sats)
		parsedData["altitude"] = 1880.0 + rand.Float64()*2
		if rand.Float64() > 0.1 { // 90% de las veces 'num_sats' es un string válido
			parsedData["num_sats"] = fmt.Sprintf("%d", rand.Intn(4)+5) // 5-8 satélites
		} else {
			parsedData["num_sats"] = "bad_data" // Simula ValueError en Python
		}
		return "$GPGGA", parsedData, nil
	} else {
		// Simula línea NMEA basura o no deseada
		return "$GPVTG", nil, nil
	}
}

// GPSReader (La "clase" que lee el sensor)
type GPSReader struct {
	register    *services.RegisterPeriods
	mqtt        *rabbitmq.RabbitMQPublisher
	prototypeID string
	device      *MockGPSDevice
}

// NewGPSReader es el constructor (no abre el puerto aún)
func NewGPSReader(serviceRegister *services.RegisterPeriods) (*GPSReader, error) {
	mqtt, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		return nil, fmt.Errorf("GPS: falló al inicializar RabbitMQ: %w", err)
	}

	return &GPSReader{
		register:    serviceRegister,
		mqtt:        mqtt,
		prototypeID: os.Getenv("ID_PROTOTYPE"),
		// device se inicializará en Start()
	}, nil
}

// knotsToKmph (Equivalente a tu @staticmethod)
func knotsToKmph(knots float64) float64 {
	return knots * 1.852
}

// Start (La función que se manda a llamar)
func (r *GPSReader) Start() {
	// try: ser = serial.Serial(...)
	ser, err := NewMockGPSDevice("/dev/serial0", 9600)
	if err != nil {
		log.Printf("[GPS] No se pudo abrir puerto: %v\n", err)
		return // Termina la gorutina (equivale a 'return' en Python)
	}
	r.device = ser

	log.Println("[GPS] Iniciando loop...")
	lastData := make(map[string]interface{})

	// while True:
	for {
		// try: (loop principal)
		// Simula: line = ser.readline()... msg = pynmea2.parse(line)
		lineType, msg, err := r.device.Read()
		if err != nil {
			log.Printf("[GPS] Error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// if line.startswith('$GPRMC'):
		if lineType == "$GPRMC" {
			lastData["prototype_id"] = r.prototypeID
			lastData["lat"] = msg["lat"]
			lastData["lon"] = msg["lon"]

			spdKnots, _ := msg["spd_over_grnd"].(float64) // 0.0 si es nil
			// last_data['spd'] = round(self.knots_to_kmph(speed_knots), 2)
			lastData["spd"] = math.Round(knotsToKmph(spdKnots)*100) / 100

			// if msg.datestamp else None
			if msg["datestamp"] != nil {
				lastData["date"] = msg["datestamp"].(time.Time).Format("2006-01-02")
			} else {
				lastData["date"] = nil
			}
			// if msg.timestamp else None
			if msg["timestamp"] != nil {
				lastData["UTC"] = msg["timestamp"].(time.Time).Format("15:04:05")
			} else {
				lastData["UTC"] = nil
			}

		// elif line.startswith('$GPGGA'):
		} else if lineType == "$GPGGA" {
			lastData["alt"] = msg["altitude"]
			
			// try: last_data['sats'] = int(msg.num_sats)
			satsStr, _ := msg["num_sats"].(string)
			sats, err := strconv.Atoi(satsStr)
			if err != nil {
				// except ValueError: last_data['sats'] = None
				lastData["sats"] = nil
			} else {
				lastData["sats"] = sats
			}
		}

		// if 'lat' in last_data and 'lon' in last_data:
		if _, latOK := lastData["lat"]; latOK {
			if _, lonOK := lastData["lon"]; lonOK {
				
				log.Printf("[GPS] %+v\n", lastData)

				// self.mqtt.send(payload=last_data, routing_key="neo")
				r.mqtt.Send(lastData, "neo")

				// self.register.registerGPS(last_data)
				r.register.RegisterGPS(lastData)
			}
		}

		// time.sleep(5)
		time.Sleep(5 * time.Second)
	}
}

// Close (Para limpiar conexiones)
func (r *GPSReader) Close() {
	r.mqtt.Close()
}