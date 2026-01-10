package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/jms-guy/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error starting tcp listener")
	}

	newServer := Server{Listener: listener}

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
	defaultHeaders := response.GetDefaultHeaders(0)
	err := response.WriteStatusLine(conn, response.Code200)
	if err != nil {
		log.Println(err)
	}
	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		log.Println(err)
	}
	conn.Close()
}
