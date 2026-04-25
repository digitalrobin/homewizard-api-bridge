package main

import (
	"net/http"
	"sort"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type server struct {
	client  *homeWizardClient
	metrics map[string]metricRoute
}

func newServer(client *homeWizardClient) *server {
	return &server{
		client:  client,
		metrics: metricRouteMap(),
	}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/docs.json", s.handleDocsJSON)
	mux.HandleFunc("/ui", s.handleUI)
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/auth/status", s.handleAuthStatus)
	mux.HandleFunc("/pair", s.handlePair)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/api/device", s.handleDevice)
	mux.HandleFunc("/api/measurement", s.handleMeasurement)
	mux.HandleFunc("/api/telegram", s.handleTelegram)

	for path := range s.metrics {
		mux.HandleFunc(path, s.handleMetric(path))
	}

	swagger := httpSwagger.Handler(
		httpSwagger.URL("/docs.json"),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DefaultModelsExpandDepth(httpSwagger.HideModel),
	)
	mux.Handle("/docs/", swagger)
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusTemporaryRedirect)
	})

	return loggingMiddleware(mux)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	paths := metricPaths(s.metrics)
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "homewizard-bridge",
		"routes": map[string]any{
			"system": []string{
				"GET /",
				"GET /docs",
				"GET /docs.json",
				"GET /ui",
				"GET /healthz",
				"GET /auth/status",
				"POST /pair",
				"GET /status",
			},
			"raw": []string{
				"GET /api/device",
				"GET /api/measurement",
				"GET /api/telegram",
			},
			"metrics": paths,
		},
	})
}

func (s *server) handleDocsJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buildOpenAPISpec(s.metrics))
}

func (s *server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(authPageHTML))
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"token_ready": s.client.tokenStore.Get() != "",
	})
}

func (s *server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	status := s.client.tokenStore.Status()
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"token_ready":   status.Token != "",
		"homewizard":    s.client.baseURL,
		"user":          s.client.userName,
		"token_updated": status.UpdatedAt,
		"ui":            "/ui",
	})
}

func (s *server) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	paired, status, err := s.client.Pair(r.Context())
	if err != nil {
		if status == http.StatusForbidden {
			writeJSON(w, http.StatusConflict, map[string]any{
				"ok":      false,
				"message": "Press the button on the HomeWizard P1 meter, then call POST /pair again within 30 seconds.",
				"error":   err.Error(),
			})
			return
		}
		writeProxyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"message":  "Pairing succeeded and token was stored locally.",
		"user":     paired.Name,
		"tokenSet": paired.Token != "",
	})
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	snapshot, err := s.client.Snapshot(r.Context())
	if err != nil {
		writeProxyError(w, err)
		return
	}

	values := make(map[string]any, len(s.metrics))
	unavailable := make([]string, 0)
	paths := metricPaths(s.metrics)
	for _, path := range paths {
		route := s.metrics[path]
		value, err := route.Resolver(snapshot)
		if err != nil {
			unavailable = append(unavailable, path)
			continue
		}
		values[path] = value
	}

	sort.Strings(unavailable)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"device":      snapshot.Device,
		"measurement": snapshot.Measurement,
		"utilities":   snapshot.Utilities,
		"values":      values,
		"missing":     unavailable,
	})
}

func (s *server) handleDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	data, err := s.client.GetDeviceInfo(r.Context())
	if err != nil {
		writeProxyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func (s *server) handleMeasurement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	data, err := s.client.GetMeasurement(r.Context())
	if err != nil {
		writeProxyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func (s *server) handleTelegram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	data, err := s.client.GetTelegram(r.Context())
	if err != nil {
		writeProxyError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(data))
}

func (s *server) handleMetric(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, http.MethodGet)
			return
		}

		route := s.metrics[path]
		snapshot, err := s.client.Snapshot(r.Context())
		if err != nil {
			writeProxyError(w, err)
			return
		}

		value, err := route.Resolver(snapshot)
		if err != nil {
			writeProxyError(w, err)
			return
		}

		writePlainValue(w, value)
	}
}
