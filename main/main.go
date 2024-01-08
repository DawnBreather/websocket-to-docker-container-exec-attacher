package main

import (
	"context"
	"log"
	"quic_shell_server/docker"
	"quic_shell_server/httphandlers"
	"time"
)

func main() {
  docker.InitializeDockerClient()

  // Start expired Docker containers cleanup job
  go func() {
    for {
      time.Sleep(6 * time.Second)
      err := docker.StopAndDeleteContainersWithTtlExpired(context.Background(), docker.Client)
      if err != nil {
        log.Printf("Failed removing expired Docker containers: %v", err)
      }
    }
  }()

  httphandlers.StartWebServer()
}
