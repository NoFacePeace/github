package main

import (
	"context"
	"log"

	rpc "github.com/NoFacePeace/github/repositories/go/project/quic-go-rpc"
)

func main() {
	go server()
	client()
}

func server() {
	rpc.RegisterName("HelloService", new(HelloService))
	rpc.ListenAndServe(context.Background(), ":1234")
}

func client() {
	client, err := rpc.NewClient(context.Background(), "localhost:1234")
	if err != nil {
		log.Fatal(err)
	}
	var reply string
	err = client.Call("HelloService.Hello", "world", &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received: %s", reply)
	client.Close()
}

type HelloService struct {
}

func (p *HelloService) Hello(req string, reply *string) error {
	*reply = "Hello, " + req
	return nil
}
