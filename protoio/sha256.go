package protoio

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
)

type Sha256Writer struct {
	dest io.Writer
	h    hash.Hash
}

func NewSha256Writer(dest io.Writer) *Sha256Writer {
	h := sha256.New()
	return &Sha256Writer{dest: dest, h: h}
}

func (w *Sha256Writer) Write(b []byte) (int, error) {
	n, err := w.h.Write(b)
	if err != nil {
		return n, err
	}
	if n != len(b) {
		return n, fmt.Errorf("error updating sha256")
	}
	return w.dest.Write(b)
}

func (w *Sha256Writer) Hash() []byte {
	return w.h.Sum(nil)
}
