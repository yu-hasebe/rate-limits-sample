package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type middleware struct {
	redisClient *redis.Client
}

func (m *middleware) limitRate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, err := getIP(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := m.consumeToken(r.Context(), ip); err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (m *middleware) getTokens(ctx context.Context, ip string) int64 {
	currentTime := time.Now().Unix()
	lastRefilledTime, err := m.redisClient.HGet(ctx, ip, "last_refilled_time").Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0
	}
	tokens, err := m.redisClient.HGet(ctx, ip, "tokens").Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0
	}
	elapsedTime := currentTime - lastRefilledTime
	if elapsedTime >= 10 {
		tokens = 10
	}
	m.redisClient.HSet(ctx, ip, "last_refilled_time", currentTime, "tokens", tokens)
	return tokens
}

func (m *middleware) consumeToken(ctx context.Context, ip string) error {
	tokens := m.getTokens(ctx, ip)
	if tokens <= 0 {
		return errors.New("not enough tokens")
	}
	m.redisClient.HIncrBy(ctx, ip, "tokens", -1)
	return nil
}

func getIP(r *http.Request) (string, error) {
	ipAddresses := r.Header.Get("X-FORWARDED-FOR")
	splitedIPs := strings.Split(ipAddresses, ",")
	if len(splitedIPs) > 0 {
		netIP := net.ParseIP(splitedIPs[len(splitedIPs)-1])
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}
		return ip, nil
	}

	return "", errors.New("IP not found")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
		PoolSize: 1000,
	})
	m := middleware{
		redisClient: redisClient,
	}
	http.HandleFunc("/", m.limitRate(handler))
	http.ListenAndServe(":8080", nil)
}
