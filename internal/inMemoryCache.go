package internal

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type InMemoryCache struct {
	Games     map[string]*Game
	UserGames map[string]string
	sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		Games:     make(map[string]*Game),
		UserGames: make(map[string]string),
	}
}

func (i *InMemoryCache) cleanupLoop(ctx context.Context, cleanupInterval, ttl time.Duration) {

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			i.Lock()
			now := time.Now()
			for id, game := range i.Games {
				if now.Sub(game.UpdatedAt) > ttl {
					delete(i.Games, id)
					if game.PlayerX != "" {
						delete(i.UserGames, game.PlayerX)
					}
					if game.PlayerO != "" {
						delete(i.UserGames, game.PlayerO)
					}
					log.Println("Удалена неактивная игра:", id)
				}
			}
			i.Unlock()

		case <-ctx.Done():
			return
		}
	}
}

func (i *InMemoryCache) CreateGame(ctx context.Context, userID string) OutgoingMessage {
	i.Lock()
	defer i.Unlock()

	gameID := strconv.Itoa(rand.Intn(10000))

	i.Games[gameID] = &Game{
		ID:        gameID,
		PlayerX:   userID,
		Turn:      userID,
		UpdatedAt: time.Now(),
	}
	i.UserGames[userID] = gameID

	return OutgoingMessage{
		UserID: userID,
		Text:   fmt.Sprintf("Игра %s создана. Ожидаем второго игрока", gameID),
	}
}

func (i *InMemoryCache) ListGames(ctx context.Context, userID string) OutgoingMessage {
	i.Lock()
	defer i.Unlock()

	var buttons []Button

	for id, game := range i.Games {
		if game.PlayerO == "" && !game.Finished {
			buttons = append(buttons, Button{
				Text:   fmt.Sprintf("Присоединиться к %s", game.ID),
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

func (i *InMemoryCache) JoinGame(ctx context.Context, userID, gameID string) OutgoingMessage {

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
	if rand.Intn(2) == 0 {
		game.Turn = game.PlayerX
	} else {
		game.Turn = game.PlayerO
	}
	game.UpdatedAt = time.Now()
	i.UserGames[userID] = gameID

	text := "Игра началась! Ходит: " + game.Turn
	return renderBoard(game, text)
}

func (i *InMemoryCache) Move(ctx context.Context, userID, coord string) OutgoingMessage {

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
	game.UpdatedAt = time.Now()

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
