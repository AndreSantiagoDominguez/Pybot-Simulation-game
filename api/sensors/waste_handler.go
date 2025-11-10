package sensors

import (
	"fmt"
	"log"
	"pybot-simulator/api/services"
	"sync" // Para el 'threading.Lock' (Mutex)
	"time"
)

type Detection struct {
	Cls  int     `json:"cls"`
	Conf float64 `json:"conf"`
}

// WasteHandler es el struct que coordina los sensores
type WasteHandler struct {
	service *services.RegisterPeriods
	idPET   int64
	idCANS  int64

	weightThreshold float64
	expireDelta     time.Duration

	// mu es el 'threading.Lock' de Python
	mu         sync.Mutex
	isCan      bool
	isPet      bool
	detectTime time.Time // Un time.Time 'cero' (vacío) es el 'None' de Python
	lastWeight float64
}

// NewWasteHandler es el constructor (equivalente a __init__)
func NewWasteHandler(service *services.RegisterPeriods, weightThreshold float64, expireSeconds int) (*WasteHandler, error) {
	log.Println("[Handler] Inicializando...")
	
	// service.createWasteCollection(1) # PET
	err := service.CreateWasteCollection(1)
	if err != nil {
		return nil, fmt.Errorf("error al crear colección PET: %w", err)
	}
	idPET := service.GetIdWasteCollectionPET()

	// service.createWasteCollection(2) # CANS
	err = service.CreateWasteCollection(2)
	if err != nil {
		return nil, fmt.Errorf("error al crear colección CANS: %w", err)
	}
	idCANS := service.GetIdWasteCollectionCANS()

	log.Printf("[Handler] IDs de recolección: PET=%d, CANS=%d\n", idPET, idCANS)

	return &WasteHandler{
		service:         service,
		idPET:           idPET,
		idCANS:          idCANS,
		weightThreshold: weightThreshold,
		expireDelta:     time.Duration(expireSeconds) * time.Second,
	}, nil
}

// ProcessDetections (Llamado por CameraReader)
func (h *WasteHandler) ProcessDetections(detections []Detection) {
	now := time.Now().UTC()
	
	// with self.lock:
	h.mu.Lock()
	defer h.mu.Unlock()

	// Reset flags si han caducado
	// if self.detect_time and now - self.detect_time > self.expire_delta:
	if !h.detectTime.IsZero() && now.Sub(h.detectTime) > h.expireDelta {
		log.Println("[Handler] Flags de detección caducados (previo a nueva detección).")
		h.isCan = false
		h.isPet = false
	}

	// Lógica de detección (adaptada de tu script)
	if len(detections) > 2 {
		h.isCan = true
		h.isPet = true
	} else if len(detections) > 0 {
		if detections[0].Cls == 0 { // 0 = PET
			h.isPet = true
		} else { // cualquier otro = can
			h.isCan = true
		}
	}

	h.detectTime = now
	log.Printf("[Handler] Detección -> is_pet=%t, is_can=%t\n", h.isPet, h.isCan)
}

// ProcessWeight (Llamado por HX711Reader)
func (h *WasteHandler) ProcessWeight(weight float64) {
	now := time.Now().UTC()
	
	// with self.lock:
	h.mu.Lock()
	defer h.mu.Unlock()

	delta := weight - h.lastWeight

	// if delta >= self.weight_threshold and self.detect_time:
	// (!h.detectTime.IsZero() es como 'self.detect_time is not None')
	if delta >= h.weightThreshold && !h.detectTime.IsZero() {
		
		elapsed := now.Sub(h.detectTime)
		if elapsed <= h.expireDelta {
			// Se concreta un residuo
			if h.isPet {
				log.Printf("[Handler] +1 PET (Δweight=%.2fg)\n", delta)
				h.service.UpdateWasteCollection(h.idPET) // Ignoramos error (como en Python)
				h.isPet = false
			}
			if h.isCan {
				log.Printf("[Handler] +1 Can (Δweight=%.2fg)\n", delta)
				h.service.UpdateWasteCollection(h.idCANS) // Ignoramos error
				h.isCan = false
			}
			// Tras concreción, borramos detect_time
			h.detectTime = time.Time{} // Resetea a valor 'cero' (None)
		}
	}

	// Caducado sin concreción
	if !h.detectTime.IsZero() && now.Sub(h.detectTime) > h.expireDelta {
		log.Println("[Handler] Detección caducada sin peso, reseteando flags")
		h.isCan = false
		h.isPet = false
		h.detectTime = time.Time{}
	}

	h.lastWeight = weight
}