package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client  *redis.Client
	ttl     time.Duration
	locker  *redislock.Client
	ttlLock time.Duration
}

func NewRedisCache(addr, password string, db int, ttl, ttlLock time.Duration) *RedisCache {

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
		ttl:    ttl,
		locker: redislock.New(client),
		ttlLock: ttlLock,
	}
}

func (r *RedisCache) CreateGame(ctx context.Context, userID string) OutgoingMessage {

	gameID := strconv.Itoa(rand.Intn(10000))
	game := Game{
		ID:      gameID,
		PlayerX: userID,
		Turn:    userID,
	}

	data, _ := json.Marshal(game)
	r.client.Set(ctx, "game:"+game.ID, data, r.ttl)
	r.client.Set(ctx, "usergame:"+userID, gameID, r.ttl)

	return OutgoingMessage{
		UserID: userID,
		Text:   fmt.Sprintf("Игра %s создана. Ожидаем второго игрока", gameID),
	}
}

func (r *RedisCache) ListGames(ctx context.Context, userID string) OutgoingMessage {

	keys, _ := r.client.Keys(ctx, "game:*").Result()
	var buttons []Button
	for _, key := range keys {
		data, _ := r.client.Get(ctx, key).Result()
		var game Game
		if err := json.Unmarshal([]byte(data), &game); err == nil {
			if game.PlayerO == "" && !game.Finished {
				buttons = append(buttons, Button{
					Text:   fmt.Sprintf("Присоединиться к %s", game.ID),
					Action: "Join:" + game.ID,
				})
			}
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
		Text:    "Выберите игру:",
		Buttons: buttons,
	}
}

func (r *RedisCache) JoinGame(ctx context.Context, userID, gameID string) OutgoingMessage {

	// Пробуем установить блокировку
	lock, err := lockGame(ctx, r.locker, gameID, r.ttlLock)
	if err != nil || err == redislock.ErrNotObtained {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Игра сейчас занята, попробуйте позже",
		}
	}
	// Откладываем ручной релиз, если функция завершится раньше
	defer lock.Release(ctx)

	data, _ := r.client.Get(ctx, "game:"+gameID).Result()
	var game Game
	_ = json.Unmarshal([]byte(data), &game)

	if game.ID == "" || game.PlayerO != "" {
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

	dataBytes, _ := json.Marshal(game)
	r.client.Set(ctx, "game:"+game.ID, dataBytes, r.ttl)
	r.client.Set(ctx, "usergame:"+userID, gameID, r.ttl)
	// продливаем ttl у оппонента
	refreshUserGameTTL(ctx, r.client, game, r.ttl)

	return renderBoard(&game, "Игра началась! Ходит: "+game.Turn)
}

func (r *RedisCache) Move(ctx context.Context, userID, coord string) OutgoingMessage {

	gameID, _ := r.client.Get(ctx, "usergame:"+userID).Result()
	if gameID == "" {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Вы не в игре",
		}
	}

	// Пробуем установить блокировку
	lock, err := lockGame(ctx, r.locker, gameID, r.ttlLock)
	if err != nil || err == redislock.ErrNotObtained {
		return OutgoingMessage{
			UserID: userID,
			Text:   "Игра сейчас занята, попробуйте позже",
		}
	}
	// Откладываем ручной релиз, если функция завершится раньше
	defer lock.Release(ctx)

	data, _ := r.client.Get(ctx, "game:"+gameID).Result()
	var game Game
	_ = json.Unmarshal([]byte(data), &game)

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
		dataBytes, _ := json.Marshal(game)
		r.client.Set(ctx, "game:"+game.ID, dataBytes, r.ttl)
		refreshUserGameTTL(ctx, r.client, game, r.ttl)
		return renderBoard(&game, fmt.Sprintf("Победа: %s", mark))
	}
	if isDraw(game.Board) {
		game.Finished = true
		dataBytes, _ := json.Marshal(game)
		r.client.Set(ctx, "game:"+game.ID, dataBytes, r.ttl)
		refreshUserGameTTL(ctx, r.client, game, r.ttl)
		return renderBoard(&game, "Ничья!")
	}
	game.Turn = game.PlayerX
	if userID == game.PlayerX {
		game.Turn = game.PlayerO
	}
	dataBytes, _ := json.Marshal(game)
	r.client.Set(ctx, "game:"+game.ID, dataBytes, r.ttl)
	refreshUserGameTTL(ctx, r.client, game, r.ttl)
	return renderBoard(&game, "Ход противника")
}

func lockGame(ctx context.Context, locker  *redislock.Client, gameID string, ttl time.Duration) (*redislock.Lock, error) {
	lockKey := "lock:game:" + gameID
	for i := 0; i < 3; i++ {
		lock, err := locker.Obtain(ctx, lockKey, ttl, nil)
		if err == nil {
			return lock, nil
		}
		if err != redislock.ErrNotObtained {
			return nil, err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, redislock.ErrNotObtained
}

func refreshUserGameTTL(ctx context.Context, client  *redis.Client, game Game, ttl time.Duration) {
	if game.PlayerX != "" {
		client.Expire(ctx, "usergame:"+game.PlayerX, ttl)
	}
	if game.PlayerO != "" {
		client.Expire(ctx, "usergame:"+game.PlayerO, ttl)
	}
}