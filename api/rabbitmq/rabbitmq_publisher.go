package rabbitmq

import (
	"context"       // Para publicar con timeouts
	"encoding/json" // Equivalente a 'import json'
	"fmt"
	"log"
	"os" // Equivalente a 'import os'
	"time"

	// Equivalente a 'from dotenv import load_dotenv'
	"github.com/joho/godotenv"
	// Equivalente a 'import pika'
	amqp "github.com/rabbitmq/amqp091-go"
)

// --- Variables de configuración (cargadas en init) ---
var (
	// CAMBIO: Ahora solo usamos una variable para la URL
	rabbitmqURL       string
	exchangeName      = "amq.topic"
	exchangeType      = "topic"
	defaultRoutingKey string
	// Las variables rabbitmqHost, rabbitmqUser, y rabbitmqPass ya no son necesarias
)

// getEnv es un helper para obtener variables de entorno con un valor por defecto
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// init() se ejecuta automáticamente al iniciar el programa
func init() {
	// load_dotenv()
	if err := godotenv.Load(); err != nil {
		log.Println("[RabbitMQ] Advertencia: No se pudo cargar el archivo .env")
	}

	// CAMBIO: Leemos la variable RABBITMQ_URL
	// El valor por defecto es la URL estándar de RabbitMQ
	rabbitmqURL = getEnv("RABBITMQ_URL", "")

	// Mantenemos la carga de la routing key
	defaultRoutingKey = getEnv("MQTT_TOPIC", "sensores/datos")
}

// RabbitMQPublisher es el struct, equivalente a tu 'class'
type RabbitMQPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQPublisher es el constructor, equivalente a tu '__init__'
func NewRabbitMQPublisher() (*RabbitMQPublisher, error) {
	// Conectamos (pika.BlockingConnection)
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Printf("[RabbitMQ] Error al conectar: %v", err)
		return nil, fmt.Errorf("error al conectar: %w", err)
	}

	// Creamos el canal (self.connection.channel())
	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // Limpiar conexión si falla el canal
		log.Printf("[RabbitMQ] Error al crear canal: %v", err)
		return nil, fmt.Errorf("error al crear canal: %w", err)
	}

	// Declaramos el exchange (self.channel.exchange_declare)
	err = ch.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		log.Printf("[RabbitMQ] Error al declarar exchange: %v", err)
		return nil, fmt.Errorf("error al declarar exchange: %w", err)
	}

	fmt.Println("[RabbitMQ] Conectado y exchange declarado")
	return &RabbitMQPublisher{
		connection: conn,
		channel:    ch,
	}, nil
}

// Send es el método para publicar, equivalente a tu 'send'
// 'payload' es 'interface{}' para aceptar cualquier struct o map (como un dict)
func (r *RabbitMQPublisher) Send(payload interface{}, routingKey string) error {
	// if not self.connection or self.connection.is_closed:
	if r.connection == nil || r.connection.IsClosed() {
		log.Println("[RabbitMQ] Conexión no disponible")
		return fmt.Errorf("conexión no disponible")
	}

	// Usar routing key por defecto si no se provee
	if routingKey == "" {
		routingKey = defaultRoutingKey
	}

	// message = json.dumps(payload)
	message, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[RabbitMQ] Error codificando JSON: %v", err)
		return fmt.Errorf("error codificando JSON: %w", err)
	}

	// Usamos un contexto para poner un timeout a la publicación
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// self.channel.basic_publish(...)
	err = r.channel.PublishWithContext(
		ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)

	if err != nil {
		// Tu código Python solo imprime el error, pero en Go es mejor retornarlo
		log.Printf("[RabbitMQ] Error publicando mensaje: %v", err)
		return fmt.Errorf("error publicando mensaje: %w", err)
	}

	// print(f"[RabbitMQ] Enviado {routing_key}:{message}") // Omitido ya que estaba comentado
	return nil
}

// Close cierra el canal y la conexión
func (r *RabbitMQPublisher) Close() {
	closed := false
	if r.channel != nil {
		r.channel.Close() // Ignoramos errores al cerrar
		closed = true
	}
	if r.connection != nil && !r.connection.IsClosed() {
		r.connection.Close() // Ignoramos errores al cerrar
		closed = true
	}

	if closed {
		fmt.Println("[RabbitMQ] Conexión cerrada")
	}
}
