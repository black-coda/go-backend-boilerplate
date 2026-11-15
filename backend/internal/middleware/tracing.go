package middleware

import (
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/labstack/echo/v5"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v5" // use nrecho-v5 for Echo v5
	"github.com/newrelic/go-agent/v3/integrations/nrpkgerrors"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type TracingMiddleware struct {
	server *server.Server
	nrApp  *newrelic.Application
}

func NewTracingMiddleware(server *server.Server, nrApp *newrelic.Application) *TracingMiddleware {
	return &TracingMiddleware{server: server, nrApp: nrApp}
}

func (tm *TracingMiddleware) NewRelicMiddleware() echo.MiddlewareFunc {
	if tm.nrApp == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}

	}
	return nrecho.Middleware(tm.nrApp)
}

func (tm *TracingMiddleware) EnhanceTracing() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Get NewRelic transaction from context
			txn := newrelic.FromContext(c.Request().Context())
			if txn == nil {
				return next(c)
			}

			// service.name and service.environment are already set in logger and New Relic config
			txn.AddAttribute("http.real_ip", c.RealIP())
			txn.AddAttribute("http.user_agent", c.Request().UserAgent())

			// Add request ID if available
			if requestID := GetRequestID(c); requestID != "" {
				txn.AddAttribute("request.id", requestID)
			}

			// Add user context if available
			if userID := c.Get("user_id"); userID != nil {
				if userIDStr, ok := userID.(string); ok {
					txn.AddAttribute("user.id", userIDStr)
				}
			}

			// Execute next handler
			err := next(c)
			// If the handler returns an error, you report it to New Relic.
			if err != nil {
				txn.NoticeError(nrpkgerrors.Wrap(err))
			}

			// Add response status
			resp, err := echo.UnwrapResponse(c.Response())
			txn.AddAttribute("http.status_code", resp.Status)

			return err
		}
	}
}
