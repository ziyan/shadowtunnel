package server

import (
	"io"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"

	"github.com/ziyan/shadowtunnel/secure"
)

var log = logging.MustGetLogger("server")

type Server struct {
	password []byte
	connect  string
	timeout  time.Duration

	listener net.Listener
}

func NewServer(password []byte, listen, connect string, timeout time.Duration) (*Server, error) {

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}

	s := &Server{
		password: password,
		connect:  connect,
		listener: listener,
		timeout:  timeout,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) listen() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Warningf("failed to accept tcp connection: %s", err)
			break
		}

		go s.accept(conn)
	}
}

func (s *Server) accept(conn net.Conn) {
	session, err := yamux.Server(secure.NewEncryptedConnection(conn, s.password), nil)
	if err != nil {
		log.Errorf("failed to create server session: %s", err)
		conn.Close()
		return
	}
	defer session.Close()

	for {
		stream, err := session.AcceptStream()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warningf("failed to accept stream: %s", err)
			break
		}

		go func() {
			conn, err := net.DialTimeout("tcp", s.connect, s.timeout)
			if err != nil {
				log.Warningf("failed to connect to remote server %s: %s", s.connect, err)
				stream.Close()
				return
			}

			go func() {
				defer conn.Close()
				defer stream.Close()
				io.Copy(conn, stream)
			}()

			go func() {
				defer conn.Close()
				defer stream.Close()
				io.Copy(stream, conn)
			}()
		}()
	}
}
