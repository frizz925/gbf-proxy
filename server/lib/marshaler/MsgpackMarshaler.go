package marshaler

import (
	"github.com/vmihailenco/msgpack/v4"
)

type MsgpackMarshaler struct{}

var _ Marshaler = (*MsgpackMarshaler)(nil)

func NewMsgpackMarshaler() *MsgpackMarshaler {
	return &MsgpackMarshaler{}
}

func (*MsgpackMarshaler) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (*MsgpackMarshaler) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}
