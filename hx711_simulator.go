package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"simulator/domain/models"
)

type Material struct {
	Tipo    string
	PesoMin float64
	PesoMax float64
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Tipos de materiales que pueden llegar al contenedor
	materiales := []Material{
		{"plastico", 20, 60}, // g
		{"lata", 12, 25},     // g
	}

	currentWeight := 0.0
	maxWeight := 1000.0 // 1 kg máximo antes de vaciar
	filling := true

	for i := 0; i < 100; i++ {
		if filling {
			// Simula detección de material aleatorio
			mat := materiales[rand.Intn(len(materiales))]

			// Peso de ese material
			add := mat.PesoMin + rand.Float64()*(mat.PesoMax-mat.PesoMin)
			currentWeight += add

			fmt.Printf("➕ Se agregó una %s (%.2f g)\n", mat.Tipo, add)

			if currentWeight >= maxWeight {
				filling = false
				fmt.Println("🧺 Contenedor lleno, vaciando...")
			}
		} else {
			// Simula vaciado total del contenedor
			currentWeight -= 150 + rand.Float64()*100
			if currentWeight <= 0 {
				currentWeight = 0
				filling = true
				fmt.Println("♻️ Contenedor vacío, reiniciando llenado...")
			}
		}

		data := models.HX711{
			Prototype_id: "HX_001",
			Weight:       fmt.Sprintf("%.2f", currentWeight),
		}

		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))

		time.Sleep(1 * time.Second)
	}
}
