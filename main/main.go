package main

import (
	"encoding/json"
	"fmt"
	"github.com/devhg/drpc"
	"github.com/devhg/drpc/codec"
	"log"
	"net"
)

func startServer(addr chan string) {
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", listen.Addr())
	addr <- listen.Addr().String()
	drpc.Accept(listen)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer conn.Close()

	_ = json.NewEncoder(conn).Encode(drpc.DefaultOption)
	gobCodec := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.tom",
			Seq:           uint64(i),
		}
		_ = gobCodec.Write(h, fmt.Sprintf("drpc req %d", h.Seq))
		_ = gobCodec.ReadHeader(h)
		var reply string
		_ = gobCodec.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
