package middleware

import (
	"context"

	"github.com/black-coda/go-boilerplate/internal/logger"
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/rs/zerolog"

	"github.com/labstack/echo/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	LoggerKey   = "logger"
)

type ContextEnhancer struct {
	server *server.Server
}

func NewContextEnhancer(s *server.Server) *ContextEnhancer {
	return &ContextEnhancer{server: s}
}

func (ce *ContextEnhancer) EnhanceContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// 1. Collect request information
			request := c.Request()
			requestID := GetRequestID(c)

			// Create enhanced logger with request context
			contextLogger := ce.server.Logger.With().
				Str("request_id", requestID).
				Str("method", c.Request().Method).
				Str("path", c.Path()).
				Str("ip", c.RealIP()).
				Logger()

			if txn := newrelic.FromContext(request.Context()); txn != nil {
				contextLogger = logger.WithTraceContext(contextLogger, txn)
			}

			// Get authenticated user information
			if userID := ce.extractUserID(c); userID != "" {
				contextLogger = contextLogger.With().Str("user_id", userID).Logger()
			}

			if userRole := ce.extractUserRole(c); userRole != "" {
				contextLogger = contextLogger.With().Str("user_role", userRole).Logger()
			}

			// Stores the logger in Echo's context.
			c.Set(LoggerKey, contextLogger)
			// Stores the logger in Go's request context.
			ctx := context.WithValue(request.Context(), LoggerKey, contextLogger)
			// Replaces the request with one containing the updated context.
			c.SetRequest(request.WithContext(ctx))
			return next(c)
		}
	}
}

func (ce *ContextEnhancer) extractUserID(c *echo.Context) string {
	return GetUserID(c)
}

func (ce *ContextEnhancer) extractUserRole(c *echo.Context) string {
	return GetUserRole(c)
}

func GetUserID(c *echo.Context) string {
	if userID, ok := c.Get(UserIDKey).(string); ok {
		return userID
	}

	return ""
}

func GetUserRole(c *echo.Context) string {
	if role, ok :=
		c.Get(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

func GetLogger(c *echo.Context) *zerolog.Logger {
	if internalLogger, ok := c.Get(LoggerKey).(*zerolog.Logger); ok {
		return internalLogger
	}

	internalLogger := zerolog.Nop()
	return &internalLogger
}
