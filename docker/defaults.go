package docker

var (
	defaultDockerContainersLabels = map[string]string{
		"by": "LMS_DOCKER_CONTAINERS_SPAWNER",
	}
	defaultDockerContainersTtlInMinutes = 180
)
