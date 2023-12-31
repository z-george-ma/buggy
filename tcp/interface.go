package tcp

import (
	"io"
)

type Reader interface {
	// read once, up to len(p) bytes into p
	io.Reader
	// ReadFull read up to full length of input slice
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

type CopyResult struct {
	Len int64
	Err error
	Dst io.ReaderFrom
}

func Copy(d io.ReaderFrom, s io.Reader, result chan CopyResult) {
	n, err := d.ReadFrom(s)
	result <- CopyResult{
		Len: n,
		Err: err,
		Dst: d,
	}
}

func Splice(dst Conn, src Conn) error {
	ret := make(chan CopyResult, 2)
	go Copy(dst, src.RawReader(), ret)
	go Copy(src, dst.RawReader(), ret)

	result := <-ret
	result.Dst.(Conn).CloseWrite()

	if result.Err != nil {
		return result.Err
	}

	result = <-ret
	result.Dst.(Conn).CloseWrite()

	return result.Err
}
