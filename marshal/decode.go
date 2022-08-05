package marshal

import (
	"bytes"
	"encoding/binary"
	"math"
)

func DecodeByte(a []byte) (byte, []byte) {
	return a[0], a[1:]
}

func DecodeBytes(a []byte, n uint32) ([]byte, []byte) {
	return a[:n], a[n:]
}

func DecodeBool(a []byte) (bool, []byte) {
	ok, a := DecodeInt32(a)
	return ok == 1, a
}

func DecodeUint8(a []byte) (uint8, []byte) {
	return a[0], a[1:]
}

func DecodeUint16(a []byte) (uint16, []byte) {
	return binary.LittleEndian.Uint16(a), a[4:]
}

func DecodeUint32(a []byte) (uint32, []byte) {
	return binary.LittleEndian.Uint32(a), a[4:]
}

func DecodeUint64(a []byte) (uint64, []byte) {
	return binary.LittleEndian.Uint64(a), a[8:]
}
func DecodeInt8(a []byte) (int8, []byte) {
	return int8(a[0]), a[1:]
}

func DecodeInt16(a []byte) (int16, []byte) {
	return int16(binary.LittleEndian.Uint16(a)), a[4:]
}

func DecodeInt32(a []byte) (int32, []byte) {
	return int32(binary.LittleEndian.Uint32(a)), a[4:]
}

func DecodeInt64(a []byte) (int64, []byte) {
	return int64(binary.LittleEndian.Uint64(a)), a[8:]
}

func DecodeFloat32(a []byte) (float32, []byte) {
	i, a := DecodeUint32(a)
	return math.Float32frombits(i), a
}

func DecodeFloat64(a []byte) (float64, []byte) {
	i, a := DecodeUint64(a)
	return math.Float64frombits(i), a
}

func DecodeCString(a []byte) (string, []byte) {
	n := bytes.Index(a, []byte{0})
	return string(a[:n]), a[n+1:]
}

func DecodeString(a []byte) (string, []byte) {
	length, a := DecodeUint32(a)

	result, a := string(a[:length]), a[length:]
	if length%4 > 0 {
		_, a = DecodeBytes(a, 4-length%4)
	}

	return result, a
}
