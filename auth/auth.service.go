package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/sjwt"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

/**
 * Regiser oAuth Providers
 */
func RegisterOAuth() {

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

/**
 * Get called after successfull sign in on provider
 */
func handleOAuthCallback(w http.ResponseWriter, req *http.Request) {
	// extract user from request
	gothUser, err := gothic.CompleteUserAuth(w, req)

	fmt.Printf("Goth User %v\n", gothUser)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// register or update existing user in db
	user := UserSignedIn(gothUser)

	if user == nil {
		fmt.Fprintln(w, errors.New("USER NOT FOUND"))
		return
	}

	// generate jwt
	claims, _ := sjwt.ToClaims(user)
	claims.SetExpiresAt(time.Now().Add(time.Hour * 24))
	jwt := claims.Generate([]byte(os.Getenv("JWT_SECRET")))

	t, err := template.ParseFiles("auth.html")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// parse jwt to auth.html
	t.Execute(w, JwtToken{jwt})

}

/**
 * This function is used for every endpoint, where its required
 * for the user be signed in (e.g. Edit-Post, Create-Post, Delete-Post).
 *
 * This function acts as a middleware and calls the handlerFunction after
 * a successfull authentisation
 */
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// Extract Auth Header and check if it exists
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			unauthorized(w, "No authorization header set")
			return
		}

		// Check if bearer token is set inside auth header
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			unauthorized(w, "No bearer token set")
			return
		}

		// extract jwt from bearer and verify signature
		jwt := bearerToken[1]
		verified := sjwt.Verify(jwt, []byte(os.Getenv("JWT_SECRET")))

		if !verified {
			unauthorized(w, "Invalid JWT")
			return
		}

		// extract jwt data
		claims, err := sjwt.Parse(jwt)
		if err == nil {
			err = claims.Validate()
		}

		if err != nil {
			unauthorized(w, err.Error())
			return
		}

		// map jwt data to User struct
		var user User
		claims.ToStruct(&user)

		// set user in context to extract later
		// used to store authorId within message object
		context.Set(req, "user", user)

		if next != nil {
			next(w, req)
		}
	})
}

/**
 * This function sends an unauthorized exception to the requester
 */
func unauthorized(w http.ResponseWriter, reason string) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(Exception{reason})
}
