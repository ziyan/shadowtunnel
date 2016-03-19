package secure

import (
	"crypto/cipher"
	"net"

	"github.com/golang/glog"
)

type EncryptedConnection struct {

	// underlying connection
	conn net.Conn

	// password for connection
	password []byte

	// encryption and decryption
	encrypter cipher.Stream
	decrypter cipher.Stream

	// persisted error
	err error
}

func NewEncryptedConnection(conn net.Conn, password []byte) *EncryptedConnection {
	return &EncryptedConnection{
		conn:     conn,
		password: password,
	}
}

func (c *EncryptedConnection) Write(b []byte) (n int, err error) {

	// persisted error
	if c.err != nil {
		return 0, err
	}

	// write header
	if c.encrypter == nil {
		encrypter, err := SendHandshake(c.conn, c.password)
		if err != nil {
			c.err = err
			return 0, err
		}
		c.encrypter = encrypter
	}

	glog.V(9).Infof("encrypting and sending to %v: %v", c.RemoteAddr(), b)

	// encrypt
	c.encrypter.XORKeyStream(b, b)

	// normal write
	size, err := c.conn.Write(b)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (c *EncryptedConnection) Read(b []byte) (n int, err error) {

	// persisted error
	if c.err != nil {
		return 0, err
	}

	// read header
	if c.decrypter == nil {
		decrypter, err := ReceiveHandshake(c.conn, c.password)
		if err != nil {
			c.err = err
			return 0, err
		}
		c.decrypter = decrypter
	}

	// normal read
	size, err := c.conn.Read(b)
	if err != nil {
		return 0, err
	}

	// decrypt
	c.decrypter.XORKeyStream(b[:size], b[:size])

	glog.V(9).Infof("received and decrypted from %v: %v", c.RemoteAddr(), b[:size])
	return size, nil
}

func (c *EncryptedConnection) Close() error {
	glog.V(9).Infof("closing encrypted connection to %v", c.RemoteAddr())
	return c.conn.Close()
}

func (c *EncryptedConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *EncryptedConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
