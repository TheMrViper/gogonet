package enet

/*
#include <enet/enet.h>
*/
import "C"

import "errors"

//Peer wraps C.ENetPeer.
type Peer struct {
	raw  *C.ENetPeer
	data interface{}
}

//SetData ...
func (peer *Peer) SetData(data interface{}) {
	peer.data = data
}

//Data ...
func (peer *Peer) Data() interface{} {
	return peer.data
}

//DisconnectNow disconnect forcly.
func (peer *Peer) DisconnectNow(data uint32) {
	C.enet_peer_disconnect_now(peer.raw, C.enet_uint32(data))
}

//Disconnect gently.
func (peer *Peer) Disconnect(data uint32) {
	C.enet_peer_disconnect(peer.raw, C.enet_uint32(data))
}

//DisconnectLater disconnect after all queued outgoing packets are sent.
func (peer *Peer) DisconnectLater(data uint32) {
	C.enet_peer_disconnect_later(peer.raw, C.enet_uint32(data))
}

//Send packet through a channel.
func (peer *Peer) Send(channelID uint8, data []byte, flags Flag) error {
	cpacket := toCPacket(data, flags)

	ret := C.enet_peer_send(peer.raw, C.enet_uint8(channelID), cpacket)
	if ret < 0 {
		return errors.New("ENet failed to send packet")
	}
	return nil
}

//Receive calls C.enet_peer_receive.
func (peer *Peer) Receive() ([]byte, uint8) {
	var channelID C.enet_uint8
	packet := C.enet_peer_receive(peer.raw, &channelID)
	return fromCPacket(packet), uint8(channelID)
}

//Reset calls C.enet_peer_reset.
func (peer *Peer) Reset() {
	C.enet_peer_reset(peer.raw)
}

//Ping calls C.enet_peer_ping.
func (peer *Peer) Ping() {
	C.enet_peer_ping(peer.raw)
}
