package httphandlers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"quic_shell_server/db"
	"quic_shell_server/docker"
	"strings"
)

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
	conn, err := WsUpgrader.Upgrade(w, r, nil)
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
			return []string{"/bin/bash"}
		}
		if msg.CMD == "" {
			return []string{"/bin/bash"}
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

type containerIDMessage struct {
	ContainerID  string `json:"container_id,omitempty"`
	Image        string `json:"image,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
	CMD          string `json:"cmd,omitempty"`
	RemainingTTL int    `json:"remaining_ttl,omitempty"`
}
