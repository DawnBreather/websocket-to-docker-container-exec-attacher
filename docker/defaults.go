package docker

import (
	"gitlab.com/avarf/getenvs"
)

var (
	defaultDockerContainersLabels = map[string]string{
		"by": "LMS_DOCKER_CONTAINERS_SPAWNER",
	}
	defaultDockerContainersTtlInMinutes = 180
	//DefaultDockerImage                  = "docker.io/library/debian:latest"
	DefaultDockerImage              = getenvs.GetEnvString("DEFAULT_DOCKER_IMAGE", "debian:latest")
	DefaultDockerContainerShell     = getenvs.GetEnvString("DEFAULT_DOCKER_CONTAINER_TERMINAL", "/bin/bash")
	DefaultUserForExecIntoContainer = getenvs.GetEnvString("DEFAULT_DOCKER_CONTAINER_USER", "")
)
