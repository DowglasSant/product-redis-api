package cache

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// Serializer define a interface para serialização de dados
type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

// JSONSerializer implementa serialização usando JSON
type JSONSerializer struct{}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s *JSONSerializer) Name() string {
	return "json"
}

// MsgpackSerializer implementa serialização usando MessagePack
type MsgpackSerializer struct{}

func NewMsgpackSerializer() *MsgpackSerializer {
	return &MsgpackSerializer{}
}

func (s *MsgpackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (s *MsgpackSerializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

func (s *MsgpackSerializer) Name() string {
	return "msgpack"
}
