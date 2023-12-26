package db

import "github.com/gorilla/websocket"

var containersTtlMap = map[string]int{}
var containersWsConnectionsMap = map[string][]*websocket.Conn{}

func GetTtlInMinutesByContainerId(containerId string) (minutes int) {
	if val, ok := containersTtlMap[containerId]; ok {
		return val
	}
	return 0
}

func SetTtlInMinutesByContainerId(containerId string, ttlMinutes int) {
	containersTtlMap[containerId] = ttlMinutes
}

func SetWsConnectionByContainerId(containerId string, conn *websocket.Conn) {
	containersWsConnectionsMap[containerId] = append(containersWsConnectionsMap[containerId], conn)
}

func GetWsConnectionByContainerId(containerId string) []*websocket.Conn {
	if val, ok := containersWsConnectionsMap[containerId]; ok {
		return val
	}
	return nil
}

func DeleteContainerById(containerId string) {
	delete(containersWsConnectionsMap, containerId)
	delete(containersTtlMap, containerId)
}
