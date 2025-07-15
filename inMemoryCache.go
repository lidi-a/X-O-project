package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

type InMemoryCache struct {
	Games map[string]*Game
	sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		Games: make(map[string]*Game),
	}
}

func (i *InMemoryCache) CreateGame(userID string) OutgoingMessage {
	i.Lock()
	defer i.Unlock()

	gameID := strconv.Itoa(rand.Intn(10000))

	i.Games[gameID] = &Game{
		ID:      gameID,
		PlayerX: userID,
		Turn:    userID,
	}

	return OutgoingMessage{
		UserID: userID,
		Text:   fmt.Sprintf("Игра %s создана. Ожидаем второго игрока", gameID),
	}
}

func (i *InMemoryCache) ListGames(userID string) OutgoingMessage {
	i.Lock()
	defer i.Unlock()

	var buttons []Button

	for id, game := range i.Games {
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

func (i *InMemoryCache) JoinGame(userID, gameID string) OutgoingMessage {

	i.Lock()
	defer i.Unlock()

	game, ok := i.Games[gameID]
	if !ok || game.PlayerO != "" {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Игра недоступна",
		}
	}

	game.PlayerO = userID

	text := "Игра началась! Ходит: " + game.Turn
	return renderBoard(game, text)
}
