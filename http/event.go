package http

import (
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Websocket metrics.
var (
	websocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wtf_http_websocket_connections",
		Help: "Total number of connected websocket users",
	})
)

// registerEventRoutes is a helper function to register event routes.
func (s *Server) registerEventRoutes(r routeGroup) {
	r.handle("GET /events", s.handleEvents)
}

// handleEvents handles the "GET /events" route. This route provides real-time
// event notification over Websockets.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	websocketConnections.Inc()
	defer websocketConnections.Dec()

	// Upgrade HTTP connection to use websockets. Accept only allows same-origin
	// requests by default which prevents cross-site WebSocket hijacking.
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		LogError(r, err)
		return
	}
	// Ensure the connection is torn down when we exit this function. This can
	// occur if the HTTP request disconnects or if the subscription from the
	// event service closes.
	defer conn.CloseNow()

	// CloseRead reads & discards all incoming messages (which we don't care
	// about) so that control frames continue to be processed. It returns a
	// context, derived from the request's, that is canceled when the client
	// closes the connection or the request ends, which lets us detect
	// disconnects below.
	ctx := conn.CloseRead(r.Context())

	// Subscribe to all events for the current user.
	sub, err := s.EventService.Subscribe(ctx)
	if err != nil {
		LogError(r, err)
		return
	}
	defer sub.Close()

	// Stream all events to outgoing websocket writer.
	for {
		select {
		case <-ctx.Done():
			return // disconnect when the websocket connection disconnects

		case event, ok := <-sub.C():
			// If subscription is closed then exit.
			if !ok {
				return
			}

			// Marshal event data to JSON.
			buf, err := json.Marshal(event)
			if err != nil {
				LogError(r, err)
				return
			}

			// Write JSON data out to the websocket connection.
			if err := conn.Write(ctx, websocket.MessageText, buf); err != nil {
				LogError(r, err)
				return
			}
		}
	}
}
