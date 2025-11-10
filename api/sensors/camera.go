package sensors

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"pybot-simulator/api/rabbitmq"
	"sync"
	"time"
)

//==================================================================
// 3. CAMERA READER (Sensor de Cámara)
//==================================================================

// Frame (Simulacro de un frame de cv2)
type Frame []byte // Usamos un slice de bytes como simulación

// AIModel (Interfaz para simular YOLO)
type AIModel interface {
	Predict(f Frame) ([]Detection, Frame, error) // Devuelve detecciones y frame anotado
}

// MockAIModel (Implementación simulada de YOLO)
type MockAIModel struct{}

func NewMockAIModel(modelPath string) *MockAIModel {
	log.Printf("[MockYOLO] Cargando modelo desde %s\n", modelPath)
	return &MockAIModel{}
}
func (m *MockAIModel) Predict(f Frame) ([]Detection, Frame, error) {
	// Simula una detección aleatoria (PET o CAN)
	detections := []Detection{}
	if rand.Float64() > 0.3 { // 70% de chance de detectar algo
		detections = append(detections, Detection{
			Cls:  rand.Intn(2), // Cls 0 (PET) o 1 (CAN)
			Conf: 0.8 + rand.Float64()*0.2, // Confianza alta
		})
	}
	// Devuelve las detecciones y el frame "anotado" (el mismo frame)
	return detections, f, nil
}

// CameraDevice (Interfaz para simular cv2.VideoCapture)
type CameraDevice interface {
	Read() (Frame, error)
	Close()
}

// MockCameraDevice (Implementación simulada de la cámara)
type MockCameraDevice struct{}

func NewMockCameraDevice(index int) (*MockCameraDevice, error) {
	log.Printf("[MockCamera] Abriendo dispositivo %d\n", index)
	return &MockCameraDevice{}, nil
}
func (m *MockCameraDevice) Read() (Frame, error) {
	// Simula la captura de un frame (solo un timestamp)
	time.Sleep(100 * time.Millisecond) // Simula 10 FPS
	return []byte(time.Now().Format(time.RFC3339Nano)), nil
}
func (m *MockCameraDevice) Close() { log.Println("[MockCamera] Dispositivo cerrado.") }

// CameraReader (La "clase" que lee la cámara)
type CameraReader struct {
	prototypeID string
	mqtt        *rabbitmq.RabbitMQPublisher
	handler     *WasteHandler
	model       AIModel
	cam         CameraDevice

	frameLock   sync.Mutex
	latestFrame Frame
}

// NewCameraReader es el constructor
func NewCameraReader(h *WasteHandler) (*CameraReader, error) {
	mqtt, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		return nil, fmt.Errorf("Camera: falló al inicializar RabbitMQ: %w", err)
	}

	// Usamos los Mocks
	model := NewMockAIModel("models/best2.pt")
	cam, err := NewMockCameraDevice(0)
	if err != nil {
		return nil, fmt.Errorf("Camera: falló al abrir dispositivo: %w", err)
	}

	return &CameraReader{
		prototypeID: os.Getenv("ID_PROTOTYPE"),
		mqtt:        mqtt,
		handler:     h,
		model:       model,
		cam:         cam,
	}, nil
}

// captureLoop (Equivalente a capture_thread)
func (r *CameraReader) captureLoop() {
	log.Println("[Camera] Iniciando hilo de captura...")
	for {
		f, err := r.cam.Read()
		if err != nil {
			log.Printf("[Camera] Error en la captura: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// with self.frame_lock:
		r.frameLock.Lock()
		r.latestFrame = f
		r.frameLock.Unlock()
	}
}

// encodeImage (Simulación de cv2.imencode y base64)
func (r *CameraReader) encodeImage(frame Frame) (string, error) {
	// En un caso real, aquí usarías una librería de CV
	// Aquí, solo codificamos el texto de simulación
	return base64.StdEncoding.EncodeToString(frame), nil
}

// Start (La función que se manda a llamar)
func (r *CameraReader) Start() {
	// t = threading.Thread(target=self.capture_thread, daemon=True)
	// t.start()
	go r.captureLoop() // Inicia la captura en una gorutina

	log.Println("[Camera] Iniciando loop de inferencia...")
	
	// while True:
	for {
		// try:
		err := r.runInferenceCycle()
		if err != nil {
			// except Exception as e:
			log.Printf("[Camera] Error en el ciclo: %v\n", err)
			time.Sleep(1 * time.Second)
		}
	}
}

// runInferenceCycle (Lógica de un ciclo para manejar errores)
func (r *CameraReader) runInferenceCycle() error {
	var f Frame

	// with self.frame_lock:
	r.frameLock.Lock()
	if r.latestFrame != nil {
		f = make(Frame, len(r.latestFrame))
		copy(f, r.latestFrame) // f = self.latest_frame.copy()
	}
	r.frameLock.Unlock()

	// if f is None: continue
	if f == nil {
		time.Sleep(10 * time.Millisecond) // Espera a que haya un frame
		return nil // No es un error, solo espera
	}

	// Simulación de resize (omitida, el mock la maneja)
	// results = self.model(small)
	detections, annFrame, err := r.model.Predict(f)
	if err != nil {
		return fmt.Errorf("error de inferencia: %w", err)
	}

	log.Printf("[Camera] Detecciones: %+v\n", detections)
	// self.handler.process_detections(detections)
	r.handler.ProcessDetections(detections)

	// Simulación de resize y encode
	// ann = results[0].plot()
	// img = cv2.resize(ann, (320,240))
	encodedImg, err := r.encodeImage(annFrame) // annFrame es el 'img'
	if err != nil {
		return fmt.Errorf("error codificando imagen: %w", err)
	}

	if encodedImg != "" {
		payload := map[string]interface{}{
			"prototype_id": r.prototypeID,
			"detections":   detections,
			"image":        encodedImg,
		}
		// self.mqtt.send(payload, routing_key='cam')
		if err := r.mqtt.Send(payload, "cam"); err != nil {
			log.Printf("[Camera] Error al enviar por MQTT: %v\n", err)
		}
	}
	
	// El 'sleep' es manejado por la cámara (cv2.waitKey o el bloqueo de Read())
	return nil
}

// Close (Para limpiar conexiones)
func (r *CameraReader) Close() {
	r.mqtt.Close()
	r.cam.Close()
}