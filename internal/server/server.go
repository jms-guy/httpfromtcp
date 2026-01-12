package server

import (
	"bytes"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/jms-guy/httpfromtcp/internal/request"
	"github.com/jms-guy/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error starting tcp listener")
	}

	newServer := Server{Listener: listener}

	go newServer.listen(handler)

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

func (s *Server) listen(handler Handler) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			continue
		}

		go s.handle(handler, conn)
	}
}

func (s *Server) handle(handler Handler, conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	var buf bytes.Buffer
	hErr := handler(&buf, req)

	if hErr != nil {
		err = response.WriteStatusLine(conn, hErr.StatusCode)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = response.WriteHeaders(conn, response.GetDefaultHeaders(len(hErr.Msg)))
		if err != nil {
			fmt.Println(err)
			return
		}
		conn.Write([]byte(hErr.Msg))
		return
	} else {
		err = response.WriteStatusLine(conn, response.Code200)
	}
	headers := response.GetDefaultHeaders(buf.Len())
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write(buf.Bytes())
}
