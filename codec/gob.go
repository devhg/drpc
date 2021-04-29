package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn   io.ReadWriteCloser
	buff   *bufio.Writer
	decode *gob.Decoder
	encode *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn:   conn,
		buff:   buf,
		decode: gob.NewDecoder(conn),
		encode: gob.NewEncoder(buf),
	}
}

func (g *GobCodec) ReadHeader(header *Header) error {
	return g.decode.Decode(header)
}

func (g *GobCodec) ReadBody(body interface{}) error {
	return g.decode.Decode(body)
}

func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buff.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()

	if err := g.encode.Encode(header); err != nil {
		log.Println("rpc codec: gob encoding header error:", err)
		return err
	}
	if err := g.encode.Encode(body); err != nil {
		log.Println("rpc codec: gob encoding body error:", err)
		return err
	}
	return nil
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}
