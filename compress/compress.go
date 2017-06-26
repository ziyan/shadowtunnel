package compress

import (
	"io"

	"github.com/golang/snappy"
)

type CompressedConnection struct {

	// underlying connection
	conn io.ReadWriteCloser

	// reader and writer
	reader *snappy.Reader
	writer *snappy.Writer
}

func (c *CompressedConnection) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

func (c *CompressedConnection) Write(b []byte) (int, error) {
	size, err := c.writer.Write(b)
	if err != nil {
		return 0, err
	}
	if err := c.writer.Flush(); err != nil {
		return 0, err
	}
	return size, nil
}

func (c *CompressedConnection) Close() error {
	return c.conn.Close()
}

func NewCompressedConnection(conn io.ReadWriteCloser) *CompressedConnection {
	return &CompressedConnection{
		conn:   conn,
		reader: snappy.NewReader(conn),
		writer: snappy.NewBufferedWriter(conn),
	}
}
