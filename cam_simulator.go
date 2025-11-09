package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
	"simulator/domain/models"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	detections := []string{"plastico", "lata"}

	
	imageDir := "./images"

	
	files, err := os.ReadDir(imageDir)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 1; i++ {
		detection := detections[rand.Intn(len(detections))]
		file := files[rand.Intn(len(files))]

		
		imgPath := filepath.Join(imageDir, file.Name())
		imgBytes, err := os.ReadFile(imgPath)
		if err != nil {
			fmt.Println("Error leyendo imagen:", err)
			continue
		}
		imgBase64 := base64.StdEncoding.EncodeToString(imgBytes)

		data := models.CAM{
			Prototype_id: fmt.Sprintf("CAM_%03d", rand.Intn(10)),
			Detections:   detection,
			Image:        imgBase64,
		}

		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
		time.Sleep(2 * time.Second)
	}
}
