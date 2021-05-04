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
