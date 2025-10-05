package main

import (
	"os"
	"os/signal"
	"syscall"

	"ep.k16/newsfeed/cmd"
	"ep.k16/newsfeed/config"
	"ep.k16/newsfeed/internal/dao/kafka_producer"
	"ep.k16/newsfeed/internal/dao/post_cache"
	"ep.k16/newsfeed/internal/dao/post_dao"
	"ep.k16/newsfeed/internal/dao/user_cache"
	"ep.k16/newsfeed/internal/dao/user_dao"
	"ep.k16/newsfeed/internal/handler/grpc"
	"ep.k16/newsfeed/internal/service/post_service"
	"ep.k16/newsfeed/internal/service/user_service"
	"ep.k16/newsfeed/pkg/logger"
)

func main() {
	// init logger
	if err := cmd.InitLogger(); err != nil {
		return
	}

	// init config
	cfg, err := config.LoadGrpcConfig()
	if err != nil {
		logger.Error("failed to init grpc config", logger.E(err))
		return
	}
	logger.Info("init grpc config successfully", logger.F("cfg", cfg))

	// create db conn -> db access object
	userDao, err := user_dao.New(&user_dao.UserDbConfig{
		Username:     cfg.DatabaseUser,
		Password:     cfg.DatabasePassword,
		Host:         cfg.DatabaseHost,
		Port:         cfg.DatabasePort,
		DatabaseName: cfg.DatabaseName,
	})
	if err != nil {
		logger.Error("failed to init dai", logger.E(err))
		return
	}

	// create cache
	var userCacheDai user_service.UserCacheDAI
	if cfg.RedisEnabled {
		userCacheDai, err = user_cache.New(user_cache.CacheConfig{
			Host: cfg.RedisHost,
			Port: cfg.RedisPort,
			TTL:  0,
		})
		if err != nil {
			logger.Error("failed to init grpc cache", logger.E(err))
			return
		}
	}

	// create db conn -> db access object
	postDao, err := post_dao.New(&post_dao.PostDbConfig{
		Username:     cfg.DatabaseUser,
		Password:     cfg.DatabasePassword,
		Host:         cfg.DatabaseHost,
		Port:         cfg.DatabasePort,
		DatabaseName: cfg.DatabaseName,
	})
	if err != nil {
		logger.Error("failed to init post dai", logger.E(err))
		return
	}

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

	userService, err := user_service.New(userDao, userCacheDai)
	if err != nil {
		logger.Error("failed to init grpc service", logger.E(err))
		return
	}
	// create msg producer
	kafkaProducer, err := kafka_producer.New(kafka_producer.KafkaConfig{
		Brokers: cfg.KafkaBrokers,
		Topic:   cfg.KafkaTopic,
	})
	if err != nil {
		logger.Error("failed to init kafka producer", logger.E(err))
		return
	}

	postService, err := post_service.New(postDao, userCacheDai, postCacheDai, kafkaProducer)
	if err != nil {
		logger.Error("failed to init post service", logger.E(err))
		return
	}

	grpcConfig := grpc.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	}
	grpcServer, err := grpc.New(grpcConfig, userService, postService)
	if err != nil {
		logger.Error("failed to init grpc grpc server", logger.E(err))
		return
	}

	// run servers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)

	go func() {
		err = grpcServer.Start() // blocking -> run it on a different goroutine so it won't block main goroutine
		if err != nil {
			logger.Error("failed to start grpc grpc server", logger.E(err))
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
	grpcServer.Stop()
	userDao.Stop()

	logger.Info("process stopped")
}
