package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/brianvoe/sjwt"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

type Exception struct {
	Message string `json:"message"`
}

type User struct {
	UserID      primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ProviderID  string             `json:"providerId" bson:"providerId"`
	Provider    string             `json:"provider" bson:"provider"`
	Name        string             `json:"name" bson:"name"`
	Avatar      string             `json:"avatar" bson:"avatar"`
	Group       string             `json:"group" bson:"group"`
	MemberSince int64              `json:"member_since" bson:"member_since"`
	LastLogin   int64              `json:"last_login" bson:"last_login"`
}

type userKey struct{}

func init() {

	godotenv.Load(".env")

	if len(os.Getenv("JWT_SECRET")) < 10 {
		log.Fatal("No env set")
	}

	fmt.Println("Register OAuth")

	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = strings.HasPrefix(os.Getenv("CALLBACK"), "https")

	gothic.Store = store

	// Providers used for GoFeed
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), os.ExpandEnv("${CALLBACK}/auth/google/callback"), "profile"),
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), os.ExpandEnv("${CALLBACK}/auth/github/callback"), "user:name"),
	)
}

func (a *AuthService) ExtractUser(req *http.Request) (*User, error) {
	var user User
	err := mapstructure.Decode(req.Context().Value(&userKey{}), &user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthService) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			unauthorized(w, "No authorization header set")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			unauthorized(w, "No bearer token set")
			return
		}

		jwt := bearerToken[1]
		verified := sjwt.Verify(jwt, []byte(os.Getenv("JWT_SECRET")))

		if !verified {
			unauthorized(w, "Invalid JWT")
			return
		}

		claims, err := sjwt.Parse(jwt)
		if err == nil {
			err = claims.Validate()
		}

		if err != nil {
			unauthorized(w, err.Error())
			return
		}

		var user User
		claims.ToStruct(&user)

		if next != nil {
			ctx := context.WithValue(req.Context(), &userKey{}, user)
			next(w, req.WithContext(ctx))
		}
	})
}

func unauthorized(w http.ResponseWriter, reason string) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(Exception{reason})
}
