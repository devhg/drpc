package drpc

import (
	"reflect"
	"sync/atomic"
)

// 通过反射实现service

// 手动封装的 rpc调用函数类型
type methodType struct {
	method   reflect.Method // 方法本身
	ArgType  reflect.Type   // 第一个参数 =>
	RetType  reflect.Type   // 第二个参数 =>
	numCalls uint64         // 统计函数调用次数（用于限流）
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value

	// 对于 arg = reflect.New(bb) 这种，因为reflect.New()默认创建的是指针格式
	// bb=*int ->  bb.Elem()=int           -> reflect.New(bb.Elem())=*int
	// bb=*int ->  reflect.New(bb)=**int
	//
	// bb=int  ->  reflect.New(bb)=*int    -> reflect.New(bb).Elem()=int
	// bb=int  ->  reflect.New(bb)=*int
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newRetv() reflect.Value {
	// retType must be a pointer type
	retV := reflect.New(m.RetType.Elem())

	switch m.RetType.Elem().Kind() {
	case reflect.Map:
		retV.Elem().Set(reflect.MakeMap(m.RetType.Elem()))
	case reflect.Slice:
		retV.Elem().Set(reflect.MakeSlice(m.RetType.Elem(), 0, 0))
	}
	return retV
}

// 一个服务，即一个结构体
type service struct {
	name   string                 // 结构体名称
	typ    reflect.Type           // 结构体类型
	rcvr   reflect.Value          // 结构体实例
	method map[string]*methodType // 服务中可能有多个函数
}

func newService(rcvr interface{}) *service {
	return &service{}
}

func (s *service) registerMethods() {

}

func (s *service) call() error {
	return nil
}
