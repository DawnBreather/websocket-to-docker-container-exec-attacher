package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"quic_shell_server/db"
	"quic_shell_server/docker"
	"quic_shell_server/utils"
	"strconv"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketAdapter struct {
	Conn *websocket.Conn
}

func (wsa *WebSocketAdapter) Read(p []byte) (int, error) {
	return readWebSocketMessage(wsa.Conn, p)
}

func (wsa *WebSocketAdapter) Write(p []byte) (int, error) {
	return writeWebSocketMessage(wsa.Conn, p)
}

func (wsa *WebSocketAdapter) Close() error {
	return wsa.Conn.Close()
}

func main() {
	docker.InitializeDockerClient()

	// Start expired Docker containers cleanup job
	go func() {
		for {
			time.Sleep(6 * time.Second)
			err := docker.StopAndDeleteContainersWithTtlExpired(context.Background(), docker.Client)
			if err != nil {
				log.Printf("Failed removing expired Docker containers: %v", err)
			}
		}
	}()

	startWebServer()
}

//func initializeDockerClient() {
//	var err error
//	dockerClient, err = docker.NewDockerClient()
//	if err != nil {
//		log.Fatalf("Failed to create Docker client: %v", err)
//	}
//}

func startWebServer() {
	//http.HandleFunc("/ws", webSocketHandler)
	//http.HandleFunc("/api/admin/container/extend_ttl", apiAdminContainerExtendTtlHandler)
	//log.Println("WebSocket server listening on :4242")
	//log.Fatal(http.ListenAndServe(":4242", nil))

	r := mux.NewRouter()
	r.HandleFunc("/ws", webSocketHandler)
	r.HandleFunc("/api/admin/container/{id}/extend_ttl/{minutes}", apiAdminContainerExtendTtlHandler).Methods("POST")
	r.HandleFunc("/api/admin/container/{id}", apiAdminContainerHandler).Methods("GET")
	log.Println("WebSocket server listening on :4242")
	log.Fatal(http.ListenAndServe(":4242", r))
}

func apiAdminContainerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	container, err := docker.GetContainerById(context.Background(), docker.Client, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}

	containerRunningTime, err := docker.GetContainerRunningTime(context.Background(), docker.Client, id)
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}

	fmt.Fprintf(w, string(utils.MustMarshal(containerIDMessage{
		ContainerID:  id,
		Image:        container.Image,
		TTL:          db.GetTtlInMinutesByContainerId(id),
		CMD:          container.Command,
		RemainingTTL: db.GetTtlInMinutesByContainerId(id) - containerRunningTime,
	})))
}

func apiAdminContainerExtendTtlHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	container, err := docker.GetContainerById(context.Background(), docker.Client, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}

	containerRunningTime, err := docker.GetContainerRunningTime(context.Background(), docker.Client, id)
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}

	minutesString := vars["minutes"]
	minutes, err := strconv.Atoi(minutesString)
	if err != nil {
		log.Printf("Failed to parse minutes { %s } into type of Int: %v", minutesString, err)
	} else {
		db.SetTtlInMinutesByContainerId(id, minutes+db.GetTtlInMinutesByContainerId(id))
	}

	fmt.Fprintf(w, string(utils.MustMarshal(containerIDMessage{
		ContainerID:  id,
		Image:        container.Image,
		TTL:          db.GetTtlInMinutesByContainerId(id),
		CMD:          container.Command,
		RemainingTTL: db.GetTtlInMinutesByContainerId(id) - containerRunningTime,
	})))
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgradeToWebSocket(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	manageWebSocketConnection(conn)
}

func upgradeToWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func manageWebSocketConnection(conn *websocket.Conn) {
	log.Printf("Starting a new WebSocket session")

	//image, cmd := prepareDockerEnvironment()
	//if err := pullDockerImage(docker.Client, image); err != nil {
	//	return
	//}

	containerIdMessage, err := readContainerIDMessage(conn)
	if err != nil {
		return
	}

	containerID, execID := setupContainerAndExec(docker.Client, containerIdMessage) //, image, cmd)
	if containerID == "" || execID == "" {
		return
	}

	db.SetWsConnectionByContainerId(containerID, conn)

	err = conn.WriteJSON(containerIDMessage{
		ContainerID: containerID,
	})
	if err != nil {
		log.Printf("Failed to send ContainerID { %s } to the client: %v", containerID, err)
	}

	attachToExecInstance(conn, docker.Client, execID)
	handleWebSocketMessages(conn, docker.Client, containerID)
}

