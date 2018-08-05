// Package encoding defines the MsgPack codec.
package encoding

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	msgCodec "github.com/shamaton/go/codec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

const (
	// Name is the name registered for the msgpack encoder.
	Name = "msgpack"
)

// codec implements encoding.Codec to encode messages into msgpack.
type codec struct {
	//m jsonpb.Marshaler
	//u jsonpb.Unmarshaler
}

/*
func (c *codec) _Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("not a proto message but %T: %v", v, v)
	}

	var w bytes.Buffer
	if err := c.m.Marshal(&w, msg); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
*/

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("not a proto message but %T: %v", v, v)
	}

	var data []byte
	mh := &msgCodec.MsgpackHandle{}
	mh.MapType = reflect.TypeOf(v)
	encoder := msgCodec.NewEncoderBytes(&data, mh)

	e := encoder.Encode(msg)
	if e != nil {
		return nil, e
	}
	return data, e
}

/*
func (c *codec) _Unmarshal(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("not a proto message but %T: %v", v, v)
	}
	return c.u.Unmarshal(bytes.NewReader(data), msg)
}
*/

func (c *codec) Unmarshal(data []byte, v interface{}) error {

	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("not a proto message but %T: %v", v, v)
	}

	// decode(codec)
	mh := &msgCodec.MsgpackHandle{RawToString: true}
	dec := msgCodec.NewDecoderBytes(data, mh)
	return dec.Decode(msg)
}

// Name returns the identifier of the codec.
func (c *codec) Name() string {
	return Name
}

func (c *codec) String() string {
	return Name
}

// New returns Codec interface.
func New() grpc.Codec {
	return new(codec)
}

func init() {
	encoding.RegisterCodec(new(codec))
}
