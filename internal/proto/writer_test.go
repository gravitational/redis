package proto_test

import (
	"bytes"
	"encoding"
	"fmt"
	"net"
	"testing"
	"time"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"

	"github.com/redis/go-redis/v9/internal/proto"
)

type MyType struct{}

var _ encoding.BinaryMarshaler = (*MyType)(nil)

func (t *MyType) MarshalBinary() ([]byte, error) {
	return []byte("hello"), nil
}

var _ = Describe("WriteBuffer", func() {
	var buf *bytes.Buffer
	var wr *proto.Writer

	BeforeEach(func() {
		buf = new(bytes.Buffer)
		wr = proto.NewWriter(buf)
	})

	It("should write args", func() {
		err := wr.WriteArgs([]interface{}{
			"string",
			12,
			34.56,
			[]byte{'b', 'y', 't', 'e', 's'},
			true,
			nil,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(buf.Bytes()).To(Equal([]byte("*6\r\n" +
			"$6\r\nstring\r\n" +
			"$2\r\n12\r\n" +
			"$5\r\n34.56\r\n" +
			"$5\r\nbytes\r\n" +
			"$1\r\n1\r\n" +
			"$0\r\n" +
			"\r\n")))
	})

	It("should append time", func() {
		tm := time.Date(2019, 1, 1, 9, 45, 10, 222125, time.UTC)
		err := wr.WriteArgs([]interface{}{tm})
		Expect(err).NotTo(HaveOccurred())

		Expect(buf.Len()).To(Equal(41))
	})

	It("should append marshalable args", func() {
		err := wr.WriteArgs([]interface{}{&MyType{}})
		Expect(err).NotTo(HaveOccurred())

		Expect(buf.Len()).To(Equal(15))
	})

	It("should append net.IP", func() {
		ip := net.ParseIP("192.168.1.1")
		err := wr.WriteArgs([]interface{}{ip})
		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal(fmt.Sprintf("*1\r\n$16\r\n%s\r\n", bytes.NewBuffer(ip))))
	})
})

type discard struct{}

func (discard) Write(b []byte) (int, error) {
	return len(b), nil
}

func (discard) WriteString(s string) (int, error) {
	return len(s), nil
}

func (discard) WriteByte(c byte) error {
	return nil
}

func BenchmarkWriteBuffer_Append(b *testing.B) {
	buf := proto.NewWriter(discard{})
	args := []interface{}{"hello", "world", "foo", "bar"}

	for i := 0; i < b.N; i++ {
		err := buf.WriteArgs(args)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestWriteStatus(t *testing.T) {
	inputStatusBytes := []byte("+status\r\n")

	// Read it.
	reader := proto.NewReader(bytes.NewReader(inputStatusBytes))
	readStatus, err := reader.ReadReply()
	if err != nil {
		t.Errorf("Failed to ReadReply: %v", err)
	}

	if readStatus != proto.StatusString("status") {
		t.Errorf("expect read %v but got %v", "status", readStatus)
	}

	// Write it.
	outputStatusBytes := new(bytes.Buffer)
	writer := proto.NewWriter(outputStatusBytes)
	err = writer.WriteArg(readStatus)
	if err != nil {
		t.Errorf("Failed to WriteArg: %v", err)
	}

	if string(inputStatusBytes) != outputStatusBytes.String() {
		t.Errorf("expect written %v but got %v", string(inputStatusBytes), outputStatusBytes.String())
	}
}
