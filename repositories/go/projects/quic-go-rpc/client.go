package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/rpc"
)

type Client struct {
	rpc *rpc.Client
}

func NewClient(ctx context.Context, addr string, options ...ClientOption) (*Client, error) {
	config := &ClientConfig{
		tls: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-go-rpc"},
		},
	}
	for _, opt := range options {
		opt.apply(config)
	}
	conn, err := Dial(ctx, addr, config.tls)
	if err != nil {
		return nil, fmt.Errorf("rpc dial error: [%w]", err)
	}
	stream, err := conn.OpenStream(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("rpc connection new stream error: [%w]", err)
	}
	return NewClientWithStream(stream), nil
}

func NewClientWithStream(stream *Stream) *Client {
	return &Client{
		rpc: rpc.NewClient(stream),
	}
}

func NewClientWithCodec(codec ClientCodec) *Client {
	return &Client{
		rpc: rpc.NewClientWithCodec(codec),
	}
}

func (client *Client) Call(serviceMethod string, args any, reply any) error {
	return client.rpc.Call(serviceMethod, args, reply)
}

func (client *Client) Close() error {
	return client.rpc.Close()
}

type ClientCodec interface {
	rpc.ClientCodec
}

type ClientConfig struct {
	tls *tls.Config
}

type ClientOption interface {
	apply(*ClientConfig)
}
