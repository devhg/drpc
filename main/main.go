package main

import (
	"context"
	"github.com/devhg/drpc"
	"github.com/devhg/drpc/registry"
	"github.com/devhg/drpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func call(registry string) {
	discovery := xclient.NewDrpcRegistryDiscovery(registry, 0)
	client := xclient.NewXClient(discovery, xclient.RandomSelect, nil)
	defer func() { _ = client.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{i, i * i}
			foo(context.Background(), client, "call", "Foo.Sum", &args)
		}(i)
	}
	wg.Wait()
}

func broadcast(registry string) {
	discovery := xclient.NewDrpcRegistryDiscovery(registry, 0)
	client := xclient.NewXClient(discovery, xclient.RandomSelect, nil)
	defer func() { _ = client.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := Args{i, i * i}
			foo(context.Background(), client, "broadcast", "Foo.Sum", &args)

			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(ctx, client, "broadcast", "Foo.Sleep", &args)
		}(i)
	}
	wg.Wait()
}

func foo(ctx context.Context, xc *xclient.XClient, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.BroadCast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func startRegistry(wg *sync.WaitGroup) {
	listen, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(listen, nil)
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	listen, _ := net.Listen("tcp", ":0")
	server := drpc.NewServer()
	_ = server.Register(&foo)
	registry.Heartbeat(registryAddr, "tcp@"+listen.Addr().String(), 0)
	wg.Done()
	server.Accept(listen)
}

func main() {
	registryAddr := "http://localhost:9999/_drpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)
}
