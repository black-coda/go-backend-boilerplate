package middleware

import "github.com/black-coda/go-boilerplate/internal/server"

type RateLimitMiddleware struct {
	server *server.Server
}

func NewRateLimitMiddleware(server *server.Server) *RateLimitMiddleware {
	return &RateLimitMiddleware{server: server}
}

func (rm RateLimitMiddleware) RecordRateLimitHit(endpoint string) {
	if rm.server.LoggerService != nil && rm.server.LoggerService.GetApplication() != nil {
		rm.server.LoggerService.GetApplication().RecordCustomEvent(
			"RateLimitHit", map[string]any{
				"endpoint": endpoint,
			},
		)
	}
}
