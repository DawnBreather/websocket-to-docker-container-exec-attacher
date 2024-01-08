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
	minutes, err := strconv.Atoi(vars["minutes"])
	if err != nil {
		log.Printf("Failed to parse minutes { %s } into type of Int: %v", vars["minutes"], err)
	} else {
		db.SetTtlInMinutesByContainerId(id, minutes+db.GetTtlInMinutesByContainerId(id))
	}
	fmt.Fprintf(w, string(utils.MustMarshal(containerIDMessage{
		ContainerID:  id,
		Image:        container.Image,
		TTL:          db.GetTtlInMinutesByContainerId(id),
		CMD:          container.Command,
		RemainingTTL: db.GetTtlInMinutesByContainerId(id) - containerRunningTime,
	})))
}
