package xclient

import (
	"context"
	"io"
	"reflect"
	"sync"

	. "github.com/devhg/drpc"
)

// XClient is a client that supports load balance
type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *Option
	mu      sync.Mutex
	clients map[string]*Client
}

var _ io.Closer = (*XClient)(nil)

// NewXClient 的构造函数需要传入三个参数，
// * 服务发现实例 Discovery
// * 负载均衡模式 SelectMode
// * 协议选项 Option
// 为了尽量地复用已经创建好的 Socket 连接，使用 clients 保存创建成功的 Client 实例，
// 并提供 Close 方法。用于在结束后，关闭已经建立的所有连接
func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
// xc will choose a proper server.
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, ret interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(ctx, rpcAddr, serviceMethod, args, ret)
}

func (xc *XClient) call(ctx context.Context,
	rpcAddr, serviceMethod string, args, ret interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, ret)
}

// 实现 client 的复用能力
func (xc *XClient) dial(rpcAddr string) (*Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	// 先查缓存
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}

	// 创建新的客户端
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

// BroadCast invokes the named function for every server registered in discovery
func (xc *XClient) BroadCast(ctx context.Context, serviceMethod string, args, ret interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := ret == nil // if reply is nil, don't need to set value  妙阿。好活
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// todo 待研究
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if ret != nil {
				clonedReply = reflect.New(reflect.ValueOf(ret).Elem().Type()).Interface()
			}

			err := xc.call(ctx, rpcAddr, serviceMethod, args, clonedReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(ret).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
