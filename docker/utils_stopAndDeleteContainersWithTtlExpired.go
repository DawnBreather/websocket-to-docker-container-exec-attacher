package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	. "github.com/docker/docker/api/types/container"
	"log"
	"quic_shell_server/db"
	"strings"
	"time"
)

func StopAndDeleteContainersWithTtlExpired(ctx context.Context, client *DockerClient) error {
	containers, err := listAllRunningContainers(ctx, client, defaultDockerContainersLabels)
	if err != nil {
		return fmt.Errorf("failed listing all running containers: %v", err)
	}

	for _, container := range containers {
		minutes, err := GetContainerRunningTime(ctx, client, container.ID)
		if err != nil {
			log.Printf("Failed getting container running time for container_id { %s }", container.ID)
			continue
		} else if minutes*60 > db.GetTtlInSecondsByContainerId(container.ID) {
			err = StopContainerAndDelete(ctx, client, container.ID)
			if err != nil {
				log.Printf(err.Error())
			}
		}
	}

	return nil
}

func StopContainerAndDelete(ctx context.Context, client *DockerClient, containerId string) error {
	var err error
	var timeout = 1
	err = client.cli.ContainerStop(ctx, containerId, StopOptions{
		Signal:  "9",
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container { %s }: %s", containerId, err)
	}

	for {
		time.Sleep(1 * time.Second)
		json, err := client.cli.ContainerInspect(ctx, containerId)
		if err != nil {
			break
		}
		if json.State.Status == "exited" {
			break
		}
	}

	err = client.cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
		Force:         true,
	})

	if err != nil && !strings.Contains(err.Error(), "No such container") {
		return fmt.Errorf("failed to remove container { %s }: %s", containerId, err)
	}

	for _, conn := range db.GetWsConnectionByContainerId(containerId) {
		conn.Close()
	}

	db.DeleteContainerById(containerId)
	return nil
}
