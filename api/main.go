package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
)

type middleware struct {
	rateLimiter *RateLimiter
}

func (m *middleware) limitRate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, err := getIP(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		isRateLimited, err := m.rateLimiter.isRateLimited(r.Context(), ip)
		if isRateLimited {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	}
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
	r := RateLimiter{
		redisClient: redisClient,
	}
	m := middleware{
		rateLimiter: &r,
	}
	http.HandleFunc("/", m.limitRate(handler))
	http.ListenAndServe(":8080", nil)
}
