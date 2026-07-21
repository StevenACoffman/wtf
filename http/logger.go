package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
)

// loggerContextKey is the context key under which the request-scoped logger is
// stored.
type loggerContextKey struct{}

// contextWithLogger returns a copy of ctx carrying the request-scoped logger.
func contextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// loggerFromContext returns the request-scoped logger stored on ctx. If none is
// present it falls back to a plain stderr logger so callers never deal with nil.
func loggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

// withLogger is middleware that attaches a request-scoped logger to the request
// context. The logger is derived from the server's base logger and tagged with a
// correlation ID plus the request method & path, so every line emitted while
// handling the request can be traced back to it. The correlation ID is taken
// from the X-Request-Id header when present, or generated, and echoed back on
// the response so clients and downstream services can correlate too.
func (s *Server) withLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)

		logger := s.Logger.With(
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
		)
		r = r.WithContext(contextWithLogger(r.Context(), logger))
		next.ServeHTTP(w, r)
	})
}

// newRequestID returns a random hex correlation ID for a request.
func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
