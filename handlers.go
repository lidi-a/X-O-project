package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	cache CacheProvider
}

type CacheProvider interface {
	CreateGame(ctx context.Context, userID string) OutgoingMessage
	ListGames(ctx context.Context, userID string) OutgoingMessage
	JoinGame(ctx context.Context, userID, gameID string) OutgoingMessage
	Move(ctx context.Context, userID, coord string) OutgoingMessage
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
			response = h.cache.CreateGame(r.Context(), msg.UserID)
		case "/list":
			response = h.cache.ListGames(r.Context(), msg.UserID)
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
			response = h.cache.JoinGame(r.Context(), msg.UserID, gameID)
		} else if strings.HasPrefix(action, "Move:") {
			coord := strings.TrimPrefix(action, "Move:")
			response = h.cache.Move(r.Context(), msg.UserID, coord)
		} else {
			response = OutgoingMessage{
				UserID: msg.UserID,
				Text:   "Неизвестное действие",
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
