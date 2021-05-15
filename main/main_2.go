package main

//
//import (
//	"context"
//	"github.com/devhg/drpc"
//	"log"
//	"net"
//	"sync"
//)
//
//type Foo int
//
//type Args struct{ Num1, Num2 int }
//
//func (f Foo) Sum(args Args, reply *int) error {
//	*reply = args.Num1 + args.Num2
//	return nil
//}
//
//func startServer(addr chan string) {
//	var foo Foo
//	if err := drpc.Register(&foo); err != nil {
//		log.Fatal("register error: ", err)
//	}
//
//	// pick a free port
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
//
//			args := Args{i, i * i}
//			var reply int
//			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
//				log.Fatal("call Foo.Sum error:", err)
//			}
//			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	wg.Wait()
//}
