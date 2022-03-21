package transport

import (
	"encoding/binary"
	"io"
	"os"
	"sync"

	"google.golang.org/protobuf/proto"
)

var lenBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, binary.MaxVarintLen64)
		return &buf
	},
}

func WriteProtoStream(dst io.Writer, msg proto.Message) error {
	msgBuf, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	lenBuf := lenBufPool.Get().(*[]byte)
	lenLen := binary.PutUvarint(*lenBuf, uint64(len(msgBuf)))
	_, err = os.Stdout.Write((*lenBuf)[:lenLen])
	lenBufPool.Put(lenBuf)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(msgBuf)
	return err
}

type StreamReader interface {
	io.Reader
	io.ByteReader
}

func ReadProtoStream(src StreamReader, msg proto.Message) error {
	msgLen, err := binary.ReadUvarint(src)
	if err != nil {
		return err
	}
	msgBuf := make([]byte, msgLen)
	remaining := msgBuf
	for {
		n, err := src.Read(remaining)
		if n == len(remaining) {
			return proto.Unmarshal(msgBuf, msg)
		} else if err != nil {
			return err
		} else {
			remaining = remaining[n:]
		}
	}
}
