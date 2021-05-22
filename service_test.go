package drpc

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestReflect(t *testing.T) {
	var wg sync.WaitGroup
	typ := reflect.TypeOf(&wg)
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		argv := make([]string, 0, method.Type.NumIn())
		rets := make([]string, 0, method.Type.NumOut())

		// 从 1 开始，第一个参数是wg自己
		for j := 1; j < method.Type.NumIn(); j++ {
			argv = append(argv, method.Type.In(j).Name())
		}

		for j := 0; j < method.Type.NumOut(); j++ {
			rets = append(rets, method.Type.Out(j).Name())
		}

		log.Printf("func (w *%s) %s(%s) %s",
			typ.Elem().Name(),
			method.Name,
			strings.Join(argv, ", "),
			strings.Join(rets, ", "),
		)
	}
}

func TestReflectNew(t *testing.T) {
	var a = 10
	var aa = reflect.TypeOf(a)
	//var bb = reflect.TypeOf(&a)
	var bb = reflect.TypeOf(a)
	fmt.Println(aa)
	fmt.Println(bb)

	var arg reflect.Value
	if bb.Kind() == reflect.Ptr {
		fmt.Println(bb)
		fmt.Println(bb.Elem())
		//arg = reflect.New(bb) 不能使用这种，因为reflect.New()默认创建的是指针格式
		arg = reflect.New(bb.Elem()) // bb=*int ->  bb.Elem()=int -> reflect.New(bb.Elem())=*int
	} else {
		arg = reflect.New(bb)
		fmt.Println(arg.Type())
		arg = arg.Elem()
		fmt.Println(arg.Type())
	}
	fmt.Println(arg.Type())
}

type User struct {
}

func (t *User) GetRetErrorType() error {
	return nil
}

func TestError(t *testing.T) {
	tt := &User{}
	of := reflect.TypeOf(tt)
	fmt.Println(of.NumMethod())
	fmt.Println(of)

	method := of.Method(0)
	mType := method.Type
	fmt.Println(mType.Out(0))
	//
	//var a = (*error)(nil)
	typ1 := reflect.TypeOf((*error)(nil)).Elem()
	fmt.Println(typ1)
}

// ----------------------functional integrity test-------------------------
type Foo int

type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(arg Args, reply *int) error {
	*reply = arg.Num1 + arg.Num2
	return nil
}

// sum is a unexported method
func (f Foo) sum(arg Args, reply *int) error {
	*reply = arg.Num1 + arg.Num2
	return nil
}

func assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	assert(len(s.method) == 1, "wrong service method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	assert(mType != nil, "wrong method, Sum won't be nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	retv := mType.newRetv()
	argv.Set(reflect.ValueOf(Args{1, 3}))

	err := s.call(mType, argv, retv)
	assert(err == nil && *retv.Interface().(*int) == 4 && mType.NumCalls == 1,
		"Failed to call Foo.Sum")
}

func Test(t *testing.T) {
	var a interface{} = 123
	a = a.(int)
	fmt.Println(a)
	fmt.Println(reflect.TypeOf(a))
}
