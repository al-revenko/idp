package apphttp

import (
	"net/http"
)

type Server struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

type handler struct {
	fn func(http.ResponseWriter, *http.Request)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.fn(w, r)
}

func NewServer(middlewares ...func(http.Handler) http.Handler) *Server {
	return &Server{
		mux:         http.NewServeMux(),
		middlewares: middlewares,
	}
}

func (s *Server) HandleFunc(path string, handlerFn func(http.ResponseWriter, *http.Request)) {
	s.Handle(path, handler{fn: handlerFn})
}

func (s *Server) Handle(path string, handler http.Handler) {
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}

	s.mux.Handle(path, finalHandler)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
