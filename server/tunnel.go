package main

import (
	"fmt"

	"github.com/z-george-ma/buggy/v2/tcp"
)

type HttpTunnelConn struct {
	tcp.Conn
	Request tcp.HttpRequest
}

var proxyResponse []byte = []byte("HTTP/1.1 200 OK\r\n\r\n")

func HttpTunnelAccept(conn tcp.Conn) (ret *HttpTunnelConn, err error) {

	ret = &HttpTunnelConn{
		Conn: conn,
	}

	ret.Request, err = tcp.ParseHttpRequest(conn)
	if err != nil {
		return
	}

	if ret.Request.Method != "CONNECT" {
		err = fmt.Errorf("Method %s not supported", ret.Request.Method)
		return
	}

	_, err = conn.Write(proxyResponse)
	if err != nil {
		return
	}

	return ret, conn.Flush()
}
