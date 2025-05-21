package server

import (
	"github.com/gorilla/mux"
)

// Server struct holds the router instance.
// The router is responsible for matching incoming requests to their respective handlers.
type Server struct {
	Router *mux.Router // Router instance from gorilla/mux
}

// New creates and returns a new Server instance.
// It initializes the router and sets up the API routes.
func New() *Server {
	a := &Server{
		Router: mux.NewRouter(),
	}

	a.initRoutes()
	return a
}

// initRoutes defines all the API routes and maps them to their handler functions.
// This method is called during the server initialization.
func (a *Server) initRoutes() {
	a.Router.HandleFunc("/", a.IndexHandler()).Methods("GET")
	a.Router.HandleFunc("/api/v1/main", a.GetMain()).Methods("GET")
	a.Router.HandleFunc("/api/v1/ovas", a.GetOvas()).Methods("GET")
	a.Router.HandleFunc("/api/v1/anime", a.GetAnime()).Methods("GET")
	a.Router.HandleFunc("/api/v1/video", a.GetVideoServers()).Methods("GET")
	a.Router.HandleFunc("/api/v1/tags", a.GetTag()).Methods("GET")
	a.Router.HandleFunc("/api/v1/search", a.SearchAnime()).Methods("GET")
}
