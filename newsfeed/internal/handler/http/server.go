package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	grpc_pb "ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/pkg/logger"
)

type Config struct {
	Host   string
	Port   int
	JwtKey []byte
}

func verifyConfig(cfg Config) error {
	if len(cfg.Host) == 0 {
		return errors.New("host is required")
	}
	if cfg.Port <= 0 {
		return errors.New("port is required")
	}
	return nil
}

type Server struct {
	config Config

	httpServer *http.Server
	router     *gin.Engine

	grpcClient grpc_pb.ServiceClient
}

func New(config Config, grpcClient grpc_pb.ServiceClient) (*Server, error) {
	err := verifyConfig(config)
	if err != nil {
		logger.Error("invalid http config", logger.E(err))
		return nil, fmt.Errorf("invalid http config: %s", err)
	}

	h := &Server{
		config:     config,
		grpcClient: grpcClient,
	}

	// init gin handlers
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(h.MonitorMiddleware())

	userRouter := router.Group("/grpc")
	userRouter.POST("/signup", h.Signup)
	userRouter.POST("/login", h.Login)

	userMeRouter := userRouter.Group("/me")
	userMeRouter.Use(h.JWTMiddleware())
	userMeRouter.POST("/follow", h.Follow)
	userMeRouter.GET("/followers", h.GetFollowers)
	userMeRouter.GET("/followings", h.GetFollowings)

	postRouter := router.Group("/post")
	postMeRouter := postRouter.Group("/me")
	postMeRouter.Use(h.JWTMiddleware())
	postMeRouter.POST("/", h.CreatePost)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h.router = router

	// init http server
	addr := fmt.Sprintf("%s:%d", h.config.Host, h.config.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	h.httpServer = server

	return h, nil
}

func (h *Server) Start() error {
	return h.httpServer.ListenAndServe()
}

func (h *Server) Stop() error {
	return h.httpServer.Shutdown(context.Background())
}
