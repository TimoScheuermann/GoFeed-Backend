package message

import (
	"fmt"
	"gofeed-go/auth"

	"github.com/gorilla/mux"
)

/**
 * Registers Routes needed for authentication
 */
func RegisterRoutes(router *mux.Router) {
	fmt.Println("Message routes registered")

	// Create Subrouter
	s := router.PathPrefix("/message").Subrouter()

	s.HandleFunc("", getMessages).Methods("GET")
	s.HandleFunc("/{id}", getMessage).Methods("GET")

	// Use middleware to authenticate user
	s.HandleFunc("", auth.AuthMiddleware(postMessage)).Methods("POST")
	s.HandleFunc("/{id}", auth.AuthMiddleware(deleteMessage)).Methods("DELETE")
	s.HandleFunc("/{id}", auth.AuthMiddleware(patchMessage)).Methods("PATCH")
}
