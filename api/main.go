package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
		PoolSize: 1000,
	})
	rdb.Set(ctx, "my-key", "my-value", 0)
	result, err := rdb.Get(ctx, "my-key").Result()
	if err != nil {
		fmt.Fprintf(w, "error: %+v\n", err)
		return
	}

	fmt.Fprintf(w, "hello world: %s\n", result)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
