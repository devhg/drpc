package drpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devhg/drpc/codec"
	"log"
	"net"
	"sync"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

func (c *Call) done() {
	c.Done <- c
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously(同时的).
// 英语真是个好东西！！！
type Client struct {
	cc     codec.Codec
	header codec.Header
	opt    *Option

	sending sync.Mutex // 防止多个请求报文的混乱，保证一次请求发送是原子的
	mu      sync.Mutex

	seq     uint64
	pending map[uint64]*Call

	closing  bool
	shutdown bool
}

var ErrShutdown = errors.New("connection is shut down")

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		err := fmt.Errorf("invalid codec type: %s", opt.CodecType)
		log.Println("rpc client: invalid codec type:", opt.CodecType)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println()
		_ = conn.Close()
		return nil, err
	}
	return newClientWithCodec(codecFunc(conn), opt), nil
}

func newClientWithCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:      cc,
		opt:     opt,
		seq:     1,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close()
}

// IsAvailable determine whether the Client is reachable
func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[c.seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) (call *Call) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return nil
	}
	call = c.pending[seq]
	delete(c.pending, seq)
	return
}

// 当服务端或者客户端发生错误的时候，终止队列中的所有Call
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err := c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = errors.New(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("rpc client: reading body: " + err.Error())
			}
			call.done()
		}
	}
	// errors occurs, so terminateCalls pending calls.
	c.terminateCalls(err)
}

// Call 是客户端暴露给用户的RPC服务调用接口，它是对 Go 的封装。
// 阻塞等待call.Done()，等待响应返回，是一个同步接口
func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	done := c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	call := <-done
	return call.Error
}

// Go 是客户端暴露给用户的RPC服务调用接口，与 Call 不同的是，
// Go 是一个异步接口，它返回一个Call实例
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}

func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// prepare request header
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	// encode and send the request
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		// call may be is nil, it usually means that Write method
		// partially failed, client has received the response and handled
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func Dial(network, addr string, opts ...*Option) (client *Client, err error) {
	option, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	// close the connection if client is nil
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, option)
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) > 1 {
		return nil, errors.New("number of option is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}
