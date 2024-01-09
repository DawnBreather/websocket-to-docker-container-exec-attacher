package httphandlers

import "github.com/gorilla/mux"

// Define HTTP routes in their own function for better readability
func defineRoutes(r *mux.Router) {
	r.HandleFunc("/ws", webSocketHandler)
	r.HandleFunc("/api/admin/container/{id}/extend_ttl/{minutes}", apiAdminContainerExtendTtlHandler).Methods("POST")
	r.HandleFunc("/api/admin/container/{id}", apiAdminContainerHandler).Methods("GET")
	r.HandleFunc("/api/aws/playground", awsPlaygroundHandler).Methods("GET")
}
