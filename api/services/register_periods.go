package services

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// RegisterPeriods es el equivalente a tu clase.
// Mantiene el estado y los servicios.
type RegisterPeriods struct {
	serviceWorkPeriods *WorkPeriodService
	serviceSensors     *SensorRegisterService
	prototypeID        string
	actualPeriodID     int64
	lastPeriodID       int64
	lastHourPeriod     string
	wIDCANS            int64
	wIDPET             int64
}

// NewRegisterPeriods es el constructor, equivalente a tu __init__.
func NewRegisterPeriods() (*RegisterPeriods, error) {
	// Cargar .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Advertencia: No se pudo cargar el archivo .env", err)
		// Podrías decidir si es un error fatal o no
	}

	return &RegisterPeriods{
		serviceWorkPeriods: NewWorkPeriodService(),
		serviceSensors:     NewSensorRegisterService(),
		prototypeID:        os.Getenv("ID_PROTOTYPE"),
	}, nil
}

func getFloat(data map[string]interface{}, key string, defaultVal float64) float64 {
	val, ok := data[key]
	if !ok {
		return defaultVal
	}
	// JSON decodifica números como float64
	floatVal, ok := val.(float64)
	if !ok {
		return defaultVal
	}
	return floatVal
}

func getString(data map[string]interface{}, key string, defaultVal string) string {
	val, ok := data[key]
	if !ok {
		return defaultVal
	}
	strVal, ok := val.(string)
	if !ok {
		return defaultVal
	}
	return strVal
}

func (r *RegisterPeriods) StatusPeriod() (bool, error) {
	res, err := r.serviceWorkPeriods.GetLastHourPeriod()
	if err != nil {
		return false, fmt.Errorf("error en GetLastHourPeriod: %w", err)
	}

	// Replicamos la lógica de `res.get('last_period').get('period_id')`
	// Esto es mucho más verboso en Go por la seguridad de tipos.
	resMap, ok := res.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("respuesta no es un mapa")
	}

	lastPeriodVal, ok := resMap["last_period"]
	if !ok {
		return false, fmt.Errorf("respuesta no contiene 'last_period'")
	}

	lastPeriod, ok := lastPeriodVal.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("'last_period' no es un mapa")
	}

	// JSON decodifica números como float64 por defecto en interface{}
	periodIDFloat, ok := lastPeriod["period_id"].(float64)
	if !ok {
		return false, fmt.Errorf("'period_id' no encontrado o no es número")
	}

	periodID := int64(periodIDFloat)
	if periodID == 0 {
		return true, nil
	}

	// Guardamos el estado si el período no es 0
	r.lastPeriodID = periodID
	r.lastHourPeriod = getString(lastPeriod, "last_hour", "")

	return false, nil
}

// CreateNewPeriod crea un nuevo período de trabajo.
func (r *RegisterPeriods) CreateNewPeriod() error {
	startHour := time.Now().UTC()
	dayWork := startHour.Format("Mon") // %a = 'Mon', 'Tue', etc.

	dBody := map[string]interface{}{
		"period_id":    0,
		"start_hour":   startHour.Format(time.RFC3339), // Formato ISO
		"end_hour":     "",
		"day_work":     dayWork,
		"prototype_id": r.prototypeID,
	}

	res, err := r.serviceWorkPeriods.CreateNewPeriod(dBody)
	if err != nil {
		return fmt.Errorf("error en CreateNewPeriod: %w", err)
	}

	// Replicamos `res.get('data').get('work_periods_id')`
	resMap, ok := res.(map[string]interface{})
	if !ok {
		return fmt.Errorf("respuesta de CreateNewPeriod no es un mapa")
	}
	data, ok := resMap["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("respuesta no contiene 'data'")
	}

	idFloat := getFloat(data, "work_periods_id", 0)
	id := int64(idFloat)

	if id != 0 {
		r.actualPeriodID = id
		fmt.Printf("p_id en cnp: %d", r.actualPeriodID)
	} else {
		fmt.Println("Ocurrio un error al crear el periodo (ID fue 0)")
		return fmt.Errorf("API devolvió un ID de 0")
	}
	return nil
}

// RegisterWeigh registra una lectura de peso.
func (r *RegisterPeriods) RegisterWeigh(weight float64) error {
	hourPeriod := time.Now().UTC().Format(time.RFC3339)
	// Go no tiene round(f, 4). Se hace así:
	roundedWeight := math.Round(weight*10000) / 10000

	dBody := map[string]interface{}{
		"weight_data_id": 0,
		"period_id":      r.actualPeriodID,
		"Hour_period":    hourPeriod,
		"Weight":         roundedWeight,
	}

	_, err := r.serviceSensors.RegisterWeightData(dBody)
	if err != nil {
		return fmt.Errorf("error en RegisterWeightData: %w", err)
	}
	return nil
}

