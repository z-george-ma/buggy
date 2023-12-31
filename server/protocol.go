package main

import (
	"fmt"

	"github.com/z-george-ma/buggy/v2/tcp"
)

var connectResponse []byte = []byte("HTTP/1.1 200 OK\r\n\r\n")

func HandleConnection(dialer *tcp.TcpDialer, conn *tcp.TlsConn) (err error) {
	if err = conn.Conn.Handshake(); err != nil {
		return
	}

	request, err := tcp.ParseHttpRequest(conn)

	if request.Method != "CONNECT" {
		err = fmt.Errorf("Method %s not supported", request.Method)
		return
	}

	if _, err = conn.Write(connectResponse); err != nil {
		return
	}

	if err = conn.Flush(); err != nil {
		return
	}

	down, err := dialer.Dial(request.Url)
	if err != nil {
		return
	}

	defer down.Close()

	return tcp.Splice(conn, down)
}
