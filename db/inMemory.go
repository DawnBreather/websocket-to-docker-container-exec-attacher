package db

import (
	"github.com/gorilla/websocket"
	"time"
)

// TODO Implement Redis of Postgres instead of in-memory storage

var containerStartDateTimeMap = map[string]time.Time{}
var containersTtlMap = map[string]int{}
var containersWsConnectionsMap = map[string][]*websocket.Conn{}

func GetTtlInSecondsByContainerId(containerId string) (seconds int) {
	if val, ok := containersTtlMap[containerId]; ok {
		return val
	}
	return 0
}

func SetStartdatetimeInSecondsByContainerId(containerId string) {
	containerStartDateTimeMap[containerId] = time.Now()
}

func GetStartdatetimeInSecondsByContainerId(containerId string) int {
	return int(time.Now().Sub(containerStartDateTimeMap[containerId]).Seconds())
}

func SetTtlInSecondsByContainerId(containerId string, ttlSeconds int) {
	containersTtlMap[containerId] = ttlSeconds
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
	delete(containerStartDateTimeMap, containerId)
}
