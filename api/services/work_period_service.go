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

type WorkPeriodService struct {
	baseURL string
	client  *http.Client
}

func NewWorkPeriodService() *WorkPeriodService {
	return &WorkPeriodService{
		baseURL: "https://pybot.aleosh.online/sensors/workPeriods",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *WorkPeriodService) GetLastHourPeriod() (interface{}, error) {
	resp, err := s.client.Get(s.baseURL + "/")
	if err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error listando datos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("[FetchAPI] Error de status %d", resp.StatusCode)
	}

	// Decodifica el JSON a un tipo genérico interface{}
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error decodificando JSON: %w", err)
	}
	return result, nil
}

// GetDistanceAndWeight obtiene distancia y peso por ID.
func (s *WorkPeriodService) GetDistanceAndWeight(id string) (interface{}, error) {
	// Construimos la URL con el query param
	fullURL := fmt.Sprintf("%s/readingsGlobal?id=%s", s.baseURL, id)

	resp, err := s.client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error listando datos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("[FetchAPI] Error de status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error decodificando JSON: %w", err)
	}
	return result, nil
}

// helperJSONRequest es una función privada para manejar POST, PUT, etc. con JSON.
func (s *WorkPeriodService) helperJSONRequest(method, url string, payload interface{}) (interface{}, error) {
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
		return nil, fmt.Errorf("[FetchAPI] Error creando recurso (%s): %w", method, err)
	}
	defer resp.Body.Close()

	// En Python, intentabas retornar resp.json() incluso en el 'except'.
	// En Go, si el status es de error, es mejor devolver solo el error.
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("[FetchAPI] Error de status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("[FetchAPI] Error decodificando JSON de respuesta: %w", err)
	}
	return result, nil
}

// CreateNewPeriod crea un nuevo período (POST).
// El payload puede ser un 'map[string]interface{}' o un struct.
func (s *WorkPeriodService) CreateNewPeriod(payload interface{}) (interface{}, error) {
	fullURL := s.baseURL + "/"
	return s.helperJSONRequest("POST", fullURL, payload)
}

// CreateNewReading crea una nueva lectura (POST).
func (s *WorkPeriodService) CreateNewReading(payload interface{}) (interface{}, error) {
	fullURL := s.baseURL + "/readings"
	return s.helperJSONRequest("POST", fullURL, payload)
}

// UpdateLastPeriod actualiza el último período (PATCH).
// Devuelve true en éxito, o false y un error en fallo.
func (s *WorkPeriodService) UpdateLastPeriod(endHour, id string) (bool, error) {
	fullURL := s.baseURL + "/"

	req, err := http.NewRequest("PATCH", fullURL, nil) // Sin body
	if err != nil {
		return false, fmt.Errorf("[FetchAPI] Error creando request: %w", err)
	}

	// Añadir query params
	q := url.Values{}
	q.Add("endHour", endHour)
	q.Add("id", id)
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("[FetchAPI] Error actualizando id=%s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("[FetchAPI] Error actualizando id=%s, status %d", id, resp.StatusCode)
	}

	return true, nil
}

// UpdateLastReadig actualiza la última lectura (PUT).
// Nota: Mantuve el typo "Readig" de tu código Python.
func (s *WorkPeriodService) UpdateLastReadig(payload interface{}) (bool, error) {
	fullURL := s.baseURL + "/"

	// Este es un PUT con body, pero sin esperar JSON de vuelta
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("[FetchAPI] Error codificando JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("[FetchAPI] Error creando request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("[FetchAPI] Error actualizando (PUT): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("[FetchAPI] Error actualizando (PUT), status %d", resp.StatusCode)
	}

	return true, nil
}
