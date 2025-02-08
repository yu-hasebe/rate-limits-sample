package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	UNIT_SECONDS     = 10
	REQUEST_PER_UNIT = 10
)

type RateLimiter struct {
	redisClient *redis.Client
}

// token bucket algorithm
func (r *RateLimiter) isRateLimited(ctx context.Context, ip string) (bool, error) {
	tokens, err := r.getTokens(ctx, ip)
	if err != nil {
		return true, err
	}
	if tokens <= 0 {
		return true, nil
	}
	r.redisClient.HIncrBy(ctx, ip, "tokens", -1)
	return false, nil
}

func (r *RateLimiter) getTokens(ctx context.Context, ip string) (int64, error) {
	currentTime := time.Now().Unix()
	lastRefilledTime, err := r.redisClient.HGet(ctx, ip, "last_refilled_time").Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}
	tokens, err := r.redisClient.HGet(ctx, ip, "tokens").Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}
	elapsedTime := currentTime - lastRefilledTime
	if elapsedTime >= UNIT_SECONDS {
		tokens = REQUEST_PER_UNIT
		r.redisClient.HSet(ctx, ip, "last_refilled_time", currentTime, "tokens", tokens)
	}

	log.Printf("ip: %s, currentTime: %d, lastRefilledTime: %d, elapsedTime: %d, token: %d", ip, currentTime, lastRefilledTime, elapsedTime, tokens)

	return tokens, nil
}
