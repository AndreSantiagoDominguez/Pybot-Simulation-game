package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
	"simulator/domain/models"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 1000; i++ {
		data := models.GPS{
			Prototype_id: fmt.Sprintf("GPS_%03d", rand.Intn(10)),
			Lat:          fmt.Sprintf("%.6f", 19.4326+rand.Float64()/1000), // CDMX aprox
			Lon:          fmt.Sprintf("%.6f", -99.1332+rand.Float64()/1000),
			Spd:          fmt.Sprintf("%.2f", 10+rand.Float64()*50), // km/h
			Date:         time.Now().Format("2006-01-02"),
			UTC:          time.Now().Format("15:04:05"),
			Alt:          fmt.Sprintf("%.1f", 2200+rand.Float64()*50),
			Sats:         fmt.Sprintf("%d", 5+rand.Intn(5)),
		}

		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
		time.Sleep(2 * time.Second)
	}
}
