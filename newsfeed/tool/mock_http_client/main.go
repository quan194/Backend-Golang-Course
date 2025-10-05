package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type config struct {
	HttpHost string `env:"HTTP_HOST"`
	HttpPort int    `env:"HTTP_PORT"`
}

func main() {
	godotenv.Load(".env")
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	log.Println("cfg", cfg.HttpHost, cfg.HttpPort)

	rateLimiter := rate.NewLimiter(10, 5)
	ctx := context.Background()
	for {
		rateLimiter.Wait(ctx)

		httpClient := &http.Client{}
		resp, err := httpClient.Get(getFollowingsUrl(cfg.HttpHost, cfg.HttpPort))
		if err != nil {
			log.Println("err", err)
		} else {
			log.Println("resp", resp.Status)
		}
	}
}

func getFollowingsUrl(host string, port int) string {
	return fmt.Sprintf("http://%s:%d/user/me/followings?last_value=%d&limit=10&mock=true", host, port, time.Now().Unix())
}
