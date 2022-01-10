package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis/v8/internal/proto"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		IdleTimeout: time.Hour,
		ReadTimeout: time.Hour,
		//TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})
	defer redisClient.Close()

	ctx := context.Background()

	pingResp := redisClient.Ping(ctx)
	if pingResp.Err() != nil {
		panic(pingResp)
	}

	clientConn, err := clientConn()
	if err != nil {
		panic(err)
	}
	defer clientConn.Close()

	fmt.Println("new client connected")

	if err := process(ctx, clientConn, redisClient); err != nil {
		panic(err)
	}
}

func process(ctx context.Context, clientConn net.Conn, redisClient *redis.Client) error {
	clientReader := proto.NewReader(clientConn)
	buf := new(bytes.Buffer)
	wr := proto.NewWriter(buf)

	for {
		cmd := &redis.Cmd{}
		if err := cmd.ReadReply(clientReader); err != nil {
			return err
		}

		fmt.Printf("client cmd: %v\n", cmd)

		val := cmd.Val().([]interface{})
		nCmd := redis.NewCmd(ctx, val...)

		err := redisClient.Process(ctx, nCmd)
		if err != nil {
			return err
		}

		vals, err := nCmd.Result()
		if err != nil {
			return err
		}

		if err := writeCmd(wr, vals); err != nil {
			return err
		}

		fmt.Printf("redis err: %v args: %v\n", err, buf)

		if _, err := clientConn.Write(buf.Bytes()); err != nil {
			return err
		}

		buf.Reset()
	}
}

func writeCmd(wr *proto.Writer, vals interface{}) error {
	switch val := vals.(type) {
	case []interface{}:
		if err := wr.WriteByte(proto.ArrayReply); err != nil {
			return err
		}
		n := len(val)
		if err := wr.WriteLen(n); err != nil {
			return err
		}

		for _, v0 := range val {
			if err := writeCmd(wr, v0); err != nil {
				return err
			}
		}
	case interface{}:
		err := wr.WriteArg(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func clientConn() (net.Conn, error) {
	conn, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		return nil, err
	}

	return conn.Accept()
}
