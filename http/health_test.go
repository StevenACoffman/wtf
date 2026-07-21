package http_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// TestServer_Health verifies the liveness and readiness probes.
func TestServer_Health(t *testing.T) {
	// Liveness always reports OK while the process is serving.
	t.Run("Healthz", func(t *testing.T) {
		s := MustOpenServer(t)
		defer MustCloseServer(t, s)

		resp, err := http.DefaultClient.Do(
			s.MustNewRequest(t, context.Background(), "GET", "/healthz", nil),
		)
		if err != nil {
			t.Fatal(err)
		} else if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Fatalf("StatusCode=%v, want %v", got, want)
		}
	})

	// With no HealthCheckFn configured, readiness reports OK.
	t.Run("ReadyzOK", func(t *testing.T) {
		s := MustOpenServer(t)
		defer MustCloseServer(t, s)

		resp, err := http.DefaultClient.Do(
			s.MustNewRequest(t, context.Background(), "GET", "/readyz", nil),
		)
		if err != nil {
			t.Fatal(err)
		} else if got, want := resp.StatusCode, http.StatusOK; got != want {
			t.Fatalf("StatusCode=%v, want %v", got, want)
		}
	})

	// A failing HealthCheckFn makes readiness report 503 Service Unavailable.
	t.Run("ReadyzUnavailable", func(t *testing.T) {
		s := MustOpenServer(t)
		defer MustCloseServer(t, s)
		s.HealthCheckFn = func(context.Context) error {
			return errors.New("database unavailable")
		}

		resp, err := http.DefaultClient.Do(
			s.MustNewRequest(t, context.Background(), "GET", "/readyz", nil),
		)
		if err != nil {
			t.Fatal(err)
		} else if got, want := resp.StatusCode, http.StatusServiceUnavailable; got != want {
			t.Fatalf("StatusCode=%v, want %v", got, want)
		}
	})
}
