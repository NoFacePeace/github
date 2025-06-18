package rpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/quic-go/quic-go"
)

type Connection struct {
	quic.Connection
}

func Dial(ctx context.Context, addr string, tls *tls.Config) (*Connection, error) {
	conn, err := quic.DialAddr(ctx, addr, tls, nil)
	if err != nil {
		return nil, fmt.Errorf("quic dial addr error: [%w]", err)
	}
	return &Connection{conn}, nil
}

func (conn *Connection) OpenStream(ctx context.Context, bind bool) (*Stream, error) {
	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection open stream sync error: [%w]", err)
	}
	if bind {
		return &Stream{conn: conn, Stream: stream}, nil
	}
	return &Stream{Stream: stream}, nil
}

func (conn *Connection) AcceptStream(ctx context.Context) (*Stream, error) {
	stream, err := conn.Connection.AcceptStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("rpc connection accept stream error: [%w]", err)
	}
	return &Stream{Stream: stream}, nil
}

type Stream struct {
	conn *Connection
	quic.Stream
}

func (stream *Stream) Close() error {
	err := stream.Stream.Close()
	if stream.conn != nil {
		err = errors.Join(err, stream.conn.CloseWithError(0, ""))
	}
	return err
}
