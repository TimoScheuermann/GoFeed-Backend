package transport

import (
	"encoding/json"
	"fmt"
	"gofeed-go/persistence"
	"gofeed-go/service"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type MessageController struct {
	s *service.MessageService
	a *service.AuthService
}

func NewMessageController(s *service.MessageService, a *service.AuthService) *MessageController {
	return &MessageController{s, a}
}

func (c *MessageController) RegisterRoutes(router *mux.Router) {

	router.HandleFunc("/message", c.getMessages).Methods("GET")
	router.HandleFunc("/message/{id}", c.getMessage).Methods("GET")

	// Use middleware to authenticate user
	router.HandleFunc("/message", c.a.Middleware(c.postMessage)).Methods("POST")
	router.HandleFunc("/message/{id}", c.a.Middleware(c.deleteMessage)).Methods("DELETE")
	router.HandleFunc("/message/{id}", c.a.Middleware(c.patchMessage)).Methods("PATCH")

	fmt.Println("Message routes registered")
}

func (c *MessageController) getMessages(w http.ResponseWriter, req *http.Request) {

	query := req.URL.Query()
	var (
		limit *int64
		skip  *int64
	)

	if l, err := strconv.ParseInt(query.Get("limit"), 10, 64); err == nil {
		limit = &l
	}
	if s, err := strconv.ParseInt(query.Get("skip"), 10, 64); err == nil {
		skip = &s
	}

	messages, err := c.s.GetMessages(req.Context(), limit, skip)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *MessageController) getMessage(w http.ResponseWriter, req *http.Request) {

	id, ok := mux.Vars(req)["id"]

	if !ok {
		http.Error(w, "Missing param: id", http.StatusBadRequest)
		return
	}

	message, err := c.s.GetMessageById(req.Context(), id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		// if errors.Is(err, service.ErrInvalidObjectID) { }
		return
	}

	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *MessageController) postMessage(w http.ResponseWriter, req *http.Request) {

	var body persistence.Message
	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := c.a.ExtractUser(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	message, err := c.s.CreateMessage(req.Context(), persistence.Message{AuthorID: user.UserID, Content: body.Content})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *MessageController) deleteMessage(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]

	if !ok {
		http.Error(w, "Missing param: id", http.StatusBadRequest)
		return
	}

	user, err := c.a.ExtractUser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	success, err := c.s.DeleteMessage(req.Context(), id, user.UserID.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if !success {
		http.Error(w, "Couldn't delete message", http.StatusInternalServerError)
		return
	}
}

func (c *MessageController) patchMessage(w http.ResponseWriter, req *http.Request) {

	id, ok := mux.Vars(req)["id"]

	if !ok {
		http.Error(w, "Missing param: id", http.StatusBadRequest)
		return
	}

	var body persistence.Message
	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := c.a.ExtractUser(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	message, err := c.s.UpdateMessage(req.Context(), id, user.UserID.Hex(), persistence.Message{Content: body.Content})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
