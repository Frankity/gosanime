package server

import (
	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func New() *Server {
	a := &Server{
		Router: mux.NewRouter(),
	}

	a.initRoutes()
	return a
}

func (a *Server) initRoutes() {
	a.Router.HandleFunc("/", a.IndexHandler()).Methods("GET")
	a.Router.HandleFunc("/api/v1/latestanimes", a.GetTopAnimes()).Methods("GET")
	a.Router.HandleFunc("/api/v1/ovas", a.GetOvas()).Methods("GET")
	a.Router.HandleFunc("/api/v1/anime", a.GetAnime()).Methods("GET")
	a.Router.HandleFunc("/api/v1/video", a.GetVideoServers()).Methods("GET")
	a.Router.HandleFunc("/api/v1/search", a.GetSearchAnime()).Methods("GET")
}
