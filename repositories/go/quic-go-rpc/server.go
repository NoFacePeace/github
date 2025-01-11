package rpc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/rpc"

	"github.com/quic-go/quic-go"
	"golang.org/x/exp/slog"
)

func ListenAndServe(ctx context.Context, addr string, options ...ServerOption) error {
	config := &ServerConfig{
		tls: generateTLSConfig(),
	}
	for _, opt := range options {
		opt.apply(config)
	}
	listener, err := Listen(addr, config.tls)
	if err != nil {
		return fmt.Errorf("rpc listen error: [%w]", err)
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		go handleConnection(ctx, conn)
	}
}

type Server struct {
	*rpc.Server
}

func NewServer() *Server {
	return &Server{
		rpc.NewServer(),
	}
}

var DefaultServer = NewServer()

func Register(rcvr any) error {
	return DefaultServer.Register(rcvr)
}

func RegisterName(name string, rcvr any) error {
	return DefaultServer.RegisterName(name, rcvr)
}

func (s *Server) Register(rcvr any) error {
	return s.Server.Register(rcvr)
}

func (s *Server) RegisterName(name string, rcvr any) error {
	return s.Server.RegisterName(name, rcvr)
}

func (s *Server) ServeStream(stream *Stream) {
	s.Server.ServeConn(stream)
}

func (s *Server) ServeCodec(codec ServerCodec) {
	s.Server.ServeCodec(codec)
}

type ServerConfig struct {
	tls *tls.Config
}

type ServerOption interface {
	apply(*ServerConfig)
}

func Listen(addr string, tls *tls.Config) (*Listener, error) {
	listener, err := quic.ListenAddr(addr, tls, nil)
	if err != nil {
		return nil, fmt.Errorf("quic listen addr error: [%w]", err)
	}
	return &Listener{listener}, nil
}

type Listener struct {
	*quic.Listener
}

func (l *Listener) Accept(ctx context.Context) (*Connection, error) {
	conn, err := l.Listener.Accept(ctx)
	if err != nil {
		return nil, fmt.Errorf("quic listener accept error: [%w]", err)
	}
	return &Connection{conn}, nil
}

type ServerCodec interface {
	rpc.ServerCodec
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-go-rpc"},
	}
}

func handleConnection(ctx context.Context, conn *Connection) {
	defer conn.CloseWithError(0, "quic connection closed")
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		go handleStream(ctx, stream)
	}
}

func handleStream(ctx context.Context, stream *Stream) {
	defer stream.Close()
	select {
	case <-ctx.Done():
		return
	default:
		DefaultServer.ServeStream(stream)
		return
	}
}
