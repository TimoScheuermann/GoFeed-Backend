package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"net/http"

	"gofeed-go/persistence"
	"gofeed-go/service"
	"gofeed-go/transport"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// Connect to Database
	db := conntectToDB()

	// Create Router
	router := mux.NewRouter()
	router.Use(routerMw)
	router.StrictSlash(true)

	// Auth Module
	as := service.NewAuthService()

	// User Module
	ur := persistence.NewUserPersistor(db.Collection("user"))
	us := service.NewUserService(ur)
	ut := transport.NewUserController(us, as)
	ut.RegisterRoutes(router)

	// Message Module
	mr := persistence.NewMessagePersistor(db.Collection("message"))
	ms := service.NewMessageService(mr)
	mt := transport.NewMessageController(ms, as)
	mt.RegisterRoutes(router)

	// Enable CORs
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Origin"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

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

func conntectToDB() *mongo.Database {

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	c, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB")

	return c.Database("gofeed-go")
}
