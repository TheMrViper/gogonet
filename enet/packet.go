package enet

/*
#include <enet/enet.h>
*/
import "C"

import "unsafe"

//A Flag represents a bitmask flag.
type Flag uint32

//Packet delivery flags.
const (
	FlagReliable           Flag = C.ENET_PACKET_FLAG_RELIABLE
	FlagUnsequenced        Flag = C.ENET_PACKET_FLAG_UNSEQUENCED
	FlagNoAllocate         Flag = C.ENET_PACKET_FLAG_NO_ALLOCATE
	FlagUnreliableFragment Flag = C.ENET_PACKET_FLAG_UNRELIABLE_FRAGMENT
)

func toCPacket(data []byte, flags Flag) *C.ENetPacket {
	cpacket := C.enet_packet_create(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.enet_uint32(flags))
	if cpacket == nil {
		panic("Allocation failure inside ENet")
	}
	return cpacket
}

func fromCPacket(cpacket *C.ENetPacket) []byte {
	if cpacket == nil {
		return nil
	}
	defer C.enet_packet_destroy(cpacket)
	return C.GoBytes(unsafe.Pointer(cpacket.data), C.int(cpacket.dataLength))
}
