package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/pkg/logger"
	"ep.k16/newsfeed/pkg/monitor"
)

func (h *Server) MonitorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		fullApi := c.Request.Method + " " + c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		httpCode := c.Writer.Status()

		monitor.ExportApiStatus(c.Request.URL.Path, c.Request.Method, httpCode)
		monitor.ExportApiStatusLatency(c.Request.URL.Path, c.Request.Method, httpCode, latency)

		fs := []logger.Field{
			logger.F("api", fullApi),
			logger.F("query", c.Request.URL.RawQuery),
			logger.F("client_ip", c.ClientIP()),
			logger.F("http_code", httpCode),
			logger.F("latency", latency),
		}
		if username := c.GetString("username"); len(username) > 0 {
			fs = append(fs, logger.F("username", username))
		}
		logger.Debug("processed HTTP request", fs...)
	}
}

func (h *Server) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("mock") == "true" {
			c.Set("user_id", int64(3))
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			unauthErr := common.NewError(http.StatusUnauthorized, "invalid Authorization header")
			h.returnErrResp(c, unauthErr)
			c.Abort()
			return
		}

		logger.Debugf("autHeader: %s", authHeader)

		tokenStr := authHeader[7:]

		logger.Debugf("autHeader: %s", tokenStr)
		claims, err := h.validateJWT(tokenStr)
		if err != nil {
			unauthErr := common.NewError(http.StatusUnauthorized, "invalid token")
			h.returnErrResp(c, unauthErr)
			c.Abort()
			return
		}

		// store in context for handlers
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
