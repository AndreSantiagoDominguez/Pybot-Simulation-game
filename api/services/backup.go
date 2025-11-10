package services

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Backup es el struct que contiene la URL base y el cliente HTTP.
// Es el equivalente a tu clase `Backup`.
type Backup struct {
	baseURL string
	client  *http.Client
}

// NewBackup es el constructor, equivalente a tu `__init__`.
func NewBackup() *Backup {
	return &Backup{
		baseURL: "https://pybot.aleosh.online/sensors/backup",
		client: &http.Client{
			Timeout: 15 * time.Second, // Timeout para las peticiones HTTP
		},
	}
}

// checkInternetConnection (privado) verifica la conectividad a un host y puerto.
// Usamos una función privada (minúscula) ya que solo la llama Start.
// Esta función es el equivalente a tu `check_internet_conection`
func (b *Backup) checkInternetConnection() bool {
	// Usamos los mismos valores por defecto de tu método
	host := "8.8.8.8"
	port := 53
	timeout := 3 * time.Second

	address := fmt.Sprintf("%s:%d", host, port)

	// net.DialTimeout es el equivalente directo y más idiomático en Go
	// para socket.socket(...).connect() con un timeout.
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		// Si hay un error, no hay conexión
		return false
	}
	// Cerramos la conexión inmediatamente, solo queríamos saber si se podía establecer.
	conn.Close()
	return true
}

// Start es el método principal, equivalente a tu `start`.
func (b *Backup) Start() {
	connection := b.checkInternetConnection()

	if connection {
		fmt.Println("Se hace el backup")

		// --- Este es el bloque que tenías comentado en Python ---
		// Hacemos un POST a la URL base, sin cuerpo (nil)
		resp, err := b.client.Get(b.baseURL+"/")
		if err != nil {
			fmt.Printf("[FetchAPI] Error al hacer el backup: %v\n", err)
			// Si la petición falla (ej. error de red), no podemos leer resp,
			// así que salimos.
			return
		}
		defer resp.Body.Close()

		// Equivalente a resp.raise_for_status()
		isError := resp.StatusCode >= 400
		if isError {
			fmt.Printf("[FetchAPI] Error al hacer el backup, status: %d\n", resp.StatusCode)
		}

		// Intentamos decodificar el JSON, tal como lo hacías en Python
		// (incluso en el bloque de error).
		var result interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
			// Si no es un error de status, pero falla el JSON, lo reportamos.
			if !isError {
				fmt.Printf("[FetchAPI] Error decodificando respuesta: %v\n", decodeErr)
			}
		} else {
			// Imprimimos la respuesta si se decodificó bien.
			fmt.Println(result)
		}
		// --- Fin del bloque ---

	} else {
		// Mantuve tu typo "Baackup"
		fmt.Println("Baackup no disponible")
	}
}