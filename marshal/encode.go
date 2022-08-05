package marshal

import (
	"encoding/binary"
	"math"
)

func EncodeByte(b byte, buffer []byte) []byte {
	return append(buffer, b)
}
func EncodeBytes(b []byte, buffer []byte) []byte {
	return append(buffer, b...)
}

func EncodeBool(v bool, buffer []byte) []byte {
	var b byte = 0
	if v {
		b = 1
	}
	return append(buffer, b)
}
func EncodeInt8(i int8, buffer []byte) []byte {
	return append(buffer, byte(i))
}

func EncodeInt16(i int16, buffer []byte) []byte {
	r := make([]byte, 2)
	binary.LittleEndian.PutUint16(r, uint16(i))

	return append(buffer, r...)
}

func EncodeInt32(i int32, buffer []byte) []byte {
	r := make([]byte, 4)
	binary.LittleEndian.PutUint32(r, uint32(i))

	return append(buffer, r...)
}
func EncodeInt64(i int64, buffer []byte) []byte {
	r := make([]byte, 8)
	binary.LittleEndian.PutUint64(r, uint64(i))

	return append(buffer, r...)
}

func EncodeUint8(i uint8, buffer []byte) []byte {
	return append(buffer, byte(i))
}

func EncodeUint16(i uint16, buffer []byte) []byte {
	r := make([]byte, 2)
	binary.LittleEndian.PutUint16(r, (i))

	return append(buffer, r...)
}

func EncodeUint32(i uint32, buffer []byte) []byte {
	r := make([]byte, 4)
	binary.LittleEndian.PutUint32(r, (i))

	return append(buffer, r...)
}
func EncodeUint64(i uint64, buffer []byte) []byte {
	r := make([]byte, 8)
	binary.LittleEndian.PutUint64(r, (i))

	return append(buffer, r...)
}

func EncodeFloat32(f float32, buffer []byte) []byte {
	i := math.Float32bits(f)
	return EncodeUint32(i, buffer)
}

func EncodeFloat64(f float64, buffer []byte) []byte {
	i := math.Float64bits(f)
	return EncodeUint64(i, buffer)
}

func EncodeCString(s string, buffer []byte) []byte {
	return append(buffer, append([]byte(s), 0x00)...)
}

func EncodeString(s string, buffer []byte) []byte {
	buffer = EncodeUint32(uint32(len(s)), buffer)
	buffer = append(buffer, []byte(s)...)
	if len(s)%4 > 0 {
		buffer = EncodeBytes(buffer, make([]byte, 4-len(s)%4))
	}

	return buffer
}
