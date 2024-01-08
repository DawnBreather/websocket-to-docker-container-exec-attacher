package httphandlers

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"quic_shell_server/db"
	"quic_shell_server/docker"
	"quic_shell_server/utils"
)

func apiAdminContainerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	container, err := docker.GetContainerById(context.Background(), docker.Client, id)
	if err != nil {
		writeResponse(w, http.StatusNotFound, err.Error())
	}

	containerRunningTime, err := docker.GetContainerRunningTime(context.Background(), docker.Client, id)
	if err != nil {
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, err.Error())
		}
	}

	fmt.Fprintf(w, string(utils.MustMarshal(containerIDMessage{
		ContainerID:  id,
		Image:        container.Image,
		TTL:          db.GetTtlInMinutesByContainerId(id),
		CMD:          container.Command,
		RemainingTTL: db.GetTtlInMinutesByContainerId(id) - containerRunningTime,
	})))
}
