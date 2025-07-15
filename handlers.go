package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

var (
	games = make(map[string]*Game)
	mu    sync.RWMutex
)

func handleMessage(w http.ResponseWriter, r *http.Request) {
	var msg IncomingMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}

	var response OutgoingMessage
	if msg.Text != nil {
		switch *msg.Text {
		case "/new":
			response = handleNewGame(msg.UserID)
		case "/list":
			response = handleListGames(msg.UserID)
		default:
			response = OutgoingMessage{
				UserID: msg.UserID,
				Text: "Неизвестная команда",
			}
		}


	} else {
		// TODO
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNewGame(userID string) OutgoingMessage {
	mu.Lock()
	defer mu.Unlock()

	gameID := strconv.Itoa(rand.Intn(10000))

	games[gameID] = &Game{
		ID:      gameID,
		PlayerX: userID,
		Turn:    userID,
	}

	return OutgoingMessage{
		UserID: userID,
		Text:   fmt.Sprintf("Игра %s создана. Ожидаем второго игрока", gameID),
	}
}

func handleListGames(userID string) OutgoingMessage {
	mu.Lock()
	defer mu.Unlock()

	var buttons []Button

	for id, game := range games {
		if game.PlayerO == "" && !game.Finished {
			buttons = append(buttons, Button{
				Text:   "Присоединиться к игре",
				Action: "Join" + id,
			})
		}
	}

	if len(buttons) == 0 {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Нет доступных игр",
		}
	}

	return OutgoingMessage{
		UserID:  userID,
		Text:    "Выберете игру",
		Buttons: buttons,
	}
}


