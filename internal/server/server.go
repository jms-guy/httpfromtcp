package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/jms-guy/httpfromtcp/internal/request"
	"github.com/jms-guy/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	Handler  Handler
	isClosed atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error starting tcp listener")
	}

	newServer := Server{Listener: listener, Handler: handler}

	go newServer.listen()

	return &newServer, nil
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	if err != nil {
		return fmt.Errorf("error closing server tcp listener")
	}
	s.isClosed.Store(true)

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp := response.Writer{ResponseWriter: conn}

	s.Handler(&resp, req)
}
