package service

import (
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/clerk/clerk-sdk-go/v2"
)

type AuthService struct {
	server *server.Server
}

func NewAuthenticationService(server *server.Server) *AuthService {
	clerk.SetKey(server.Config.Auth.Clerk.ClerkSecretKey)
	return &AuthService{server: server}
}
