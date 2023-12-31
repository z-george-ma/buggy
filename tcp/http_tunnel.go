package tcp

import (
	"fmt"
)

type HttpTunnelConn struct {
	Conn
	Request HttpRequest
}

var proxyResponse []byte = []byte("HTTP/1.1 200 OK\r\n\r\n")

func HttpTunnelAccept(conn Conn) (ret *HttpTunnelConn, err error) {

	ret = &HttpTunnelConn{
		Conn: conn,
	}

	ret.Request, err = ParseHttpRequest(conn)
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
