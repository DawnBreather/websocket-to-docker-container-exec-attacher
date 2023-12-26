package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

// GetContainerRunningTime returns the running time of a container in minutes.
func listAllRunningContainers(ctx context.Context, client *DockerClient, labels map[string]string) (containers []types.Container, err error) {

	containers, err = client.cli.ContainerList(ctx, types.ContainerListOptions{
		Size:   false,
		All:    false,
		Latest: false,
		Since:  "",
		Before: "",
		Limit:  0,
		Filters: func(labels map[string]string) (labelFilter filters.Args) {
			labelFilter = filters.NewArgs()
			for key, value := range labels {
				labelFilter.Add("label", key+"="+value)
			}
			return
		}(labels),
	})

	return containers, nil

}
