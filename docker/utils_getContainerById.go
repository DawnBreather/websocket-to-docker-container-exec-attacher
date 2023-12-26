package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func GetContainerById(ctx context.Context, dockerClient *DockerClient, containerID string) (types.Container, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("id", containerID)

	containers, err := dockerClient.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return types.Container{}, err
	}

	if len(containers) == 0 {
		return types.Container{}, fmt.Errorf("no container found with ID %s", containerID)
	}

	// Assuming the first match is the desired container since IDs are unique
	return containers[0], nil
}
