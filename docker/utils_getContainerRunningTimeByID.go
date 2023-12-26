package docker

import (
	"context"
	"fmt"
	"time"
)

// GetContainerRunningTime returns the running time of a container in minutes.
func GetContainerRunningTime(ctx context.Context, client *DockerClient, containerID string) (minutes int, err error) {

	// Get container details
	containerJSON, err := client.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, fmt.Errorf("error inspecting container: %s", err)
	}

	// Parse the creation time
	creationTime, err := time.Parse(time.RFC3339Nano, containerJSON.Created)
	if err != nil {
		return 0, fmt.Errorf("error parsing creation time: %s", err)
	}

	// Calculate the difference in minutes
	duration := time.Since(creationTime)
	return int(duration.Minutes()), nil
}
