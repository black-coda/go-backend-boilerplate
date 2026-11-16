package middleware

import (
	"errors"
	"net/http"

	"github.com/black-coda/go-boilerplate/internal/errs"
	"github.com/black-coda/go-boilerplate/internal/server"
	"github.com/black-coda/go-boilerplate/internal/sqlerr"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

type GlobalMiddlewares struct {
	server server.Server
}

func NewGlobalMiddlewares(server server.Server) *GlobalMiddlewares {
	return &GlobalMiddlewares{
		server: server,
	}
}

func (g *GlobalMiddlewares) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

		},
	)
}

func (g *GlobalMiddlewares) CORS() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(
		middleware.CORSConfig{
			AllowOrigins: g.server.Config.Server.CORSAllowedOrigins,
		},
	)
}

func (g *GlobalMiddlewares) Recover() echo.MiddlewareFunc {
	return middleware.Recover()
}

func (g *GlobalMiddlewares) Secure() echo.MiddlewareFunc {
	return middleware.Secure()
}

// GlobalErrorHandler runs when any handler or middleware returns an error.
// It turns every error into one JSON response, in the same shape.
// Python comparison: @app.errorhandler(Exception) in Flask.
func (g *GlobalMiddlewares) GlobalErrorHandler(err error, c *echo.Context) {
	// Save the original error. We log it at the end.
	// The variable "err" may change below, and we do not want to lose the real cause.
	originalErr := err

	// STEP 1: Convert the error, if needed.
	// Goal: change unknown errors into our own HTTPError type.

	// This variable is empty now. errors.As fills it if "err" is an HTTPError.
	var httpErr *errs.HTTPError

	// Only convert errors that are NOT already our custom HTTPError.
	// If it is already an HTTPError, we keep it as it is.
	if !errors.As(err, &httpErr) {
		// Check if it is an error from Echo itself.
		// Example: Echo returns this error when no route matches the URL.
		if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
			// Echo says "404". We replace it with our own "not found" error,
			// so the client gets our JSON shape.
			if echoErr.Code == http.StatusNotFound {
				err = errs.NewNotFoundError("Route not found", false, nil)
			}
			// Other Echo errors stay as they are. Step 2 handles them.
		} else {
			// The error is not ours and not from Echo.
			// It can be a database error, for example a duplicate email.
			// sqlerr.HandleError changes known database errors into HTTPError
			// (for example, unique violation becomes a 409 error).
			// If the error is unknown, it stays the same.
			err = sqlerr.HandleError(err)
		}
	}

	// STEP 2: Read the status, code, and message from the error.
	// "err" may be a different error now, so we check its type again.

	// These variables hold the values for the JSON response.
	var echoErr *echo.HTTPError
	var status int                    // HTTP status code, for example 404
	var code string                   // short code, for example "NOT_FOUND"
	var message string                // text for people
	var fieldErrors []errs.FieldError // errors for single form fields
	var action *errs.Action           // what the client must do next, can be nil

	// A switch with no value works like if / else if / else.
	// Go runs the first case that is true and skips the rest.
	switch {
	// Case 1: our own error. We trust all of its fields.
	case errors.As(err, &httpErr):
		status = httpErr.Status
		code = httpErr.Code
		message = httpErr.Message
		fieldErrors = httpErr.Errors
		action = httpErr.Action

	// Case 2: an Echo error that we did not convert.
	case errors.As(err, &echoErr):
		status = echoErr.Code
		// http.StatusText(404) returns "Not Found".
		// MakeUpperCaseWithUnderscores changes it to "NOT_FOUND".
		code = errs.MakeUpperCaseWithUnderscores(http.StatusText(status))

		// echoErr.Message can have any type. We need a string.
		// This check is like "isinstance(msg, str)" in Python.
		msg := echoErr.Message
		if msg == "" {
			message = msg
		} else {
			// It is not a string. Use the standard text for the status.
			message = http.StatusText(echoErr.Code)
		}

	// Default: an unknown error. We do not know what went wrong.
	// Send 500 and a safe message. Never send the raw error text to the client.
	// Raw text can show private details, such as table names or file paths.
	default:
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(
			http.StatusText(http.StatusInternalServerError),
		)
		message = http.StatusText(http.StatusInternalServerError)
	}

	// STEP 3: Write the log line.
	// We log the ORIGINAL error, not the converted one.
	// The original error has the real cause, so it helps you debug.
	// The client sees the safe message. The log has the full detail.

	// GetLogger returns the logger for this request.
	// It already has the request ID, the user ID, and other fields.
	logger := *GetLogger(c)

	// .Stack() adds the stack trace to the log line.
	logger.Error().Stack().
		Err(originalErr).
		Int("status", status).
		Str("error_code", code).
		Msg(message)

	// STEP 4: Send the JSON response to the client.

	// "Committed" is true when Echo has already started to send a response.
	// You cannot send two responses. In that case, we skip this step.
	if resp, _ := echo.UnwrapResponse(c.Response()); resp != nil && !resp.Committed {
		// The "_ =" ignores the error from c.JSON.
		// At this point, we have no better way to report a write failure.
		_ = c.JSON(
			status, errs.HTTPError{
				Code:    code,
				Message: message,
				Status:  status,
				// Override is true only if the original error was our HTTPError
				// and it had Override set. "httpErr != nil" prevents a crash
				// when httpErr is empty.
				Override: httpErr != nil && httpErr.Override,
				Errors:   fieldErrors,
				Action:   action,
			},
		)
	}
}

func (g *GlobalMiddlewares) RequestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(
		middleware.RequestLoggerConfig{
			LogURI:     true,
			LogStatus:  true,
			LogLatency: true,
			LogHost:    true,
			LogMethod:  true,
			LogURIPath: true,
			// Echo calls this function after it processes the request.
			LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
				statusCode := v.Status

				if v.Error != nil {
					var echoErr *echo.HTTPError

					if _, ok := errors.AsType[*errs.HTTPError](v.Error); ok {
						statusCode = echoErr.Code
					} else if errors.As(v.Error, &echoErr) {
						statusCode = echoErr.Code
					}
				}

				logger := GetLogger(c)

				var e *zerolog.Event

				switch {
				case statusCode >= 500:
					e = logger.Error().Err(v.Error)
				case statusCode >= 400:
					e = logger.Warn()
				default:
					e = logger.Info()
				}

				// Add request ID if available
				if requestID := GetRequestID(c); requestID != "" {
					e = e.Str("request_id", requestID)
				}

				// Add user context if available
				if userID := GetUserID(c); userID != "" {
					e = e.Str("user_id", userID)
				}

				e.
					Dur("latency", v.Latency).
					Int("status", statusCode).
					Str("method", v.Method).
					Str("uri", v.URI).
					Str("host", v.Host).
					Str("ip", c.RealIP()).
					Str("user_agent", c.Request().UserAgent()).
					Msg("API")

				return nil
			},
		},
	)
}
