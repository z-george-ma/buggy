package tcp

import (
	"net"
)

type TcpDialer struct {
	Dialer        *net.Dialer
	tcpNoDelay    bool
	readerBufSize int
	writerBufSize int
}

func NewDialer(tcpNoDelay bool, readerBufSize, writerBufSize int) *TcpDialer {
	return &TcpDialer{
		Dialer:        &net.Dialer{},
		tcpNoDelay:    tcpNoDelay,
		readerBufSize: readerBufSize,
		writerBufSize: writerBufSize,
	}
}

func (self *TcpDialer) Dial(address string) (*TcpConn, error) {
	conn, err := self.Dialer.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	tcpConn := conn.(*net.TCPConn)

	if !self.tcpNoDelay {
		tcpConn.SetNoDelay(self.tcpNoDelay)
	}

	return &TcpConn{
		Reader:  NewReader(conn, self.readerBufSize),
		Writer:  NewWriter(conn, self.writerBufSize),
		TCPConn: tcpConn,
	}, nil
}
