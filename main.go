package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}

type FloodClient struct {
	client *redis.Client
	N      time.Duration
	K      int64
}

func New(client *redis.Client, N time.Duration, K int64) *FloodClient {
	return &FloodClient{
		client: client,
		N:      N,
		K:      K,
	}
}

func (fl *FloodClient) Check(ctx context.Context, userID int64) (bool, error) {

	id := strconv.FormatInt(userID, 10)

	first, err := fl.client.Get(id).Int64()
	if errors.Is(err, redis.Nil) {
		return false, err
	}

	res, err := fl.client.Incr(id).Result()
	if err != nil {
		return false, err
	}

	now := time.Now()
	if now.Sub(time.Unix(first, 0)) >= fl.N {
		return res <= fl.K, nil
	}
	return true, nil
}

func main() {

	redisAddr := "localhost:6379"

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	var flood FloodControl
	flood = New(client, 5*time.Second, 5)

	// Проверка алгоритма на 10 запросах
	for i := 0; i < 10; i++ {
		res, err := flood.Check(context.Background(), 1)
		if err != nil {
			fmt.Println("Error checking flood control")
			return
		}
		if res {
			fmt.Println("Request allowed")
		} else {
			fmt.Println("Request denied for")
		}
	}
}
