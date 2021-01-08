package enet

/*
#include <enet/enet.h>
*/
import "C"

import "errors"
import "net"
import "time"
import "unsafe"

//Host warps C.ENetHost.
type Host struct {
	chost *C.ENetHost
	peers map[*C.ENetPeer]*Peer
}

//CreateHost use net.UDPAddr instead of C.ENetHost, this address is used for bind.
func CreateHost(address *net.UDPAddr, peerCount uint, channelLimit uint, incomingBandwidth uint32, outgoingBandwith uint32) (*Host, error) {
	var chost *C.ENetHost
	if address != nil {
		caddr := toUDPAddr(address)
		chost = C.enet_host_create(&caddr, C.size_t(peerCount), C.size_t(channelLimit),
			C.enet_uint32(incomingBandwidth), C.enet_uint32(outgoingBandwith))
	} else {
		chost = C.enet_host_create(nil, C.size_t(peerCount), C.size_t(channelLimit),
			C.enet_uint32(incomingBandwidth), C.enet_uint32(outgoingBandwith))
	}

	if chost == nil {
		return nil, errors.New("enet failed to create an ENetHost")
	}

	//Sets the packet compressor the host should use to compress and decompress packets.
	C.enet_host_compress(chost, nil)
	return &Host{chost, make(map[*C.ENetPeer]*Peer)}, nil
}

//Close it.
func (host *Host) Close() {
	if host != nil && host.chost != nil {
		C.enet_host_destroy(host.chost)
		host.chost = nil
	}
}

//Connect to target host using address, a uint32 data can be carried.
//There are many channels in one connection.
func (host *Host) Connect(address *net.UDPAddr, channelCount uint, data uint) (*Peer, error) {
	caddr := toUDPAddr(address)
	cpeer := C.enet_host_connect(host.chost, &caddr, C.size_t(channelCount), C.enet_uint32(data))

	if cpeer == nil {
		return nil, errors.New("no available peers for initiating an ENet connection")
	}

	peer := &Peer{cpeer, nil}
	host.peers[cpeer] = peer
	return peer, nil
}

//Service calls C.enet_host_service.
func (host *Host) Service(timeout time.Duration) (*Event, error) {
	if timeout < 0 {
		return nil, errors.New("timeout duration was negative")
	}

	var cevent C.ENetEvent
	ret := C.enet_host_service(host.chost, &cevent, C.enet_uint32(timeout/time.Millisecond))
	switch {
	case ret < 0:
		return nil, errors.New("enet internal error")
	case ret > 0:
		evt := toEvent(&cevent)
		//recover peer
		peer := host.peers[cevent.peer]
		if peer == nil {
			peer = &Peer{cevent.peer, nil}
			host.peers[cevent.peer] = peer
		}
		evt.Peer = peer
		return evt, nil
	default:
		return nil, nil
	}
}

//Flush todo
func (host *Host) Flush() {
	C.enet_host_flush(host.chost)
}

//Broadcast todo
func (host *Host) Broadcast(channelID uint8, packet []byte, flags Flag) {
	C.enet_host_broadcast(host.chost, C.enet_uint8(channelID), toCPacket(packet, flags))
}

func toUDPAddr(address *net.UDPAddr) (caddr C.ENetAddress) {
	cip := C.CString(address.IP.String())
	defer C.free(unsafe.Pointer(cip))

	C.enet_address_set_host(&caddr, cip)
	caddr.port = C.enet_uint16(address.Port)
	return
}
