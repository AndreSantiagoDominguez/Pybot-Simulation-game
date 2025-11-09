package models

import (

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients  map[string]map[string]*websocket.Conn  // Sensor -> Prototype_id -> connectionws
	register chan Registration
	send 	 chan SensorMessage
}

func NewHub() *Hub {
	h:= &Hub{
		clients: make(map[string]map[string]*websocket.Conn),
		register: make(chan Registration),
		send: make(chan SensorMessage),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case reg := <- h.register:
			if _, ok := h.clients[reg.Sensor]; !ok {
				h.clients[reg.Sensor] = make(map[string]*websocket.Conn)
			}
			h.clients[reg.Sensor][reg.Prototype_id] = reg.Conn
			
		case msg := <- h.send: 
			if conns, ok := h.clients[msg.Sensor]; ok {
				if conn, ok2 := conns[msg.Prototype_id]; ok2 {
					conn.WriteMessage(websocket.TextMessage, msg.Payload)
				}
			}
		}

	}
}

func (h* Hub) Register(sensor, prototype_id string, conn *websocket.Conn) {
	h.register <- Registration{Sensor: sensor, Prototype_id: prototype_id, Conn: conn}
}

func (h *Hub) Send(msg SensorMessage) {
	h.send <- msg
}