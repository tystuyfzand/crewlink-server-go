package server

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type stats struct {
	Address   string
	Connected int64
}

type healthStatus struct {
	Uptime            uint64   `json:"uptime"`
	Connections       int64    `json:"connectionCount"`
	Address           string   `json:"address"`
	Name              string   `json:"name"`
	SupportedVersions []string `json:"supportedVersions"`
}

// urlCleaner resolves an issue where CrewLink double slashes the file path
func urlCleaner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())

		routePath := rctx.RoutePath
		if routePath == "" {
			if r.URL.RawPath != "" {
				routePath = r.URL.RawPath
			} else {
				routePath = r.URL.Path
			}
			rctx.RoutePath = strings.Replace(routePath, "//", "/", -1)
		}

		next.ServeHTTP(w, r)
	})
}

// schemeSetter sets the request's scheme, if possible.
func schemeSetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "" {
			r.URL.Scheme = "http"

			if r.TLS != nil {
				r.URL.Scheme = "https"
			}
		}

		next.ServeHTTP(w, r)
	})
}

// httpHealth is the handler for the "/health" endpoint.
func (s *Server) httpHealth(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(healthStatus{
		Uptime:            uint64(time.Now().Sub(s.startTime) / time.Second),
		Connections:       atomic.LoadInt64(&s.connected),
		Address:           r.URL.Scheme + "://" + r.Host,
		Name:              s.Name,
		SupportedVersions: s.supportedVersions,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))

	w.Write(b)
}
