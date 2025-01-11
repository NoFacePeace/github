package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"
const message = "Hello, world!"

func main() {
	go func() {
		log.Fatal(echoServer())
	}()
	if err := clientMain(); err != nil {
		log.Fatal(err)
	}
}

func echoServer() error {
	// create listener
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		return fmt.Errorf("quic listen addr error: [%w]", err)
	}
	defer listener.Close()
	// create connection
	conn, err := listener.Accept(context.Background())
	if err != nil {
		return fmt.Errorf("listener accept error: [%w]", err)
	}
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return fmt.Errorf("connection accept stream error: [%w]", err)
	}
	defer stream.Close()
	if _, err := io.Copy(loggingWriter{stream}, stream); err != nil {
		return fmt.Errorf("io copy error: [%w]", err)
	}
	return nil
}

func clientMain() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(context.Background(), addr, tlsConf, nil)
	if err != nil {
		return fmt.Errorf("quic dial addr error: [%w]", err)
	}
	defer conn.CloseWithError(0, "")
	stream, err := conn.OpenStreamSync(conn.Context())
	if err != nil {
		return fmt.Errorf("connection open stream sync error: [%w]", err)
	}
	defer stream.Close()
	if _, err := stream.Write([]byte(message)); err != nil {
		return fmt.Errorf("stream write error: [%w]", err)
	}
	buf := make([]byte, len(message))
	if _, err := io.ReadFull(stream, buf); err != nil {
		return fmt.Errorf("io read full error: [%w]", err)
	}
	fmt.Printf("Client: Got '%s'\n", buf)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return fmt.Errorf("io read full error: [%w]", err)
	}
	fmt.Printf("Client: Got '%s'\n", buf)

	return nil
}

func generateTLSConfig() *tls.Config {
	// generate rsa key
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Fatal(err)
	}
	// generate certificate template
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	// create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatal(err)
	}
	// generate key
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	// generate cert
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	// create tls config
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}

type loggingWriter struct {
	io.Writer
}

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}
