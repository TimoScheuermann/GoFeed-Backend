package transport

import (
	"encoding/json"
	"fmt"
	"gofeed-go/service"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/brianvoe/sjwt"
	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
)

type UserController struct {
	s *service.UserService
	a *service.AuthService
}

type JwtToken struct {
	Token string `json:"token"`
}

func NewUserController(s *service.UserService, a *service.AuthService) *UserController {
	return &UserController{s, a}
}

func (c *UserController) RegisterRoutes(router *mux.Router) {

	// Special Route for UserInformation
	router.HandleFunc("/user/{id}", c.getUserInfo).Methods("GET")

	// Use middleware to authenticate user
	router.HandleFunc("/auth/valid", c.a.Middleware(nil)).Methods("POST")

	router.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler).Methods("GET")
	router.HandleFunc("/auth/{provider}/callback", c.handleOAuthCallback).Methods("GET")

	fmt.Println("User routes registered")
}

func (c *UserController) getUserInfo(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]

	if !ok {
		http.Error(w, "Missing param: id", http.StatusBadRequest)
		return
	}

	userInfo, err := c.s.GetUserInfo(req.Context(), id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *UserController) handleOAuthCallback(w http.ResponseWriter, req *http.Request) {
	// extract user from request
	gothUser, err := gothic.CompleteUserAuth(w, req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// register or update existing user in db
	user, err := c.s.UserSignedIn(req.Context(), gothUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// generate jwt
	claims, err := sjwt.ToClaims(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	claims.SetExpiresAt(time.Now().Add(time.Hour * 24))
	jwt := claims.Generate([]byte(os.Getenv("JWT_SECRET")))

	t, err := template.ParseFiles("auth.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// parse jwt to auth.html
	t.Execute(w, JwtToken{jwt})
}