// RegisterGPS registra datos de GPS.
// Nota: data es map[string]interface{} para simular el dict de Python.
func (r *RegisterPeriods) RegisterGPS(data map[string]interface{}) error {
	// Replicamos la lógica `or` de Python usando nuestros helpers
	dateStr := getString(data, "date", "2015-07-13")
	utcStr := getString(data, "UTC", "")

	var hourUTC string
	// Esta lógica replica tu `or '2025-07-12T...'`
	if dateStr == "2015-07-13" || utcStr == "" {
		hourUTC = "2025-07-12T20:14:07.057608+00:00"
	} else {
		hourUTC = fmt.Sprintf("%sT%s+00:00", dateStr, utcStr)
	}
	fmt.Printf("p_id en rgps: %d", r.actualPeriodID)
	dBody := map[string]interface{}{
		//"gps_data_id": 0,
		"period_id":   r.actualPeriodID,
		"latitude":    getFloat(data, "lat", 0.0),
		"longitude":   getFloat(data, "lon", 0.0),
		"altitude":    getFloat(data, "alt", 0.0),
		"speed":       getFloat(data, "spd", 0.0),
		"date_gps":    dateStr,
		"hour_UTC":    hourUTC,
	}

	_, err := r.serviceSensors.RegisterGPSData(dBody)
	if err != nil {
		return fmt.Errorf("error en RegisterGPSData: %w", err)
	}
	return nil
}

// CreateVoidReading crea una lectura vacía para el período actual.
func (r *RegisterPeriods) CreateVoidReading() error {
	dBody := map[string]interface{}{
		"period_id":         r.actualPeriodID,
		"distance_traveled": 0.0,
		"weight_waste":      0.0,
	}
	_, err := r.serviceWorkPeriods.CreateNewReading(dBody)
	if err != nil {
		return fmt.Errorf("error en CreateNewReading: %w", err)
	}
	return nil
}

// CompleteLastPeriod completa el período anterior y crea uno nuevo.
func (r *RegisterPeriods) CompleteLastPeriod() error {
	id := strconv.FormatInt(r.lastPeriodID, 10) // Convierte int64 a string

	res, err := r.serviceWorkPeriods.GetDistanceAndWeight(id)
	if err != nil {
		return fmt.Errorf("error en GetDistanceAndWeight: %w", err)
	}
	fmt.Println(res) // Imprime la respuesta

	resMap, ok := res.(map[string]interface{})
	if !ok {
		return fmt.Errorf("respuesta de GetDistanceAndWeight no es un mapa")
	}

	res1, err := r.serviceWorkPeriods.UpdateLastPeriod(r.lastHourPeriod, id)
	if err != nil {
		return fmt.Errorf("error en UpdateLastPeriod: %w", err)
	}
	fmt.Println(res1) // Imprime la respuesta

	lastReading, ok := resMap["last_reading"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("respuesta no contiene 'last_reading'")
	}

	dBody := map[string]interface{}{
		"period_id":         r.lastPeriodID,
		"distance_traveled": getFloat(lastReading, "distance_traveled", 0.0),
		"weight_waste":      getFloat(lastReading, "weight_waste", 0.0),
	}

	res3, err := r.serviceWorkPeriods.UpdateLastReadig(dBody)
	if err != nil {
		return fmt.Errorf("error en UpdateLastReadig: %w", err)
	}
	fmt.Println(res3) // Imprime la respuesta

	// Encadenamos la creación del nuevo período
	if err := r.CreateNewPeriod(); err != nil {
		return fmt.Errorf("error en CreateNewPeriod (parte de CompleteLastPeriod): %w", err)
	}
	if err := r.CreateVoidReading(); err != nil {
		return fmt.Errorf("error en CreateVoidReading (parte de CompleteLastPeriod): %w", err)
	}
	return nil
}

// CreateWasteCollection crea un registro de recolección de basura.
func (r *RegisterPeriods) CreateWasteCollection(wasteID int64) error {
	dBody := map[string]interface{}{
		"waste_collection_id": 0,
		"period_id":           r.actualPeriodID,
		"amount":              1,
		"waste_id":            wasteID,
	}

	res, err := r.serviceSensors.RegisterWasteCollection(dBody)
	if err != nil {
		return fmt.Errorf("error en RegisterWasteCollection: %w", err)
	}

	// Replicamos `res.get('data').get('waste_collection_id')`
	resMap, ok := res.(map[string]interface{})
	if !ok { return fmt.Errorf("respuesta de RegisterWasteCollection no es un mapa") }
	data, ok := resMap["data"].(map[string]interface{})
	if !ok { return fmt.Errorf("respuesta no contiene 'data'") }

	id := int64(getFloat(data, "waste_collection_id", 0))

	if wasteID == 1 {
		r.wIDPET = id
	} else {
		r.wIDCANS = id
	}
	return nil
}

// UpdateWasteCollection actualiza una recolección.
func (r *RegisterPeriods) UpdateWasteCollection(wasteCollectionID int64) error {
	idStr := strconv.FormatInt(wasteCollectionID, 10)
	res, err := r.serviceSensors.UpdateWasteCollection(idStr)
	if err != nil {
		fmt.Println(res) // Imprime respuesta aunque haya error
		return fmt.Errorf("error en UpdateWasteCollection: %w", err)
	}
	fmt.Println(res)
	return nil
}

// GetIdWasteCollectionPET es un "getter" simple.
func (r *RegisterPeriods) GetIdWasteCollectionPET() int64 {
	return r.wIDPET
}

// GetIdWasteCollectionCANS es un "getter" simple.
func (r *RegisterPeriods) GetIdWasteCollectionCANS() int64 {
	return r.wIDCANS
}

// GetActualPeriodID es un "getter" para el ID del período actual.
func (r *RegisterPeriods) GetActualPeriodID() int64 {
	return r.actualPeriodID
}