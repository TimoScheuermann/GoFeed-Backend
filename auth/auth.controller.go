package auth

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
)

/**
 * Registers Routes needed for authentication
 */
func RegisterRoutes(router *mux.Router) {
	fmt.Println("Auth routes registered")

	// Create Subrouter
	s := router.PathPrefix("/auth").Subrouter()

	s.HandleFunc("/{provider}", gothic.BeginAuthHandler).Methods("GET")
	s.HandleFunc("/{provider}/callback", handleOAuthCallback).Methods("GET")
	// Use middleware to authenticate user
	s.HandleFunc("/valid", AuthMiddleware(nil)).Methods("POST")

	// Special Route for UserInformation
	router.HandleFunc("/user/{id}", GetUserInfo).Methods("GET")
}
