package server

import (
	"encoding/json"
	"net/http"
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
		r.URL.Path = strings.Replace(r.URL.Path, "//", "/", -1)

		next.ServeHTTP(w, r)
	})
}

// schemeSetter sets the request's scheme, if possible.
func schemeSetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Scheme == "" {
			r.URL.Scheme = "http"
		}

		// TODO Use a trusted proxy list instead
		if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
			r.URL.Scheme = forwardedProto
		}

		next.ServeHTTP(w, r)
	})
}

// httpHealth is the handler for the "/health" endpoint.
func (s *Server) httpHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(healthStatus{
		Uptime:            uint64(time.Now().Sub(s.startTime) / time.Second),
		Connections:       atomic.LoadInt64(&s.connected),
		Address:           r.URL.Scheme + "://" + r.Host,
		Name:              s.Name,
		SupportedVersions: s.supportedVersions,
	})
}
