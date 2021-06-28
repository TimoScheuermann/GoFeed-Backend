package message

import (
	"fmt"
	"gofeed-go/auth"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	fmt.Println("Message routes registered")
	s := router.PathPrefix("/message").Subrouter()

	s.HandleFunc("", getMessages).Methods("GET")
	s.HandleFunc("/{id}", getMessage).Methods("GET")

	s.HandleFunc("", auth.AuthMiddleware(postMessage, nil)).Methods("POST")
	s.HandleFunc("/{id}", auth.AuthMiddleware(deleteMessage, nil)).Methods("DELETE")
	s.HandleFunc("/{id}", auth.AuthMiddleware(patchMessage, nil)).Methods("PATCH")
}
