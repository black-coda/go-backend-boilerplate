package handler

import (
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/black-coda/go-boilerplate/internal/service"
)

type Handlers struct {
}

func NewHandler(s server.Server, services service.Service) *Handlers {
	return &Handlers{}
}
