package protoio

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

type Writer interface {
	Write(msg proto.Message) error
}

type ParseFunc func([]byte) (proto.Message, error)

type Reader interface {
	Read(ParseFunc) (proto.Message, error)
}

type writerImpl struct {
	dest io.Writer
}

func NewWriter(dest io.Writer) Writer {
	return &writerImpl{dest: dest}
}

func (w *writerImpl) Write(msg proto.Message) error {
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	l := len(b)
	lenbuf := make([]byte, 8)
	n := binary.PutUvarint(lenbuf, uint64(l))
	if n <= 0 {
		return fmt.Errorf("failed to write varint")
	}
	n1, err := w.dest.Write(lenbuf[:n])
	if err != nil {
		return err
	}
	if n != n1 {
		return fmt.Errorf("failed to write proto")
	}
	n1, err = w.dest.Write(b)
	if err != nil {
		return err
	}
	if n1 != len(b) {
		return fmt.Errorf("failed to write proto")
	}
	return nil
}

type byteReaderWrapper struct {
	src io.Reader
}

func (r *byteReaderWrapper) ReadByte() (byte, error) {
	var buf [1]byte
	n, err := r.src.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, fmt.Errorf("failed to read byte")
	}
	return buf[0], nil
}

type readerImpl struct {
	src io.Reader
}

func NewReader(src io.Reader) Reader {
	return &readerImpl{src: src}
}

func (r *readerImpl) Read(f ParseFunc) (proto.Message, error) {
	byteReader := &byteReaderWrapper{src: r.src}
	l, err := binary.ReadUvarint(byteReader)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, l)
	n, err := r.src.Read(buf)
	if err != nil {
		return nil, err
	}
	if uint64(n) != l {
		return nil, fmt.Errorf("failed to read proto")
	}
	return f(buf)
}
