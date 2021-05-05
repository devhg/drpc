package drpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

// 通过反射实现service

// 手动封装的 rpc调用函数类型
type methodType struct {
	method   reflect.Method // 方法本身 func Foo(r *xxx.Request, resp *xxx.Response) error {}
	ArgType  reflect.Type   // 第一个参数 => *xxx.Request
	RetType  reflect.Type   // 第二个参数 => *xxx.Response
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
	name string       // 结构体名称
	typ  reflect.Type // 结构体类型

	// 结构体实例
	// 保留 receiver 是因为在调用时需要 receiver 作为第 0 个参数
	receiver reflect.Value
	method   map[string]*methodType // 服务中可能有多个方法
}

func newService(rcvr interface{}) *service {
	s := new(service)
	s.receiver = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.receiver).Type().Name()
	s.typ = reflect.TypeOf(rcvr)

	// 判断了类型是否为导出类型
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

// registerMethods 过滤出了符合条件的方法：
// 两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身，类似于 python 的 self，java 中的 this）
// 返回值有且只有 1 个，类型为 error
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)

		mType := method.Type

		// 只有三个参数，Foo(*self, *in, *out) error
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		argType, retType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(retType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:  method,
			ArgType: argType,
			RetType: retType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(mTyp *methodType, argV, retV reflect.Value) error {
	atomic.AddUint64(&mTyp.numCalls, 1)
	f := mTyp.method.Func

	// 反射调用函数
	returnValues := f.Call([]reflect.Value{s.receiver, argV, retV})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
