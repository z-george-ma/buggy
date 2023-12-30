package tcp

import (
	"context"
	"errors"
	"net"
	"time"
)

type TcpServer struct {
	ListenConfig  *net.ListenConfig
	tcpNoDelay    bool
	listener      *net.TCPListener
	onConnect     func(*TcpConn)
	onAcceptError func(error) bool
	readerBufSize int
	writerBufSize int
	loopCancel    context.CancelFunc
	loopEnded     chan struct{}
}

func NewServer(tcpNoDelay bool, readerBufSize, writerBufSize int, onAcceptError func(error) bool) *TcpServer {
	return &TcpServer{
		ListenConfig:  &net.ListenConfig{},
		tcpNoDelay:    tcpNoDelay,
		onAcceptError: onAcceptError,
		readerBufSize: readerBufSize,
		writerBufSize: writerBufSize,
		loopEnded:     make(chan struct{}),
	}
}

func (self *TcpServer) loop(ctx context.Context) {
	defer close(self.loopEnded)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := self.listener.AcceptTCP()
		if err != nil && self.onAcceptError(err) {
			return
		} else {
			if !self.tcpNoDelay {
				err := conn.SetNoDelay(self.tcpNoDelay)
				if err != nil && self.onAcceptError != nil {
					if self.onAcceptError(err) {
						return
					}
				}
			}

			go self.onConnect(&TcpConn{
				Reader:  NewReader(conn, self.readerBufSize),
				Writer:  NewWriter(conn, self.writerBufSize),
				TCPConn: conn,
			})
		}
	}
}

func (self *TcpServer) Close(ctx context.Context) {
	if self.loopCancel == nil {
		return
	}

	self.loopCancel()
	self.loopCancel = nil
	select {
	case <-ctx.Done():
	case <-self.loopEnded:
	}

	self.listener.Close()
}

func (self *TcpServer) Start(ctx context.Context, network string, address string) (err error) {
	if self.listener != nil {
		return errors.New("Another listener has already started")
	}

	listener, err := self.ListenConfig.Listen(ctx, network, address)
	if err != nil {
		return
	}

	self.listener = listener.(*net.TCPListener)
	loopCtx, cancel := context.WithCancel(context.Background())
	self.loopCancel = func() {
		self.listener.SetDeadline(time.Now())
		cancel()
	}

	go self.loop(loopCtx)
	return
}

func (self *TcpServer) OnConnect(connect func(*TcpConn)) {
	self.onConnect = connect
}
