package gogonet

import (
	"bytes"
	"encoding/binary"
)

func decode_cstring(a []byte) (string, []byte) {
	n := bytes.Index(a, []byte{0})
	return string(a[:n]), a[n+1:]
}

func encode_cstring(a []byte, s string) []byte {
	return append(a, append([]byte(s), 0x00)...)
}

func decode_uint8(a []byte) (uint8, []byte) {
	return a[0], a[1:]
}

func decode_uint16(a []byte) (uint16, []byte) {
	return binary.LittleEndian.Uint16(a), a[4:]
}

func decode_uint32(a []byte) (uint32, []byte) {
	return binary.LittleEndian.Uint32(a), a[4:]
}

func decode_uint64(a []byte) (uint64, []byte) {
	return binary.LittleEndian.Uint64(a), a[8:]
}
func decode_int8(a []byte) (int8, []byte) {
	return int8(a[0]), a[1:]
}

func decode_int16(a []byte) (int16, []byte) {
	return int16(binary.LittleEndian.Uint16(a)), a[4:]
}

func decode_int32(a []byte) (int32, []byte) {
	return int32(binary.LittleEndian.Uint32(a)), a[4:]
}

func decode_int64(a []byte) (int64, []byte) {
	return int64(binary.LittleEndian.Uint64(a)), a[8:]
}

func encode_int8(i int8, b []byte) []byte {
	return append(b, byte(i))
}

func encode_int32(i int32, b []byte) []byte {
	r := make([]byte, 4)
	binary.LittleEndian.PutUint32(r, uint32(i))

	return append(b, r...)
}
