package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/black-coda/go-boilerplate/internal/errs"
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/labstack/echo/v5"
	zlog "github.com/rs/zerolog/log"
)

type AuthMiddleware struct {
	server *server.Server
}

func NewAuthMiddleware(s *server.Server) *AuthMiddleware {
	return &AuthMiddleware{
		server: s,
	}
}

func sendError401(auth *AuthMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		response := map[string]string{
			"error":    "Unauthorized",
			"message":  "You are not authorized to access this resource.",
			"override": "false",
			"status":   "401",
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			auth.server.Logger.Error().Err(err).Str("function", "RequireAuth").Dur(
				"duration", time.Since(start),
			).Msg("failed to write JSON response")
		} else {
			auth.server.Logger.Info().Str("function", "RequireAuth").Dur(
				"duration", time.Since(start),
			).Msg("Unauthorized access attempt")
		}
	}
}

func (auth *AuthMiddleware) RequireAuth2(next echo.HandlerFunc) echo.HandlerFunc {
	failureHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			response := map[string]string{
				"error":    "Unauthorized",
				"message":  "You are not authorized to access this resource.",
				"override": "false",
				"status":   "401",
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				auth.server.Logger.Error().Err(err).Str("function", "RequireAuth").Dur(
					"duration", time.Since(start),
				).Msg("failed to write JSON response")
			} else {
				auth.server.Logger.Info().Str("function", "RequireAuth").Dur(
					"duration", time.Since(start),
				).Msg("Unauthorized access attempt")
			}

		},
	)

	clerkMiddleware := clerkhttp.WithHeaderAuthorization(
		clerkhttp.AuthorizationFailureHandler(failureHandler),
	)

	echoMiddleware := echo.WrapMiddleware(clerkMiddleware)

	claimHandler := func(c *echo.Context) error {
		// Authentication Logic
		start := time.Now()

		claims, ok := clerk.SessionClaimsFromContext(c.Request().Context())

		if !ok {
			auth.server.Logger.Error().Str("function", "RequireAuth").
				Str("request_id", GetRequestID(c)).
				Dur("duration", time.Since(start)).
				Msg("could not get session claims from context")

			return errs.NewUnauthorizedError("Unauthorized: could not get session claims from context", false)
		}

		c.Set("user_id", claims.Subject)
		c.Set("user_role", claims.ActiveOrganizationRole)
		c.Set("permissions", claims.Claims.ActiveOrganizationPermissions)

		auth.server.Logger.Info().Str("function", "RequireAuth").Str("request_id", GetRequestID(c)).Str(
			"user_id",
			claims.Subject,
		).Dur("duration", time.Since(start)).Msg("Authorized and Authenticated successfully")
		return next(c)
	}
	return echoMiddleware(claimHandler)
}

// RequireAuth Clerk handles authentication. Your Go backend handles application logic and authorization.
// Middleware execute in registration order before the handler. They execute in reverse order after the handler.
func (auth *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	// This converts Clerk's standard HTTP middleware into an Echo handler.
	// This changes the Clerk middleware into an Echo middleware. Echo cannot use the Clerk kind directly.
	return echo.WrapMiddleware(
		clerkhttp.WithHeaderAuthorization(
			// This is the handler that will be called when authorization fails. It sends a 401 Unauthorized response.
			clerkhttp.AuthorizationFailureHandler(sendError401(auth)),
		),
	)(
		func(c *echo.Context) error {
			// Authentication Logic
			start := time.Now()

			claims, ok := clerk.SessionClaimsFromContext(c.Request().Context())

			if !ok {
				auth.server.Logger.Error().Str("function", "RequireAuth").
					Str("request_id", GetRequestID(c)).
					Dur("duration", time.Since(start)).
					Msg("could not get session claims from context")

				return errs.NewUnauthorizedError("Unauthorized: could not get session claims from context", false)
			}

			c.Set("user_id", claims.Subject)
			c.Set("user_role", claims.ActiveOrganizationRole)
			c.Set("permissions", claims.Claims.ActiveOrganizationPermissions)

			auth.server.Logger.Info().Str("function", "RequireAuth").Str("request_id", GetRequestID(c)).Str(
				"user_id",
				claims.Subject,
			).Dur("duration", time.Since(start)).Msg("Authorized and Authenticated successfully")
			return next(c)
		},
	)
}

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		fmt.Println("Before LoggerMiddleware")
		start := time.Now()
		err := next(c)
		zlog.Info().Str("LoggerMiddleware", "Before").Str("function", "LoggerMiddleware").Dur(
			"duration",
			time.Since(start),
		).Msg("Before LoggerMiddleware")
		fmt.Println("After LoggerMiddleware")
		return err
	}
}
