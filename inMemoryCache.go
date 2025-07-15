package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

type InMemoryCache struct {
	Games map[string]*Game
	UserGames map[string]string
	sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		Games: make(map[string]*Game),
		UserGames: make(map[string]string),
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
	i.UserGames[userID] = gameID

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
	i.UserGames[userID] = gameID

	text := "Игра началась! Ходит: " + game.Turn
	return renderBoard(game, text)
}

func (i *InMemoryCache) Move(userID, coord string) OutgoingMessage {

	i.Lock()
	defer i.Unlock()

	gameID, ok := i.UserGames[userID]
	if !ok {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Вы не в игре",
		}
	}

	game := i.Games[gameID]
	if game.Finished || game.Turn != userID {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Сейчас не ваш ход",
		}
	}

	x, y := parseCoord(coord)
	if x == -1 || game.Board[x][y] != "" {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Недопустимый ход",
		}
	}

	mark := "X"
	if userID == game.PlayerO {
		mark = "O"
	}
	game.Board[x][y] = mark

	if winner := checkWinner(game.Board); winner != "" {
		game.Finished = true
		game.Winner = userID
		return renderBoard(game, fmt.Sprintf("Победа: %s", mark))
	}

	if isDraw(game.Board) {
		game.Finished = true
		return renderBoard(game, "Ничья!")
	}

	if userID == game.PlayerX {
		game.Turn = game.PlayerO
	} else {
		game.Turn = game.PlayerX
	}

	return renderBoard(game, "Ход противника")
}
