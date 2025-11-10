package sensors

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"pybot-simulator/api/rabbitmq"
	"time"
)

// RealTimeCamera handles reading images and publishing them.
type RealTimeCamera struct {
	publisher    *rabbitmq.RabbitMQPublisher
	imagePaths   []string
	lastSentTime time.Time
}

// NewRealTimeCamera initializes the camera sensor.
func NewRealTimeCamera() (*RealTimeCamera, error) {
	// Initialize RabbitMQ publisher
	publisher, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RabbitMQ publisher: %w", err)
	}

	// Load image paths
	imageDir := "api/dataset/camera"
	files, err := ioutil.ReadDir(imageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read image directory '%s': %w", imageDir, err)
	}

	var imagePaths []string
	for _, file := range files {
		if !file.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(imageDir, file.Name()))
		}
	}

	if len(imagePaths) == 0 {
		log.Println("Warning: No images found in api/dataset/camera")
	}

	rand.Seed(time.Now().UnixNano())

	return &RealTimeCamera{
		publisher:  publisher,
		imagePaths: imagePaths,
	}, nil
}

// PublishRandomImage selects a random image, reads it, and publishes it to RabbitMQ.
func (c *RealTimeCamera) PublishRandomImage() {
	if len(c.imagePaths) == 0 {
		return // No images to send
	}

	// Select a random image path
	randomImagePath := c.imagePaths[rand.Intn(len(c.imagePaths))]

	// Read the image file
	imageData, err := ioutil.ReadFile(randomImagePath)
	if err != nil {
		log.Printf("Error reading image file %s: %v", randomImagePath, err)
		return
	}

	// Publish to RabbitMQ
	sent, err := c.publisher.Send(
		map[string]interface{}{
                        "prototype_id": "a99fd25c7e4a4e2cb5b7a1d1",
                        "detections": map[string]interface{}{
								"cls": 1,
								"conf": 0.4,
							},
                        "image": imageData,
                    }, "cam")
	if err != nil {
		log.Printf("Error sending image to RabbitMQ: %v", err)
	} else if sent {
		log.Printf("Successfully sent image %s to 'cam' queue", randomImagePath)
	}
}

// Close cleans up the camera resources.
func (c *RealTimeCamera) Close() {
	if c.publisher != nil {
		c.publisher.Close()
	}
}
