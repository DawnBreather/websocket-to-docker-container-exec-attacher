package docker

var (
	defaultDockerContainersLabels = map[string]string{
		"by": "LMS_DOCKER_CONTAINERS_SPAWNER",
	}
	defaultDockerContainersTtlInMinutes = 180
	//DefaultDockerImage                  = "docker.io/library/debian:latest"
	DefaultDockerImage = "debian:latest"
)
