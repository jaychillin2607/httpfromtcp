package server

import (
	"fmt"
	"log"
	"net"

	"www.github.com/jaychillin2607/httpfromtcp/internal/request"
	"www.github.com/jaychillin2607/httpfromtcp/internal/response"
)

type Server struct {
	port     int
	closed   bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, h Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		closed:   false,
		listener: listener,
		handler:  h,
	}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}

func (s *Server) listen() {
	for {
		if s.closed {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	log.Println("received request:", conn)
	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		headers := response.GetDefaultHeaders(0)
		responseWriter.WriteStatusLine(response.BAD_REQUEST)
		responseWriter.WriteHeaders(headers)
		return
	}

	s.handler(responseWriter, r)

}
