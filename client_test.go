package drpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

//func assert(condition bool, msg string, v ...interface{}) {
//	if !condition {
//		panic(fmt.Sprintf("assertion failed: "+msg, v...))
//	}
//}

// 用于测试连接超时。NewClient 函数耗时 2s，ConnectionTimeout 分别设置为 1s 和 0 两种场景。
func TestClient_dialTimeout(t *testing.T) {
	t.Parallel()
	listen, _ := net.Listen("tcp", ":0")

	f := func(conn net.Conn, opt *Option) (client *Client, err error) {
		_ = conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}

	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", listen.Addr().String(), &Option{
			ConnectTimeout: time.Second,
		})
		fmt.Println(err)
		assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "expect a timeout error")
	})

	t.Run("0", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", listen.Addr().String(), &Option{ConnectTimeout: 0})
		assert(err == nil, "0 means no limit")
	})
}

// 用于测试处理超时。 Bar.Timeout 耗时 2s，
//场景一：客户端设置超时时间为 1s，服务端无限制；场景二，服务端设置超时时间为1s，客户端无限制。
type Bar int

func (b Bar) Timeout(argv int, reply *int) error {
	time.Sleep(time.Second * 2)
	return nil
}

func startServer(addr chan string) {
	var b Bar
	_ = Register(&b)
	// pick a free port
	l, _ := net.Listen("tcp", ":0")
	addr <- l.Addr().String()
	Accept(l)
}

func TestClient_Call(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	go startServer(addrCh)

	addr := <-addrCh
	time.Sleep(time.Second)
	t.Run("client timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr)
		ctx, _ := context.WithTimeout(context.Background(), time.Second)

		var reply int
		err := client.Call(ctx, "Bar.Timeout", 1, &reply)
		assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect a timeout error")
	})

	t.Run("server timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr, &Option{HandleTimeout: time.Second})

		var reply int
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		fmt.Print(err)
		assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")
	})
}
