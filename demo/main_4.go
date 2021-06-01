package main

//
//import (
//	"context"
//	"fmt"
//	"github.com/devhg/drpc"
//	"github.com/devhg/drpc/xclient"
//	"log"
//	"net"
//	"sync"
//	"time"
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
//func (f Foo) Sleep(args Args, reply *int) error {
//	time.Sleep(time.Second * time.Duration(args.Num1))
//	*reply = args.Num1 + args.Num2
//	return nil
//}
//
//func startServer(addr chan string) {
//	var foo Foo
//	// pick a free port
//	listen, err := net.Listen("tcp", ":0")
//	if err != nil {
//		log.Fatal("network error:", err)
//	}
//	server := drpc.NewServer()
//	if err := server.Register(&foo); err != nil {
//		log.Fatal("register error: ", err)
//	}
//
//	log.Println("start rpc server on", listen.Addr())
//	addr <- listen.Addr().String()
//	server.Accept(listen)
//}
//
//func call(addr1, addr2 string) {
//	fmt.Println(addr1, addr2)
//	discovery := xclient.NewMultiServerDiscovery([]string{
//		"tcp@" + addr1,
//		"tcp@" + addr2,
//	})
//	client := xclient.NewXClient(discovery, xclient.RandomSelect, nil)
//	defer func() { _ = client.Close() }()
//
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := Args{i, i * i}
//			foo(context.Background(), client, "call", "Foo.Sum", &args)
//		}(i)
//	}
//	wg.Wait()
//}
//
//func broadcast(addr1, addr2 string) {
//	discovery := xclient.NewMultiServerDiscovery([]string{
//		"tcp@" + addr1,
//		"tcp@" + addr2,
//	})
//	client := xclient.NewXClient(discovery, xclient.RandomSelect, nil)
//	defer func() { _ = client.Close() }()
//
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := Args{i, i * i}
//			foo(context.Background(), client, "broadcast", "Foo.Sum", &args)
//
//			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
//			foo(ctx, client, "broadcast", "Foo.Sleep", &args)
//		}(i)
//	}
//	wg.Wait()
//}
//
//func foo(ctx context.Context, xc *xclient.XClient, typ, serviceMethod string, args *Args) {
//	var reply int
//	var err error
//	switch typ {
//	case "call":
//		err = xc.Call(ctx, serviceMethod, args, &reply)
//	case "broadcast":
//		err = xc.BroadCast(ctx, serviceMethod, args, &reply)
//	}
//	if err != nil {
//		log.Printf("%s %s error: %v", typ, serviceMethod, err)
//	} else {
//		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
//	}
//}
//
//func demo() {
//	ch1 := make(chan string)
//	ch2 := make(chan string)
//	go startServer(ch1)
//	go startServer(ch2)
//
//	addr1 := <-ch1
//	addr2 := <-ch2
//	call(addr1, addr2)
//	broadcast(addr1, addr2)
//}
