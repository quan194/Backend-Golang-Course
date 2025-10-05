package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/pkg/logger"
)

func CustomizedInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		grpcCode := status.Convert(err).Code().String()
		method := info.FullMethod
		latency := time.Since(start)

		logFields := []logger.Field{
			logger.F("method", method),
			logger.F("request", req),
			logger.F("response", resp),
			logger.F("latency", latency),
			logger.F("code", grpcCode),
		}
		if err != nil {
			logFields = append(logFields, logger.E(err))
			logger.Error("processed grpc request with error", logFields...)

			// if err is AppError, replace it
			appErr := &common.AppError{}
			if errors.As(err, &appErr) {
				err = common.ToGRPCError(appErr)
			}
			return resp, err
		}

		logger.Info("processed grpc request", logFields...)
		return resp, err
	}
}
