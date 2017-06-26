package client

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/op/go-logging"

	"github.com/ziyan/shadowtunnel/secure"
)

var log = logging.MustGetLogger("client")

type Client struct {
	connect  string
	password []byte
	timeout  time.Duration

	listener net.Listener
	mutex    sync.Mutex
	session  *yamux.Session
}

func NewClient(password []byte, listen, connect string, timeout time.Duration) (*Client, error) {

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}

	c := &Client{
		password: password,
		connect:  connect,
		timeout:  timeout,
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
	session, err := c.open()
	if err != nil {
		log.Errorf("failed to create client session: %s", err)
		conn.Close()
		return
	}

	stream, err := session.OpenStream()
	if err != nil {
		log.Errorf("failed to create client stream: %s", err)
		session.Close()
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

	session, err := yamux.Client(secure.NewEncryptedConnection(conn, c.password), nil)
	if err != nil {
		conn.Close()
		return nil, err
	}

	c.session = session
	return c.session, nil
}
