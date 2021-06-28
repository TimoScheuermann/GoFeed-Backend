package main

import (
	"fmt"
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

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		mongo := os.Getenv("MONGO_URI")
		if len(mongo) < 5 {
			log.Fatal("No env set")
		} else {
			fmt.Println("Env loaded via docker-compose")
			fmt.Printf("Mongo URI: %v\n", mongo)
		}
	}
}

func main() {

	auth.RegisterOAuth()
	database.Connect()

	router := mux.NewRouter()
	router.Use(routerMw)
	router.StrictSlash(true)

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Origin"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	auth.RegisterRoutes(router)
	message.RegisterRoutes(router)

	server := &http.Server{
		Addr:         ":443",
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	cert := "/etc/letsencrypt/live/api.timos.design/fullchain.pem"
	key := "/etc/letsencrypt/live/api.timos.design/privkey.pem"

	err := server.ListenAndServeTLS(cert, key)

	if err == nil {
		fmt.Println("Serving over https")
	}

	if err != nil {
		server.Addr = ":80"
		fmt.Printf("Serving over http due to %v\n", err)
		defer log.Fatal(server.ListenAndServe())
	}
}

func routerMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/callback") {
			w.Header().Set("content-type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}
