package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	//serial "go.bug.st/serial.v1"
	"github.com/larsgk/serial"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ListEvent struct {
	Type string
	Data []interface{}
}

func sendJsonEvent(w http.ResponseWriter, daEvent interface{}) {
	js, err := json.Marshal(daEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-cache")
	w.Write(js)
}

func handleListCommPortsEvent(w http.ResponseWriter, r *http.Request) {
	reply := ListEvent{Type: "CommPorts"}

	start := time.Now()
	ports, _ := GetSerialPortList()
	log.Printf("Serial port PnP lookup took %s\n", time.Since(start))

	for _, port := range ports {
		reply.Data = append(reply.Data, port)
	}

	sendJsonEvent(w, reply)
}

var upgrader = websocket.Upgrader{}

func handleWSConnect(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle potential panics better
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	devicePath := r.FormValue("path")
	baudRate, err := strconv.Atoi(r.FormValue("baudrate"))
	if err != nil || baudRate < 1 {
		baudRate = 9600 // Default to 9600
	}

	log.Printf("Requesting WebSocket <-> Serial connection to %v (baudrate:%d)", devicePath, baudRate)

	if len(devicePath) == 0 {
		http.Error(w, "devicePath missing", http.StatusInternalServerError)
		return
	}

	// TODO: Make this configurable (from get params) - although I never saw anything in use but xxxN81 ;)
	// mode := &serial.Mode{
	// 	BaudRate: baudRate,
	// 	Parity:   serial.NoParity,
	// 	DataBits: 8,
	// 	StopBits: serial.OneStopBit,
	// }

	// port, err := serial.Open(devicePath, mode)

	port, err := serial.Open(&serial.Config{Address: devicePath, BaudRate: baudRate})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade:", err)
		return
	}

	log.Print("WebSocket ok")

	buff := make([]byte, 512) // 512 bytes is max for high speed USB

	go func() {
		for {
			n, err := port.Read(buff)
			if err != nil {
				log.Println(err.Error())
				break
			}
			if n > 0 {
				_ = c.WriteMessage(websocket.BinaryMessage, buff[:n])
			} else {
				log.Println("EOF")
				break
			}
			// log.Printf("Length: %d, data: %v", n, buff[:n])
		}
		log.Println("[Serial port closed] - close WS...")
		c.Close()
	}()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("[WS] Error:", err)
				break
			}
			log.Printf("[WS -> Serial] Sending: %v, '%s'", message, string(message))
			port.Write(message)
		}
		log.Println("[WS closed] - close serial port...")
		port.Close()
	}()
}
