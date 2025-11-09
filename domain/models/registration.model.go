package models

import "github.com/gorilla/websocket"

type Registration struct {
	Sensor       string
	Prototype_id string `json:"prototype_id"`
	Conn         *websocket.Conn
}