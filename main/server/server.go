package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"xyz.frankity/gosanime/main/logger"
)

type Server struct {
	Router *mux.Router
}

func New() *Server {
	a := &Server{
		Router: mux.NewRouter(),
	}
	a.Router.Use(loggingMiddleware)
	a.initRoutes()
	return a
}

func (a *Server) initRoutes() {
	a.Router.HandleFunc("/", a.IndexHandler()).Methods("GET")
	a.Router.HandleFunc("/api/v1/main", a.GetMain()).Methods("GET")
	a.Router.HandleFunc("/api/v1/ovas", a.GetOvas()).Methods("GET")
	a.Router.HandleFunc("/api/v1/anime", a.GetAnime()).Methods("GET")
	a.Router.HandleFunc("/api/v1/video", a.GetVideoServers()).Methods("GET")
	a.Router.HandleFunc("/api/v1/tags", a.GetTag()).Methods("GET")
	a.Router.HandleFunc("/api/v1/search", a.SearchAnime()).Methods("GET")
}

// statusRecorder wraps ResponseWriter to capture the status code written by handlers.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		query := r.URL.RawQuery
		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
			"remote", r.RemoteAddr,
		}
		if query != "" {
			args = append(args, "query", query)
		}
		logger.L.Info("request", args...)
	})
}
