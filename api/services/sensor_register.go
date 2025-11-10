package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SensorRegisterService maneja la comunicación con la API de registro de sensores.
type SensorRegisterService struct {
	baseURL string
	client  *http.Client
}

// NewSensorRegisterService crea un nuevo cliente de servicio.
// Es el equivalente a tu __init__.
func NewSensorRegisterService() *SensorRegisterService {
	return &SensorRegisterService{
		baseURL: "https://pybot.aleosh.online/sensors/sensors",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// helperJSONRequest es una función privada para manejar POST con JSON.
// Esta función también intenta decodificar el JSON incluso si hay un error HTTP (status 4xx o 5xx),
// para imitar la lógica de `return resp.json()` en tu bloque `except` de Python.
func (s *SensorRegisterService) helperJSONRequest(method, url string, payload interface{}) (interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("[FetchAPI] Error codificando JSON: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error creando request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		// Error de red o conexión
		return nil, fmt.Errorf("[FetchAPI] Error creando recurso (%s): %w", method, err)
	}
	defer resp.Body.Close()

	// Decodificar el JSON de la respuesta, *incluso si es un error*.
	var result interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)

	// Ahora, verificar el status code
	if resp.StatusCode >= 400 {
		errMsg := fmt.Errorf("[FetchAPI] Error creando recurso, status %d", resp.StatusCode)
		if decodeErr == nil {
			// Pudimos decodificar el cuerpo del error, lo devolvemos junto al error
			return result, errMsg
		}
		// No pudimos decodificar, solo devolvemos el error de status
		return nil, errMsg
	}

	// Si el status estuvo OK (2xx), pero falló la decodificación
	if decodeErr != nil {
		return nil, fmt.Errorf("[FetchAPI] Error decodificando JSON de respuesta: %w", decodeErr)
	}

	// Todo OK
	return result, nil
}

// RegisterWeightData registra datos de peso (POST a /weight).
func (s *SensorRegisterService) RegisterWeightData(payload interface{}) (interface{}, error) {
	fullURL := s.baseURL + "/weight"
	return s.helperJSONRequest("POST", fullURL, payload)
}

// RegisterGPSData registra datos de GPS (POST a /gps).
func (s *SensorRegisterService) RegisterGPSData(payload interface{}) (interface{}, error) {
	fullURL := s.baseURL + "/gps"
	return s.helperJSONRequest("POST", fullURL, payload)
}

// RegisterWasteCollection registra una recolección de basura (POST a /waste).
func (s *SensorRegisterService) RegisterWasteCollection(payload interface{}) (interface{}, error) {
	fullURL := s.baseURL + "/waste"
	return s.helperJSONRequest("POST", fullURL, payload)
}

// UpdateWasteCollection actualiza una recolección (PATCH a /?Id=...).
// Nota: 'wID' debe ser un string para coincidir con `str(w_id)` de Python.
func (s *SensorRegisterService) UpdateWasteCollection(wID string) (interface{}, error) {
	fullURL := s.baseURL + "/"

	req, err := http.NewRequest("PATCH", fullURL, nil) // Sin body
	if err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error creando request: %w", err)
	}

	// Añadir query params
	q := url.Values{}
	q.Add("Id", wID) // 'Id' con mayúscula, como en tu código
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error actualizando recurso: %w", err)
	}
	defer resp.Body.Close()

	// Decodificamos la respuesta
	var result interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)

	// Verificamos el status
	if resp.StatusCode >= 400 {
		errMsg := fmt.Errorf("[FetchAPI] Error actualizando recurso, status %d", resp.StatusCode)
		if decodeErr == nil {
			return result, errMsg
		}
		return nil, errMsg
	}

	// Status OK, pero error de decodificación
	if decodeErr != nil {
		return nil, fmt.Errorf("[FetchAPI] Error decodificando JSON de respuesta: %w", decodeErr)
	}

	// Todo OK
	return result, nil
}