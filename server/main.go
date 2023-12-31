package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"os/signal"
	"syscall"

	stdlog "log"

	"github.com/z-george-ma/buggy/v2/conf"
	"github.com/z-george-ma/buggy/v2/log"
	"github.com/z-george-ma/buggy/v2/tcp"
)

func main() {
	logger, err := log.NewLogger(func(err error, msg []byte) bool {
		stdlog.Output(1, err.Error())
		stdlog.Print(string(msg))
		return true
	})

	if err != nil {
		stdlog.Output(0, err.Error())
		return
	}

	defer logger.Close(context.Background())

	log := logger.With().Unit("buggy-server").Logger()

	config := conf.LoadConfig[Config]()

	sig, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM) // alloc

	server := tcp.NewServer(true, 8192, 0, func(err error) bool {
		if _, ok := err.(*net.OpError); ok {
			// accept deadline reached
			cancel()
			return true
		}

		log.Err().Error(1, err)
		return false
	})

	certPool := x509.NewCertPool()
	cert, err := os.ReadFile(config.ClientRootCA)
	if err != nil {
		log.Err().Error(0, err)
		return
	}
	certPool.AppendCertsFromPEM(cert)

	certs, err := tls.LoadX509KeyPair(config.ServerCert, config.ServerKey)
	if err != nil {
		log.Err().Error(0, err)
		return
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{certs},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}

	tcpDialer := tcp.NewDialer(true, 8192, 0)
	server.OnConnect(func(tc *tcp.TcpConn) {
		connLog := log.With().Value("client_ip", tc.TCPConn.RemoteAddr().String()).Logger()
		conn := tcp.TlsBind(tc, &tlsConfig)
		defer conn.Close()

		err := conn.Conn.Handshake()
		if err != nil {
			connLog.Err().Error(0, err)
			return
		}

		httpTunnel, err := tcp.HttpTunnelAccept(conn)
		if err != nil {
			connLog.Err().Error(0, err)
			return
		}

		down, err := tcpDialer.Dial(httpTunnel.Request.Url)
		if err != nil {
			connLog.Err().Error(0, err)
			return
		}

		defer down.Close()

		err = tcp.Splice(httpTunnel, down)
		if err != nil {
			connLog.Err().Error(0, err)
			return
		}
	})

	err = server.Start(context.Background(), "tcp", config.ListenAddr)
	if err != nil {
		log.Err().Error(0, err)
		return
	}
	defer server.Close(context.Background())

	<-sig.Done()
	log.Info().Msg("Exiting application")
}
