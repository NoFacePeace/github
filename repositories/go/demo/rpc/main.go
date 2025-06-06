package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func main() {
	go server()
	client()
}

func server() {
	rpc.RegisterName("HelloService", new(HelloService))
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	rpc.ServeConn(conn)
	rpc.ServeCodec(jsonrpc.NewServerCodec(conn))

}

func client() {
	client, err := rpc.Dial("tcp", "localhost:1234")
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
