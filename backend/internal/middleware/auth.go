package middleware

import (
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/labstack/echo/v5"
	  "github.com/clerk/clerk-sdk-go/v2"

)

type AuthMiddleware struct {
	server *server.Server
}

func NewAuthMiddleware(s *server.Server) *AuthMiddleware {
	return &AuthMiddleware{
		server: s,
	}
}

// Clerk handles authentication. Your Go backend handles application logic and authorization.

func (auth *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.WrapMiddleware()()
}
