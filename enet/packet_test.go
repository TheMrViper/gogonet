package enet

import "testing"
import "unsafe"

func TestCreatepacket(t *testing.T) {
	//bytes := []byte("Test string")
	bytes := []byte{1, 2, 3, 4}
	p := toCPacket(bytes, FlagReliable)
	if p.flags != _Ctype_enet_uint32(FlagReliable) {
		t.Logf("expected %#v, got %#v", p.flags, FlagReliable)
	}
	if p.dataLength != _Ctype_size_t(len(bytes)) {
		t.Logf("expected %#v, got %#v", p.dataLength, len(bytes))
	}
	for i, v := range bytes {
		bp := (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(p.data)) + uintptr(i)))
		b := *bp
		if b != bytes[i] {
			t.Errorf("index: %v, expected %v, go %v", i, v, b)
		}
	}
}

func TestPacketroundtrip(t *testing.T) {
	bytes := []byte("TEST STRING")
	p := toCPacket(bytes, FlagReliable)
	bytes2 := fromCPacket(p)
	for i := range bytes {
		if bytes[i] != bytes2[i] {
			t.Errorf("index: %v, expected %v, go %v", i, bytes[i], bytes2[i])
		}
	}
}
