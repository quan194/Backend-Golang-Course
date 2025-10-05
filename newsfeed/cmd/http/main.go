package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ep.k16/newsfeed/cmd"
	"ep.k16/newsfeed/config"
	"ep.k16/newsfeed/internal/handler/http"
	grpc_pb "ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/pkg/logger"
)

func main() {
	// init logger
	if err := cmd.InitLogger(); err != nil {
		return
	}

	// init config
	cfg, err := config.LoadHttpConfig()
	if err != nil {
		logger.Error("failed to init http config", logger.E(err))
		return
	}
	logger.Info("init http config successfully", logger.F("cfg", cfg))

	// init dependencies: grpc client
	grpcAddr := fmt.Sprintf("%s:%d", cfg.GrpcHost, cfg.GrpcPort)
	grpcConn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to init grpc grpc client", logger.E(err))
		return
	}
	grpcCli := grpc_pb.NewServiceClient(grpcConn)

	// create http server
	httpServer, err := http.New(http.Config{
		Host:   cfg.Host,
		Port:   cfg.Port,
		JwtKey: []byte(cfg.JwtKey),
	}, grpcCli)
	if err != nil {
		logger.Error("failed to init http server", logger.E(err))
		return
	}

	// run servers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)

	go func() {
		err = httpServer.Start() // blocking -> run it on a different goroutine so it won't block main goroutine
		if err != nil {
			logger.Error("failed to start http server", logger.E(err))
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		logger.Info("process received signal, shutting down", logger.F("signal", sig))
	case err := <-errChan:
		logger.Info("process met error, shutting down", logger.E(err))
		logger.Error("process met error, shutting down", logger.E(err))
	}

	// shutdown
	// TODO: add timeout for shutdown
	httpServer.Stop()
	grpcConn.Close()

	logger.Info("process stopped")
}
