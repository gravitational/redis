package redis

import (
	"bytes"
	"io"

	"github.com/redis/go-redis/v9/internal/proto"
)

type Reader = proto.Reader
type Writer = proto.Writer
type RedisError = proto.RedisError

const ArrayReply = proto.RespArray
const IntReply = proto.RespInt

func NewReader(rd io.Reader) *Reader {
	return proto.NewReader(rd)
}

func NewWriter(wr *bytes.Buffer) *Writer {
	return proto.NewWriter(wr)
}
