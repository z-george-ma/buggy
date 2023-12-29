package tcp

import (
	"crypto/tls"
	"io"
	"net"
)

type TlsConn struct {
	Reader
	Writer
	*tls.Conn
}

func (self *TlsConn) Read(p []byte) (n int, err error) {
	return self.Reader.Read(p)
}

func (self *TlsConn) Write(p []byte) (n int, err error) {
	return self.Writer.Write(p)
}

func (self *TlsConn) Reset() error {
	tcpConn := self.Conn.NetConn().(*net.TCPConn)
	err := tcpConn.SetLinger(0)
	if err != nil {
		return err
	}
	return self.Conn.Close()
}

func (self *TlsConn) CloseWrite() error {
	err := self.Writer.Flush()
	if err != nil {
		return err
	}
	return self.Conn.CloseWrite()
}

func (self *TlsConn) Close() error {
	err := self.Writer.Flush()
	if err != nil {
		return err
	}
	return self.Conn.Close()
}

func (self *TlsConn) RawReader() io.Reader {
	return self.Conn
}

func (self *TlsConn) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(self.Conn, r)
}

func TlsConnect(conn *TcpConn, config *tls.Config) *TlsConn {
	c := tls.Client(conn.TCPConn, config)
	conn.Reader.(*ReaderImpl).Reset(c)
	conn.Writer.(*WriterImpl).Reset(c)
	return &TlsConn{
		Reader: conn.Reader,
		Writer: conn.Writer,
		Conn:   c,
	}
}

func TlsBind(conn *TcpConn, config *tls.Config) *TlsConn {
	c := tls.Server(conn.TCPConn, config)
	conn.Reader.(*ReaderImpl).Reset(c)
	conn.Writer.(*WriterImpl).Reset(c)
	return &TlsConn{
		Reader: conn.Reader,
		Writer: conn.Writer,
		Conn:   c,
	}
}
