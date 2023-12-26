package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	. "github.com/docker/docker/api/types/container"
	"log"
	"quic_shell_server/db"
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
		} else if minutes > db.GetTtlInMinutesByContainerId(container.ID) {
			timeout := 3
			err = client.cli.ContainerStop(ctx, container.ID, StopOptions{
				Signal:  "9",
				Timeout: &timeout,
			})
			if err != nil {
				log.Printf("Failed to stop container { %s }", container.ID)
			}

			err = client.cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
				RemoveVolumes: true,
				RemoveLinks:   true,
				Force:         true,
			})

			if err != nil {
				log.Printf("Failed to remove container { %s }", container.ID)
			}

			for _, conn := range db.GetWsConnectionByContainerId(container.ID) {
				conn.Close()
			}

			db.DeleteContainerById(container.ID)
		}
	}

	return nil
}
