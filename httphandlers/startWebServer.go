package httphandlers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Define HTTP routes in their own function for better readability
func defineRoutes(r *mux.Router) {
	r.HandleFunc("/ws", webSocketHandler)
	r.HandleFunc("/api/admin/container/{id}/extend_ttl/{minutes}", apiAdminContainerExtendTtlHandler).Methods("POST")
	r.HandleFunc("/api/admin/container/{id}", apiAdminContainerHandler).Methods("GET")
}

func StartWebServer() {
	r := mux.NewRouter()

	defineRoutes(r) // call the new function for defining routes

	log.Println("WebSocket server listening on :4242")
	log.Fatal(http.ListenAndServe(":4242", r))
}
