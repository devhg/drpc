package main

// import (
// 	"context"
// 	"github.com/devhg/drpc"
// 	"log"
// 	"net"
// 	"net/http"
// 	"sync"
// 	"time"
// )

// type Foo int

// type Args struct{ Num1, Num2 int }

// func (f Foo) Sum(args Args, reply *int) error {
// 	*reply = args.Num1 + args.Num2
// 	return nil
// }

// func startServer(addr chan string) {
// 	//var foo Foo
// 	//l, _ := net.Listen("tcp", ":9999")
// 	//_ = drpc.Register(&foo)
// 	//drpc.HandleHTTP()
// 	//addr <- l.Addr().String()
// 	//_ = http.Serve(l, nil)

// 	var foo Foo

// 	// pick a free port
// 	listen, err := net.Listen("tcp", ":9990")
// 	if err != nil {
// 		log.Fatal("network error:", err)
// 	}

// 	if err := drpc.Register(&foo); err != nil {
// 		log.Fatal("register error: ", err)
// 	}

// 	drpc.HandleHTTP()
// 	log.Println("start rpc server on", listen.Addr())

// 	addr <- listen.Addr().String()
// 	_ = http.Serve(listen, nil)
// }

// func call(addr chan string) {
// 	client, _ := drpc.DialHTTP("tcp", <-addr)
// 	defer func() { _ = client.Close() }()

// 	time.Sleep(time.Second)
// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()

// 			args := Args{i, i * i}
// 			var reply int
// 			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
// 				log.Fatal("call Foo.Sum error:", err)
// 			}
// 			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }

// func demo() {
// 	addr := make(chan string)
// 	go call(addr)
// 	startServer(addr)
// }
