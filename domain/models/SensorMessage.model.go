package models

type SensorMessage struct {
	Sensor       string 
	Prototype_id string `json:"prototype_id"`
	Payload      []byte
}
