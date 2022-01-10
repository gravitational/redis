package redis

import (
	"bytes"
	"io"

	"github.com/go-redis/redis/v8/internal/proto"
)

type Reader = proto.Reader
type Writer = proto.Writer

const ArrayReply = proto.ArrayReply

func NewReader(rd io.Reader) *Reader {
	return proto.NewReader(rd)
}

func NewWriter(wr *bytes.Buffer) *Writer {
	return proto.NewWriter(wr)
}
