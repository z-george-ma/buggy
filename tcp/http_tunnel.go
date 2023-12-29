package tcp

import (
	"fmt"
)

type HttpTunnelConn struct {
	Conn
	RequestHeader HttpRequestHeader
}

var proxyResponse []byte = []byte("HTTP/1.1 200 OK\r\n\r\n")

func HttpTunnelAccept(conn Conn) (ret *HttpTunnelConn, err error) {

	ret = &HttpTunnelConn{
		Conn: conn,
	}

	ret.RequestHeader, err = ParseHttpHeader(conn)
	if err != nil {
		return
	}

	if ret.RequestHeader.Method != "CONNECT" {
		err = fmt.Errorf("Method %s not supported", ret.RequestHeader.Method)
		return
	}

	_, err = conn.Write(proxyResponse)
	if err != nil {
		return
	}

	return ret, conn.Flush()
}
