package servertest_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/juanfont/headscale/hscontrol/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoiseRouterDoesNotServeMetrics pins that the router behind /ts2021 does
// not expose Prometheus metrics. The Noise handshake accepts any machine key,
// so every route on that router is reachable without credentials; metrics
// belong only on metrics_listen_addr. The /machine/whoami case is a control:
// its 501 proves the request reached the Noise router, so the 404 for
// /metrics comes from the missing route and not from a broken tunnel.
func TestNoiseRouterDoesNotServeMetrics(t *testing.T) {
	t.Parallel()

	h := servertest.NewHarness(t, 1)

	// Noise requests are addressed with the https scheme; the control client
	// routes them over the established Noise connection.
	noiseURL := strings.Replace(h.Server.URL, "http://", "https://", 1)

	tests := []struct {
		path string
		want int
	}{
		{"/machine/whoami", http.StatusNotImplemented},
		{"/metrics", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(
				ctx, http.MethodGet, noiseURL+tt.path, nil,
			)
			require.NoError(t, err)

			resp, err := h.Client(0).Direct().DoNoiseRequest(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tt.want, resp.StatusCode)
		})
	}
}
