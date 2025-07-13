package ginserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default
		return true
	},
}

var (
	connections   = make(map[string]*websocket.Conn)
	connectionsMu sync.Mutex
)

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	ip := conn.RemoteAddr().String()
	connectionsMu.Lock()
	connections[ip] = conn
	connectionsMu.Unlock()

	defer func() {
		connectionsMu.Lock()
		delete(connections, ip)
		connectionsMu.Unlock()
	}()

	fmt.Println("Client connected:", ip)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			fmt.Println("JSON Unmarshal Error:", err)
			continue
		}

		switch msg.Event {
		case "connect":
			fmt.Printf("New Connection: %s\n", msg.Data)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"system","data":"Welcome!"}`))

		case "message":
			fmt.Printf("Received message: %s\n", msg.Data)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"echo","data":"`+msg.Data+`"}`))

		case "upload":
			fmt.Printf("File uploaded: %s\n", msg.Data)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"upload","data":"`+msg.Data+`"}`))

		default:
			fmt.Println("Unknown event:", msg.Event)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"error","data":"Unknown event"}`))
		}
	}
}

func PeerDiscovery() {
	myPodIP := os.Getenv("MY_POD_IP")
	serviceDNS := "server-headless.default.svc.cluster.local"
	port := 8080

	go func() {
		for {
			addrs, err := net.LookupHost(serviceDNS)
			if err != nil {
				log.Printf("DNS resolution failed: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, ip := range addrs {
				if ip == myPodIP {
					continue // Skip self
				}

				wsURL := fmt.Sprintf("ws://%s:%d/ws", ip, port)

				connectionsMu.Lock()
				_, alreadyConnected := connections[ip]
				connectionsMu.Unlock()

				if alreadyConnected {
					continue
				}

				go connectWithRetry(ip, wsURL)
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func connectWithRetry(ip, wsURL string) {
	var conn *websocket.Conn
	var err error

	for attempt := 1; attempt <= 5; attempt++ {
		conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			log.Printf("Connected to %s", wsURL)
			connectionsMu.Lock()
			connections[ip] = conn
			connectionsMu.Unlock()
			return
		}

		log.Printf("Attempt %d: Failed to connect to %s: %v", attempt, wsURL, err)
		time.Sleep(time.Duration(attempt*2) * time.Second)
	}

	log.Printf("Failed to connect to %s after retries", wsURL)
}
