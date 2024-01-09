package httphandlers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func StartWebServer() {
	r := mux.NewRouter()

	defineRoutes(r) // call the new function for defining routes

	log.Println("WebSocket server listening on :4242")
	log.Fatal(http.ListenAndServe(":4242", r))
}
