package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	cache CacheProvider
}

type CacheProvider interface {
	CreateGame(userID string) OutgoingMessage
	ListGames(userID string) OutgoingMessage
	JoinGame(userID, gameID string) OutgoingMessage
}

func NewHandler(cacheProvider CacheProvider) (*Handler, error) {
	if cacheProvider == nil {
		return nil, errors.New("missing cache provider")
	}
	h := &Handler{
		cache: cacheProvider,
	}
	return h, nil
}

func (h *Handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	var msg IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}

	var response OutgoingMessage
	if msg.Text != nil {
		switch *msg.Text {
		case "/new":
			response = h.cache.CreateGame(msg.UserID)
		case "/list":
			response = h.cache.ListGames(msg.UserID)
		default:
			response = OutgoingMessage{
				UserID: msg.UserID,
				Text:   "Неизвестная команда",
			}
		}

	} else if msg.Action != nil {
		action := *msg.Action
		if strings.HasPrefix(action, "Join:") {
			gameID := strings.TrimPrefix(action, "Join:")
			response = h.cache.JoinGame(msg.UserID, gameID)
		} else {
			// TODO
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
