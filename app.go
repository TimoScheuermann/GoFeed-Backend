package main

import (
	"log"
	"os"
	"strings"
	"time"

	"net/http"

	"gofeed-go/auth"
	"gofeed-go/database"
	"gofeed-go/message"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

// load env variables
func init() {

	// load .env file (if exists)
	godotenv.Load(".env")

	// its possible, that env vars might be set by Docker (docker-compose.yml)
	// => no .env need, therefor check if JWT_SECRET exists
	if len(os.Getenv("JWT_SECRET")) < 10 {

		// .env doesnt exists and docker hasnt specified any .env => abort
		log.Fatal("No env set")
	}
}

/**
 * The Main Function.
 */
func main() {

	// Register oAuth Providers
	auth.RegisterOAuth()

	// Connect to Database
	database.Connect()

	// Create Router
	router := mux.NewRouter()
	router.Use(routerMw)
	router.StrictSlash(true)

	// Enable CORs
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Origin"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	// Register Routes
	auth.RegisterRoutes(router)
	message.RegisterRoutes(router)

	// Configure server
	server := &http.Server{
		Addr:         os.ExpandEnv("${host}:3000"),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start Server
	defer log.Fatal(server.ListenAndServe())
}

/**
 * This middleware is used to set the content-type to application/json
 * for every callback, except auth callback.
 * Auth Callback uses a .html file
 */
func routerMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/callback") {
			w.Header().Set("content-type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}
