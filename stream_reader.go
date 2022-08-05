package gogonet

import "github.com/TheMrViper/gogonet/marshal"

type StreamReader struct {
	buffer []byte
}

func NewStreamReader(a []byte) *StreamReader {
	// skip first byte, we dont need it here
	return &StreamReader{
		buffer: a[1:],
	}
}

func (r *StreamReader) ReadInt32() (i int32) {
	// skip var type, we know it already
	_, r.buffer = marshal.DecodeInt32(r.buffer)
	i, r.buffer = marshal.DecodeInt32(r.buffer)

	return
}
