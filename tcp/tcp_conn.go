package tcp

import (
	"io"
	"net"
)

type TcpConn struct {
	Reader
	Writer
	*net.TCPConn
}

func (self *TcpConn) Reset() error {
	err := self.TCPConn.SetLinger(0)
	if err != nil {
		return err
	}

	return self.TCPConn.Close()
}

func (self *TcpConn) CloseWrite() error {
	err := self.Writer.Flush()
	if err != nil {
		return err
	}
	return self.TCPConn.CloseWrite()
}

func (self *TcpConn) Close() error {
	err := self.Writer.Flush()
	if err != nil {
		return err
	}
	return self.TCPConn.Close()
}

func (self *TcpConn) RawReader() io.Reader {
	return self.TCPConn
}
