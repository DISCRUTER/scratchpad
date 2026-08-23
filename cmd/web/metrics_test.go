package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/discruter/scratchpad/internal/assert"
	"github.com/discruter/scratchpad/internal/metrics"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// newTestMetrics builds an isolated metrics instance wired to its own
// registry, plus a mux exposing the /metrics scrape endpoint. It mirrors
// internal/metrics.GetMetrics while keeping the registry reachable for
// assertions.
func newTestMetrics(t *testing.T) (*metrics.Metrics, *http.ServeMux) {
	t.Helper()

	reg := prometheus.NewRegistry()
	m := &metrics.Metrics{
		RequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "scratchpad",
			Name:      "http_request_total",
			Help:      "Total number of http request.",
		}, []string{"method", "endpoint", "status_code"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "scratchpad",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
		}, []string{"method", "endpoint"}),
		ActiveRequests: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "scratchpad",
			Name:      "http_active_requests",
			Help:      "Number of active HTTP requests",
		}),
	}
	reg.MustRegister(m.RequestCount, m.RequestDuration, m.ActiveRequests)

	pMux := http.NewServeMux()
	pMux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	return m, pMux
}

func scrapeMetrics(t *testing.T, pMux *http.ServeMux) string {
	t.Helper()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	pMux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("scrape failed: got status %d", w.Code)
	}

	return w.Body.String()
}

func TestStatusRecorder(t *testing.T) {
	t.Run("Captures explicit status code and delegates headers/body", func(t *testing.T) {
		rr := httptest.NewRecorder()
		recorder := &statusRecorder{ResponseWriter: rr, status: http.StatusOK}

		recorder.Header().Set("X-Test", "yes")
		recorder.WriteHeader(http.StatusTeapot)
		recorder.Write([]byte("body"))

		assert.Equal(t, recorder.status, http.StatusTeapot)
		assert.Equal(t, rr.Header().Get("X-Test"), "yes")
		assert.Equal(t, rr.Code, http.StatusTeapot)
		assert.Equal(t, rr.Body.String(), "body")
	})

	t.Run("Defaults to 200 when only Write is called", func(t *testing.T) {
		rr := httptest.NewRecorder()
		recorder := &statusRecorder{ResponseWriter: rr, status: http.StatusOK}

		recorder.Write([]byte("implicit ok"))

		assert.Equal(t, recorder.status, http.StatusOK)
		assert.Equal(t, rr.Code, http.StatusOK)
	})

	t.Run("First WriteHeader wins", func(t *testing.T) {
		rr := httptest.NewRecorder()
		recorder := &statusRecorder{ResponseWriter: rr, status: http.StatusOK}

		recorder.WriteHeader(http.StatusOK)
		recorder.WriteHeader(http.StatusInternalServerError)

		assert.Equal(t, recorder.status, http.StatusOK)
	})

	t.Run("Unwrap exposes the underlying ResponseWriter", func(t *testing.T) {
		rr := httptest.NewRecorder()
		recorder := &statusRecorder{ResponseWriter: rr, status: http.StatusOK}

		unwrapped, ok := recorder.Unwrap().(*httptest.ResponseRecorder)
		if !ok {
			t.Fatal("Unwrap did not return the wrapped writer")
		}
		assert.Equal(t, unwrapped, rr)
	})
}

func TestRequestRoutePattern(t *testing.T) {
	t.Run("Returns the matched ServeMux pattern", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/pads/view/1", nil)
		r.Pattern = "GET /pads/view/{id}"

		assert.Equal(t, requestRoutePattern(r), "GET /pads/view/{id}")
	})

	t.Run("Falls back to unmatched for unrouted requests", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/nowhere", nil)

		assert.Equal(t, requestRoutePattern(r), "unmatched")
	})
}

func TestTrackMetricsRecordsStatusAndRoutePattern(t *testing.T) {
	app := newTestApplication(t)
	m, pMux := newTestMetrics(t)
	app.metrics = m

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name         string
		method       string
		urlPath      string
		wantCode     int
		wantEndpoint string
	}{
		{
			name:         "Rendered page records pattern label",
			method:       http.MethodGet,
			urlPath:      "/pads/view/1",
			wantCode:     http.StatusOK,
			wantEndpoint: "GET /pads/view/{id}",
		},
		{
			name:         "Unmatched route falls back to unmatched label",
			method:       http.MethodGet,
			urlPath:      "/does/not/exist",
			wantCode:     http.StatusNotFound,
			wantEndpoint: "unmatched",
		},
		{
			name:         "Redirect from protected route is recorded",
			method:       http.MethodGet,
			urlPath:      "/pads/create",
			wantCode:     http.StatusSeeOther,
			wantEndpoint: "GET /pads/create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, ts.URL+tt.urlPath, nil)
			if err != nil {
				t.Fatal(err)
			}
			rs, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer rs.Body.Close()
			io.Copy(io.Discard, rs.Body)

			assert.Equal(t, rs.StatusCode, tt.wantCode)
		})
	}

	body := scrapeMetrics(t, pMux)

	for _, want := range []string{
		`scratchpad_http_request_total{endpoint="GET /pads/view/{id}",method="GET",status_code="200"} 1`,
		`scratchpad_http_request_total{endpoint="GET /pads/create",method="GET",status_code="303"} 1`,
		`scratchpad_http_request_total{endpoint="unmatched",method="GET",status_code="404"} 1`,
		`scratchpad_http_request_duration_seconds_count{endpoint="GET /pads/view/{id}",method="GET"} 1`,
		`scratchpad_http_active_requests 0`,
	} {
		assert.StringContains(t, body, want)
	}

	// Raw paths must never leak into labels (cardinality safety).
	if strings.Contains(body, `/pads/view/1"`) || strings.Contains(body, `/does/not/exist"`) {
		t.Errorf("raw request path leaked into metric labels:\n%s", body)
	}
}

func TestTrackMetricsRecoversFromPanic(t *testing.T) {
	app := newTestApplication(t)
	m, pMux := newTestMetrics(t)
	app.metrics = m

	panicMux := http.NewServeMux()
	panicMux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	ts := httptest.NewServer(alice.New(app.trackMetrics, app.recoverPanic).Then(panicMux))
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/boom")
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rs.StatusCode, http.StatusInternalServerError)
	assert.StringContains(t, string(body), "Internal Server Error")

	metricsBody := scrapeMetrics(t, pMux)

	for _, want := range []string{
		`scratchpad_http_request_total{endpoint="GET /boom",method="GET",status_code="500"} 1`,
		`scratchpad_http_active_requests 0`,
	} {
		assert.StringContains(t, metricsBody, want)
	}
}

func TestTrackMetricsConcurrentRequestsReturnGaugeToZero(t *testing.T) {
	app := newTestApplication(t)
	m, pMux := newTestMetrics(t)
	app.metrics = m

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	const workers = 25
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			rs, err := ts.Client().Get(ts.URL + "/ping")
			if err != nil {
				t.Error(err)
				return
			}
			defer rs.Body.Close()
			io.Copy(io.Discard, rs.Body)
			if rs.StatusCode != http.StatusOK {
				t.Errorf("/ping: got status %d, want %d", rs.StatusCode, http.StatusOK)
			}
		}()
	}
	wg.Wait()

	gauge := scrapeMetrics(t, pMux)
	assert.StringContains(t, gauge, `scratchpad_http_active_requests 0`)
	assert.StringContains(t, gauge, `scratchpad_http_request_total{endpoint="GET /ping",method="GET",status_code="200"} 25`)
}
