package tcp

import (
	"io"
)

type Reader interface {
	// read once, up to len(p) bytes into p
	io.Reader
	// read up to full len(p) bytes into p
	ReadFull(p []byte) (n int, err error)
	// ReadAll reads until EOF. p is optional parameter to supply buffer
	ReadAll(p *[]byte) (b []byte, err error)
	// ReadSlice reads until delimiter. p is optional parameter to supply buffer
	ReadSlice(delim byte, p *[]byte) ([]byte, error)
	// set snapshot
	// seek
	// set deadline
}

type Writer interface {
	io.Writer
	// write full content
	WriteAll(p ...[]byte) (n int, err error)
	// flush
	Flush() error
}

type Conn interface {
	Reset() error
	CloseWrite() error
	Close() error

	Reader
	Writer

	RawReader() io.Reader
	io.ReaderFrom
}

func onewayCopy(s io.ReaderFrom, d io.Reader, result chan error) {
	_, err := s.ReadFrom(d)
	result <- err
}

func Pipe(dst Conn, src Conn) error {
	ret := make(chan error, 2)
	go onewayCopy(dst, src.RawReader(), ret)
	go onewayCopy(src, dst.RawReader(), ret)

	err := <-ret
	if err != nil {
		return err
	}
	return <-ret
}
