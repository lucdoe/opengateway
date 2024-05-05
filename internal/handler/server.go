package handler

import "net/http"

type ServerI interface {
	ListenAndServe(addr string, handler http.Handler) error
}

type Server struct{}

func (s *Server) ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}

func APIGatewayServer() ServerI {
	return &Server{}
}
