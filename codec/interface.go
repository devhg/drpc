package codec

import "io"

// codec package 主要负责编解码相关的工作

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}
