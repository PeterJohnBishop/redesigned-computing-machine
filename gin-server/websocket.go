package ginserver

import (
	"encoding/json"
	"fmt"
	"net/http"

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

var connections = make(map[*websocket.Conn]bool)

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	connections[conn] = true
	defer delete(connections, conn)

	fmt.Println("Client connected:", conn.RemoteAddr())

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

		default:
			fmt.Println("Unknown event:", msg.Event)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"error","data":"Unknown event"}`))
		}
	}
}