func readWebSocketMessage(conn *websocket.Conn, p []byte) (int, error) {
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Error reading WebSocket message: %v", err)
		return 0, err
	}

	logReceivedMessage(messageType, message)
	copy(p, message)
	return len(message), nil
}

func writeWebSocketMessage(conn *websocket.Conn, p []byte) (int, error) {
	logSendMessage(p)

	err := conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Printf("Error writing WebSocket message: %v", err)
		return 0, err
	}
	return len(p), nil
}

func logReceivedMessage(messageType int, message []byte) {
	if messageType == websocket.TextMessage {
		log.Printf("Received text message: %s", message)
	} else if messageType == websocket.BinaryMessage {
		log.Printf("Received binary message: %s", hex.EncodeToString(message))
	}
}

func logSendMessage(p []byte) {
	if isLikelyText(p) {
		log.Printf("Sending text message: %s", p)
	} else {
		log.Printf("Sending binary message: %s", hex.EncodeToString(p))
	}
}

func isLikelyText(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 127 {
			return false
		}
	}
	return true
}

//func prepareDockerEnvironment() (string, []string) {
//	image := "docker.io/library/alpine:latest"
//	cmd := []string{"/bin/sh"}
//	return image, cmd
//}

//func pullDockerImage(client *docker.DockerClient, image string) error {
//	if err := client.PullImage(image); err != nil {
//		log.Printf("Failed to pull image: %v", err)
//		return err
//	}
//	return nil
//}

func readContainerIDMessage(conn *websocket.Conn) (containerIDMessage, error) {
	var msg containerIDMessage
	err := conn.ReadJSON(&msg)
	if err != nil {
		log.Printf("Failed parsing incoming ContainerIDMessage: %v", err)
		return msg, err
	}
	fmt.Printf("\nIncoming message: %+v\n", msg)
	return msg, nil
}

// func setupContainerAndExec(client *docker.DockerClient, msg containerIDMessage, image string, cmd []string) (string, string) {
func setupContainerAndExec(client *docker.DockerClient, msg containerIDMessage) (string, string) {
	containerID := msg.ContainerID
	image := func() string {
		if msg.Image == "" {
			return docker.DefaultDockerImage
		} else {
			elements := strings.Split(msg.Image, "/")
			if strings.Contains(elements[0], ".") {
				return msg.Image
			} else {
				return fmt.Sprintf("docker.io/library/%s", msg.Image)
			}
		}
	}()

	cmd := func() []string {
		if len(msg.CMD) == 0 {
			return []string{"/bin/sh"}
		}
		if msg.CMD == "" {
			return []string{"/bin/sh"}
		}
		return []string{msg.CMD}
	}()

	if containerID == "" {
		err := client.PullImage(image)
		if err != nil {
			log.Printf("Failed pulling image { %s }: %v", image, err)
			return "", ""
		}
		containerID, err = client.CreateAndStartContainer(image, cmd, msg.TTL) //(image, cmd)
		if err != nil {
			log.Printf("Failed to create and start container: %v", err)
			return "", ""
		}
	}

	execID, err := client.CreateExecInstance(containerID, cmd) //cmd)
	if err != nil {
		log.Printf("Failed to create exec instance: %v", err)
		return "", ""
	}

	return containerID, execID
}

func attachToExecInstance(conn *websocket.Conn, client *docker.DockerClient, execID string) {
	wsAdapter := &WebSocketAdapter{Conn: conn}
	if err := client.AttachExecInstance(execID, wsAdapter); err != nil {
		log.Printf("Failed to attach to exec instance: %v", err)
	}
}

func handleWebSocketMessages(conn *websocket.Conn, client *docker.DockerClient, containerID string) {
	//defer stopAndRemoveContainer(client, containerID)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		if messageType == websocket.TextMessage {
			log.Printf("Received message: %s", string(message))
		}
	}
}

func stopAndRemoveContainer(client *docker.DockerClient, containerID string) {
	if err := client.StopAndRemoveContainer(containerID); err != nil {
		log.Printf("Failed to stop and remove container: %v", err)
	}
}

type containerIDMessage struct {
	ContainerID  string `json:"container_id,omitempty"`
	Image        string `json:"image,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
	CMD          string `json:"cmd,omitempty"`
	RemainingTTL int    `json:"remaining_ttl,omitempty"`
}
