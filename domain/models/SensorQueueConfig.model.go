package models

// SensorConfig reúne parámetros de exchange, queue y bind
type SensorConfig struct {
	Exchange string
	Queue    string
	RoutingKey string
}

// SensorConfigs devuelve configuración de las 3 colas
func SensorConfigs() []SensorConfig {
	return []SensorConfig{
		{Exchange: "amq.topic", Queue: "sensor_HX", RoutingKey: "hx"},
		{Exchange: "amq.topic", Queue: "sensor_NEO", RoutingKey: "neo"},
		{Exchange: "amq.topic", Queue: "sensor_CAM", RoutingKey: "cam"},
	}
}