package client

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"

	"github.com/ziyan/shadowtunnel/compress"
	"github.com/ziyan/shadowtunnel/secure"
)

var log = logging.MustGetLogger("client")

type Client struct {
	connect  string
	password []byte
	compress bool
	timeout  time.Duration

	listener net.Listener
	mutex    sync.Mutex
	session  *yamux.Session
}

func NewClient(password []byte, listen, connect string, compress bool, timeout time.Duration) (*Client, error) {

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}

	c := &Client{
		password: password,
		connect:  connect,
		timeout:  timeout,
		compress: compress,
		listener: listener,
	}

	go c.listen()

	return c, nil
}

func (c *Client) Close() {
	c.listener.Close()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.session != nil {
		c.session.Close()
		c.session = nil
	}
}

func (c *Client) listen() {
	defer c.listener.Close()

	for {
		conn, err := c.listener.Accept()
		if err != nil {
			log.Warningf("failed to accept tcp connection: %s", err)
			break
		}

		go c.accept(conn)
	}
}

func (c *Client) accept(conn net.Conn) {
	defer conn.Close()

	log.Infof("accepted connection in client mode from: %v", conn.RemoteAddr())

	session, err := c.open()
	if err != nil {
		log.Errorf("failed to create client session: %s", err)
		return
	}

	stream, err := session.OpenStream()
	if err != nil {
		log.Errorf("failed to create client stream: %s", err)
		session.Close()
		return
	}
	defer stream.Close()

	log.Infof("established stream %v in client mode: %v -> %v", stream.StreamID(), conn.RemoteAddr(), c.connect)

	done1 := make(chan struct{})
	go func() {
		io.Copy(conn, stream)
		close(done1)
	}()

	done2 := make(chan struct{})
	go func() {
		io.Copy(stream, conn)
		close(done2)
	}()

	select {
	case <-done1:
	case <-done2:
	}

	log.Infof("closing stream %v in client mode: %v -> %v", stream.StreamID(), conn.RemoteAddr(), c.connect)
}

func (c *Client) open() (*yamux.Session, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.session != nil && !c.session.IsClosed() {
		return c.session, nil
	}

	conn, err := net.DialTimeout("tcp", c.connect, c.timeout)
	if err != nil {
		return nil, err
	}

	var connection io.ReadWriteCloser
	connection = secure.NewEncryptedConnection(conn, c.password)
	if c.compress {
		connection = compress.NewCompressedConnection(connection)
	}

	session, err := yamux.Client(connection, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	c.session = session
	return c.session, nil
}
