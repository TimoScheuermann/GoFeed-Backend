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

const sessionKey = "superDuperSicheresPW"
const jwtSecret = "JWTsSindTollUndS1ch3r!"

func RegisterOAuth() {

	fmt.Println("Register OAuth")

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = strings.HasPrefix(os.Getenv("CALLBACK"), "https")

	gothic.Store = store

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), os.ExpandEnv("${CALLBACK}/auth/google/callback"), "profile"),
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), os.ExpandEnv("${CALLBACK}/auth/github/callback"), "user:name"),
	)
}

func handleOAuthCallback(w http.ResponseWriter, req *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, req)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	user := UserSignedIn(gothUser)

	if user == nil {
		fmt.Fprintln(w, errors.New("USER NOT FOUND"))
		return
	}

	claims, _ := sjwt.ToClaims(user)
	claims.SetExpiresAt(time.Now().Add(time.Hour * 24))
	jwt := claims.Generate([]byte(jwtSecret))

	t, _ := template.ParseFiles("auth.html")
	t.Execute(w, JwtToken{jwt})

}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func AuthMiddleware(next http.HandlerFunc, groups []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			bearerToken := strings.Split(authHeader, " ")

			if len(bearerToken) == 2 {
				jwt := bearerToken[1]
				if verified := sjwt.Verify(jwt, []byte(jwtSecret)); verified {
					claims, err := sjwt.Parse(jwt)

					if err == nil {
						err = claims.Validate()
					}

					if err != nil {
						w.WriteHeader(http.StatusUnauthorized)
						w.Header().Add("content-type", "application/json")
						json.NewEncoder(w).Encode(Exception{Message: err.Error()})
						return
					}

					var user User
					claims.ToStruct(&user)

					if groups != nil && !contains(groups, user.Group) {
						w.WriteHeader(http.StatusUnauthorized)
						w.Header().Add("content-type", "application/json")
						json.NewEncoder(w).Encode(Exception{Message: "Not allowed"})
						return
					}

					context.Set(req, "user", user)

					if next != nil {
						next(w, req)
					}
					return
				}
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(Exception{Message: "An authorization header is required"})
	})
}
