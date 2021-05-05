package drpc

import (
	"encoding/json"
	"errors"
	"github.com/devhg/drpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

const MagicNumber = 0x3bef5c

// 报文形式
// | Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
// | <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|

// 在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文可能是这样的。
// | Option | Header1 | Body1 | Header2 | Body2 | ...

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
		}
		go server.ServeConn(conn)
	}
}

// Register publishes the receiver's methods in the DefaultServer
func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc server: service already defined: " + s.name)
	}
	return nil
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: decode Option error:", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x\n", opt.MagicNumber)
		return
	}
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		log.Printf("rpc server: invalid codec type %s\n", opt.CodecType)
		return
	}
	cc := codecFunc(conn)
	server.ServeCodec(cc)
}

func (server *Server) findService(serviceMethod string) (servci *service, mTyp *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request illegal-formed: " + serviceMethod)
		return
	}

	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	serviceInter, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service: " + serviceName)
		return
	}

	servci = serviceInter.(*service)
	mTyp = servci.method[methodName]
	if mTyp == nil {
		err = errors.New("rpc server: can't find method: " + methodName)
		return
	}
	return
}

var invalidRequest = struct{}{}

func (server *Server) ServeCodec(cc codec.Codec) {
	//var sending *sync.Mutex // make sure to send a complete response
	//var wg *sync.WaitGroup

	// 不能使用上面的，因为这样传参会传输nil ！！！！本函数直接使用的话是延迟初始化，没有问题。
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h      *codec.Header // header of request
	argv   reflect.Value
	replyv reflect.Value

	mTyp   *methodType
	servci *service
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	header, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: header}
	req.servci, req.mTyp, err = server.findService(header.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mTyp.newArgv()
	req.replyv = req.mTyp.newRetv()

	// make sure that argvi is a pointer interface{},
	// because ReadBody needs a pointer as parameter
	argvInter := req.argv.Interface()
	if req.argv.Kind() != reflect.Ptr {
		argvInter = req.argv.Addr().Interface()
	}

	if err = cc.ReadBody(argvInter); err != nil {
		log.Println("rpc server: read body error:", err)
		return req, err
	}
	return req, nil
}

func (server *Server) handleRequest(cc codec.Codec, req *request,
	sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	err := req.servci.call(req.mTyp, req.argv, req.replyv)
	if err != nil {
		req.h.Error = err.Error()
		server.sendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header,
	body interface{}, sending *sync.Mutex) {
	// TODO: 并发问题，保证发送过程是原子的
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}
