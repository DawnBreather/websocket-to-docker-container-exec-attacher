package main

import (
  "encoding/hex"
  "fmt"
  "log"
  "net/http"
  "quic_shell_server/docker"

  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
}

// WebSocketAdapter wraps a WebSocket connection to implement io.ReadWriteCloser.
type WebSocketAdapter struct {
  Conn *websocket.Conn
}

// Read reads a message from the WebSocket connection and logs it.
func (wsa *WebSocketAdapter) Read(p []byte) (int, error) {
  messageType, message, err := wsa.Conn.ReadMessage()
  if err != nil {
    log.Printf("Error reading WebSocket message: %v", err)
    return 0, err
  }

  // Log the received message
  if messageType == websocket.TextMessage {
    log.Printf("Received text message: %s", message)
  } else if messageType == websocket.BinaryMessage {
    log.Printf("Received binary message: %s", hex.EncodeToString(message))
  }

  copy(p, message)
  return len(message), nil
}

// Write writes a message to the WebSocket connection and logs it.
func (wsa *WebSocketAdapter) Write(p []byte) (int, error) {
  // Determine if the message is likely text
  if isLikelyText(p) {
    log.Printf("Sending text message: %s", p)
  } else {
    log.Printf("Sending binary message: %s", hex.EncodeToString(p))
  }

  err := wsa.Conn.WriteMessage(websocket.TextMessage, p)
  if err != nil {
    log.Printf("Error writing WebSocket message: %v", err)
    return 0, err
  }
  return len(p), nil
}

// isLikelyText checks if a byte slice is likely to be readable text
func isLikelyText(data []byte) bool {
  // This is a simple check; you can expand it as needed
  for _, b := range data {
    if b < 32 || b > 127 {
      return false // Non-printable characters suggest binary data
    }
  }
  return true
}

// Close closes the WebSocket connection.
func (wsa *WebSocketAdapter) Close() error {
  return wsa.Conn.Close()
}

func main() {
  http.HandleFunc("/ws", handleWebSocket)
  log.Println("WebSocket server listening on :4242")
  log.Fatal(http.ListenAndServe(":4242", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println(err)
    return
  }
  defer conn.Close()

  handleConnection(conn)
}

type containerIDMessage struct {
  ContainerID string `json:"container_id"`
}

func handleConnection(conn *websocket.Conn) {
  log.Printf("Starting a new WebSocket session")

  // Initialize Docker client
  dockerClient, err := docker.NewDockerClient()
  if err != nil {
    log.Printf("Failed to create Docker client: %v", err)
    return
  }

  // Specify the Docker image and command for the container
  image := "docker.io/library/alpine:latest"
  // cmd := []string{"/bin/sh", "-c", "stty -echo && /bin/sh"}
  cmd := []string{"/bin/sh"}

  // Pull the image if necessary
  if err := dockerClient.PullImage(image); err != nil {
    log.Printf("Failed to pull image: %v", err)
    return
  }

  var containerIdMessage containerIDMessage

  err = conn.ReadJSON(&containerIdMessage)
  if err != nil {
    log.Printf("Failed parsing incoming ContainerIDMessage (JSON) (i.e. {\"container_id\":\"some-container-id\"})")
  }

  fmt.Printf("\nIncoming message: %+v\n", containerIdMessage)

  if containerIdMessage.ContainerID == "" {
    // Create and start a new container
    containerIdMessage.ContainerID, err = dockerClient.CreateAndStartContainer(image, cmd)
    if err != nil {
      log.Printf("Failed to create and start container: %v", err)
      return
    }
  }

  // Now, containerID contains the ID of the newly created container
  // Create an exec instance in the container
  execID, err := dockerClient.CreateExecInstance(containerIdMessage.ContainerID, cmd)
  if err != nil {
    log.Printf("Failed to create exec instance: %v", err)
    return
  }

  // Wrap the WebSocket connection in the adapter
  wsAdapter := &WebSocketAdapter{Conn: conn}

  // Attach to the exec instance and set the input/output
  err = conn.WriteJSON(map[string]string{
    "containerId": containerIdMessage.ContainerID,
  })
  if err != nil {
    log.Printf("Failed writing ContainerId { %s } to WebSocket connection", containerIdMessage.ContainerID)
  }
  if err := dockerClient.AttachExecInstance(execID, wsAdapter); err != nil {
    log.Printf("Failed to attach to exec instance: %v", err)
    return
  }

  //// After attaching to the exec instance
  //if err := conn.WriteMessage(websocket.TextMessage, []byte("Session started. Type your commands:\n")); err != nil {
  //	log.Printf("Failed to write to WebSocket: %v", err)
  //}

  defer func() {
    if err := dockerClient.StopAndRemoveContainer(containerIdMessage.ContainerID); err != nil {
      log.Printf("Failed to stop and remove container: %v", err)
    }
  }()

  // WebSocket message handling loop
  for {
    messageType, message, err := conn.ReadMessage()
    if err != nil {
      log.Printf("Read error: %v", err)
      break
    }
    if messageType == websocket.TextMessage {
      // Handle text message
      log.Printf("Received message: %s", string(message))
    }
  }
}
