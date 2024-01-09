package httphandlers

import (
	"encoding/hex"
	"github.com/gorilla/websocket"
	"log"
)

var WsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketAdapter struct {
	Conn *websocket.Conn
}

func (wsa *WebSocketAdapter) Read(p []byte) (n int, err error) {
	var messageType int
	var message []byte
	messageType, message, err = wsa.Conn.ReadMessage()
	if err != nil {
		log.Printf("Error reading WebSocket message: %v", err)
	} else {
		wsa.logMessage("Received", messageType, message)
		n = copy(p, message)
	}
	return
}

//func (wsa *WebSocketAdapter) Write(p []byte) (n int, err error) {
//	wsa.logMessage("Sending", websocket.TextMessage, p)
//	err = wsa.Conn.WriteMessage(websocket.TextMessage, p)
//	if err != nil {
//		log.Printf("Error writing WebSocket message: %v", err)
//	} else {
//		n = len(p)
//	}
//	return
//}

func (wsa *WebSocketAdapter) Write(p []byte) (n int, err error) {
	var messageType int
	//if wsa.isLikelyText(p) {
	//	messageType = websocket.TextMessage
	//} else {
	//	messageType = websocket.BinaryMessage
	//}

	messageType = websocket.BinaryMessage

	wsa.logMessage("Sending", messageType, p)
	err = wsa.Conn.WriteMessage(messageType, p)
	if err != nil {
		log.Printf("Error writing WebSocket message: %v", err)
	} else {
		n = len(p)
	}
	return
}

func (wsa *WebSocketAdapter) Close() error {
	return wsa.Conn.Close()
}

func (wsa *WebSocketAdapter) logMessage(action string, messageType int, message []byte) {
	messageTypes := map[int]string{
		websocket.TextMessage:   "text",
		websocket.BinaryMessage: "binary",
	}
	msgType, exists := messageTypes[messageType]
	if !exists {
		msgType = "unknown"
	}
	msgContent := message
	if msgType == "binary" {
		msgContent = []byte(hex.EncodeToString(message))
	}
	log.Printf("%s %s message: %s", action, msgType, msgContent)
}

func (wsa *WebSocketAdapter) isLikelyText(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 127 {
			return false
		}
	}
	return true
}
