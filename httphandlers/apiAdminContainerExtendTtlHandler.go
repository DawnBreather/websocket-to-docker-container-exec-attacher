package httphandlers

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"quic_shell_server/db"
	"quic_shell_server/docker"
	"quic_shell_server/utils"
	"strconv"
)

func apiAdminContainerExtendTtlHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	id := vars["id"]
	container, err := docker.GetContainerById(ctx, docker.Client, id)
	if err != nil {
		writeResponse(w, http.StatusNotFound, err.Error())
	}
	containerRunningTime, err := docker.GetContainerRunningTime(ctx, docker.Client, id)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, err.Error())
	}
	seconds, err := strconv.Atoi(vars["seconds"])
	if err != nil {
		log.Printf("Failed to parse seconds { %s } into type of Int: %v", vars["seconds"], err)
	} else {
		db.SetTtlInSecondsByContainerId(id, seconds+db.GetTtlInSecondsByContainerId(id))
	}
	fmt.Fprintf(w, string(utils.MustMarshal(containerIDMessage{
		ContainerID:  id,
		Image:        container.Image,
		TTL:          db.GetTtlInSecondsByContainerId(id),
		CMD:          container.Command,
		RemainingTTL: db.GetTtlInSecondsByContainerId(id) - containerRunningTime*60,
	})))
}

func apiAdminContainerStopAndRemoveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	id := vars["id"]

	err := docker.StopContainerAndDelete(ctx, docker.Client, id)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, err.Error())
	}
	writeResponse(w, http.StatusOK, "{}")
}
