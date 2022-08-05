package enet

/*
#cgo pkg-config: libenet
#include <enet/enet.h>
*/
import "C"
import "unsafe"

type PacketFlag uint32

const (
	PacketFlagReliable           PacketFlag = C.ENET_PACKET_FLAG_RELIABLE
	PacketFlagUnsequenced        PacketFlag = C.ENET_PACKET_FLAG_UNSEQUENCED
	PacketFlagNoAllocate         PacketFlag = C.ENET_PACKET_FLAG_NO_ALLOCATE
	PacketFlagUnreliableFragment PacketFlag = C.ENET_PACKET_FLAG_UNRELIABLE_FRAGMENT
)

type EventType uint32

const (
	EventTypeNone       EventType = C.ENET_EVENT_TYPE_NONE
	EventTypeConnect    EventType = C.ENET_EVENT_TYPE_CONNECT
	EventTypeDisconnect EventType = C.ENET_EVENT_TYPE_DISCONNECT
	EventTypeReceive    EventType = C.ENET_EVENT_TYPE_RECEIVE
)

func init() {
	err := enet_initialize()
	if err != 0 {
		panic("Cant init ENet")
	}
}

//Initialize must be called before use enet.
func enet_initialize() int {
	return int(C.enet_initialize())
}

//Deinitialize must be called after use enet.
func enet_einitialize() {
	C.enet_deinitialize()
}

func enet_host_create(caddr *C.ENetAddress, peerCount uint32, channelLimit int8, incomingBandwidth uint32, outgoingBandwith uint32) *C.ENetHost {
	return C.enet_host_create(caddr, C.size_t(peerCount), C.size_t(channelLimit), C.enet_uint32(incomingBandwidth), C.enet_uint32(outgoingBandwith))
}

func enet_host_service(chost *C.ENetHost, cevent *C.ENetEvent, timeout uint32) int {
	return int(C.enet_host_service(chost, cevent, C.enet_uint32(timeout)))
}

func enet_host_connect(chost *C.ENetHost, caddr *C.ENetAddress, channelCount uint, data uint32) *C.ENetPeer {
	return C.enet_host_connect(chost, caddr, C.size_t(channelCount), C.enet_uint32(data))
}

func enet_host_destroy(chost *C.ENetHost) {
	C.enet_host_destroy(chost)
}

func enet_host_compress(chost *C.ENetHost) {
	C.enet_host_compress(chost, nil)
}

func enet_host_flush(chost *C.ENetHost) {
	C.enet_host_flush(chost)
}

func enet_host_broadcast(chost *C.ENetHost, channelID SystemChannelFlag, packet *C.ENetPacket) {
	C.enet_host_broadcast(chost, C.enet_uint8(channelID), packet)
}

func enet_address_set_host(caddr *C.ENetAddress, ip string) {
	cip := C.CString(ip)
	defer C.free(unsafe.Pointer(cip))
	C.enet_address_set_host(caddr, cip)
}

func enet_address_set_port(caddr *C.ENetAddress, port uint16) {
	caddr.port = C.enet_uint16(port)
}

func enet_packet_create(data []byte, flags PacketFlag) *C.ENetPacket {
	return C.enet_packet_create(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.enet_uint32(flags))
}

func enet_packet_destroy(cpacket *C.ENetPacket) {
	C.enet_packet_destroy(cpacket)
}

func enet_peer_reset(peer *C.ENetPeer) {
	C.enet_peer_reset(peer)
}

func enet_peer_send(peer *C.ENetPeer, channelID SystemChannelFlag, packet *C.ENetPacket) {
	C.enet_peer_send(peer, C.enet_uint8(channelID), packet)
}
