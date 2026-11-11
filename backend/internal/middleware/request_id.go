package middleware

import (
	"context"
	"uuid"

	"github.com/labstack/echo/v5"
)

const (
	// RequestIDHeader is the header name for the request ID.
	// The client can send it. The server sends it back in the response.
	RequestIDHeader = "X-Request-ID"

	// RequestIDKey is the key for Echo's own store (c.Set and c.Get).
	RequestIDKey = "request_id"

	// maxRequestIDLen is the longest ID we accept from a client.
	// A longer value is replaced with a new ID.
	maxRequestIDLen = 64
)

// ctxKey is the key for the request ID in a Go context.
// The type is private, so no other package can use the same key.
type ctxKey struct{}

// RequestID gives each request a unique ID.
// Register it first, so all other middleware can use the ID.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Use the ID from the client, if the header exists and is not too long.
			requestID := c.Request().Header.Get(RequestIDHeader)
			if requestID == "" || len(requestID) > maxRequestIDLen {
				// Example: 4c90fc3f-39cc-4b04-af21-c83ee64aa67e
				requestID = uuid.New().String()
			}

			// Save the ID in Echo's store. Only handlers and middleware can read it.
			c.Set(RequestIDKey, requestID)

			// Send the ID back to the client in the response header.
			c.Response().Header().Set(RequestIDHeader, requestID)

			// Make a new Go context that contains the ID.
			// A context cannot change, so WithValue returns a new one.
			ctx := context.WithValue(c.Request().Context(), ctxKey{}, requestID)

			// Put the new context in the request. This step is necessary.
			// Without it, c.Request().Context() returns the old context.
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// GetRequestID reads the ID from Echo's store.
// Use it in handlers and middleware, where you have the Echo context.
func GetRequestID(c *echo.Context) string {
	if requestID, ok := c.Get(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// FromContext reads the ID from a Go context.
// Use it in services, repos, and jobs, where you have no Echo context.
// It returns an empty string if the context has no ID.
func FromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(ctxKey{}).(string)
	return requestID
}
