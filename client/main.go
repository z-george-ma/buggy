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

	log := logger.With().Unit("buggy-client").Logger()

	config := conf.LoadConfig[Config]()

	serverAddr, err := tcp.UrlToAddress(config.RemoteUrl)

	if err != nil {
		log.Err().Error(0, err)
		return
	}

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

	var rootCAs *x509.CertPool
	if config.RootCA != "" {
		rootCAs = x509.NewCertPool()
		cert, err := os.ReadFile(config.RootCA)
		if err != nil {
			log.Err().Error(0, err)
			return
		}
		rootCAs.AppendCertsFromPEM(cert)
	}

	certs, err := tls.LoadX509KeyPair(config.ClientCert, config.ClientKey)
	if err != nil {
		log.Err().Error(0, err)
		return
	}

	tlsConfig := tls.Config{
		ServerName:   serverAddr.Host,
		Certificates: []tls.Certificate{certs},
		RootCAs:      rootCAs,
	}

	tcpDialer := tcp.NewDialer(true, 8192, 8192)

	server.OnConnect(func(tc *tcp.TcpConn) {
		defer tc.Close()
		connLog := log.With().Value("client_ip", tc.TCPConn.RemoteAddr().String()).Logger()
		down, err := tcpDialer.Dial(serverAddr.Address)

		if err != nil {
			connLog.Err().Error(0, err)
			return
		}

		defer down.Close()

		tls := tcp.TlsConnect(down, &tlsConfig)

		err = tcp.Splice(tc, tls)
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
