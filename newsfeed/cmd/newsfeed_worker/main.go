package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"ep.k16/newsfeed/cmd"
	"ep.k16/newsfeed/config"
	"ep.k16/newsfeed/internal/dao/post_cache"
	"ep.k16/newsfeed/internal/dao/user_cache"
	"ep.k16/newsfeed/internal/handler/newsfeed_processor"
	"ep.k16/newsfeed/internal/service/post_service"
	"ep.k16/newsfeed/pkg/logger"
)

func main() {
	// init logger
	if err := cmd.InitLogger(); err != nil {
		return
	}

	// init config
	cfg, err := config.LoadNewsfeedWorkerConfig()
	if err != nil {
		logger.Error("failed to init newsfeed worker config", logger.E(err))
		return
	}
	logger.Info("init newsfeed worker config successfully", logger.F("cfg", cfg))

	// create cache
	var postCacheDai post_service.PostCacheDAI
	postCacheDai, err = post_cache.New(post_cache.CacheConfig{
		Host: cfg.RedisHost,
		Port: cfg.RedisPort,
		TTL:  0,
	})
	if err != nil {
		logger.Error("failed to init post cache", logger.E(err))
		return
	}

	var userCacheDai post_service.UserCacheDAI
	userCacheDai, err = user_cache.New(user_cache.CacheConfig{
		Host: cfg.RedisHost,
		Port: cfg.RedisPort,
		TTL:  0,
	})
	if err != nil {
		logger.Error("failed to init grpc cache", logger.E(err))
		return
	}

	// create service
	newsfeedService, err := post_service.New(nil, userCacheDai, postCacheDai, nil)
	if err != nil {
		logger.Error("failed to init post service", logger.E(err))
		return
	}

	// create handler
	newsfeedConfig := newsfeed_processor.Config{
		Brokers:       cfg.KafkaBrokers,
		Topic:         cfg.KafkaTopic,
		ConsumerGroup: cfg.KafkaConsumerGroup,
	}
	newsfeedProcessor, err := newsfeed_processor.New(newsfeedConfig, newsfeedService)
	if err != nil {
		logger.Error("failed to init newsfeed processor", logger.E(err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	// run servers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)

	go func() {
		err = newsfeedProcessor.Start(ctx) // blocking -> run it on a different goroutine so it won't block main goroutine
		if err != nil {
			logger.Error("failed to start newsfeed processor", logger.E(err))
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		logger.Info("process received signal, shutting down", logger.F("signal", sig))
	case err := <-errChan:
		logger.Error("process met error, shutting down", logger.E(err))
	}

	// shutdown
	// TODO: add timeout for shutdown
	cancel()
	newsfeedProcessor.Stop()

	logger.Info("process stopped")
}
