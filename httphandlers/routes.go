package httphandlers

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Define HTTP routes in their own function for better readability
func defineRoutes(r *mux.Router) {
	//http.Handle("/", http.FileServer(http.Dir("./front/")))
	r.PathPrefix("/webterminal").Handler(http.StripPrefix("/webterminal", http.FileServer(http.Dir("./front/"))))
	r.HandleFunc("/ws", webSocketHandler)
	r.HandleFunc("/api/admin/container/{id}/extend_ttl/{seconds}", apiAdminContainerExtendTtlHandler).Methods("POST")
	r.HandleFunc("/api/admin/container/{id}/stop_and_remove", apiAdminContainerStopAndRemoveHandler).Methods("PUT")
	r.HandleFunc("/api/admin/container/{id}", apiAdminContainerHandler).Methods("GET")
	r.HandleFunc("/api/aws/playground", awsPlaygroundHandler).Methods("GET")
}
