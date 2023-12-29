package tcp

import "net"

func Connect(address string, tcpNoDelay bool, readerBufSize, writerBufSize int) (*TcpConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)

	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	if !tcpNoDelay {
		conn.SetNoDelay(tcpNoDelay)
	}

	return &TcpConn{
		Reader:  NewReader(conn, readerBufSize),
		Writer:  NewWriter(conn, writerBufSize),
		TCPConn: conn,
	}, nil
}
