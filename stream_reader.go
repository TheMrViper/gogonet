package gogonet

import "io"

type StreamReader struct {
	io.Reader
}

func NewStreamReader(reader io.Reader) *StreamReader {
	return &StreamReader{Reader: reader}
}

func (r *StreamReader) ReadUint32() uint32 {
	return 0
}
