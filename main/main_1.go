package main

//
//import (
//	"fmt"
//	"github.com/devhg/drpc"
//	"log"
//	"net"
//	"sync"
//)
//
//func startServer(addr chan string) {
//	listen, err := net.Listen("tcp", ":0")
//	if err != nil {
//		log.Fatal("network error:", err)
//	}
//	log.Println("start rpc server on", listen.Addr())
//	addr <- listen.Addr().String()
//	drpc.Accept(listen)
//}
//
//func main() {
//	addr := make(chan string)
//	go startServer(addr)
//
//	client, _ := drpc.Dial("tcp", <-addr)
//	defer func() { _ = client.Close() }()
//
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := fmt.Sprintf("drpc req %d", i)
//			var reply string
//			if err := client.Call("Foo.Sum", args, &reply); err != nil {
//				log.Fatal("call Foo.Sum error:", err)
//			}
//			log.Println("reply:", reply)
//		}(i)
//		//h := &codec.Header{
//		//	ServiceMethod: "Foo.tom",
//		//	Seq:           uint64(i),
//		//}
//		//_ = gobCodec.Write(h, fmt.Sprintf("drpc req %d", h.Seq))
//		//_ = gobCodec.ReadHeader(h)
//		//var reply string
//		//_ = gobCodec.ReadBody(&reply)
//		//log.Println("reply:", reply)
//	}
//	wg.Wait()
//}
