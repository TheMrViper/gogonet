package enet

/*
#include <enet/enet.h>
*/
import "C"

import (
	"unsafe"

	"github.com/TheMrViper/gogonet/marshal"
	"github.com/TheMrViper/gogonet/signals"
	"github.com/TheMrViper/gogonet/utils"
)

type CompressionModeFlag int32

func (i CompressionModeFlag) Int32() int32 {
	return int32(i)
}

const (
	CompressNone       CompressionModeFlag = 0
	CompressRangeCoder CompressionModeFlag = 1
	CompressFastLZ     CompressionModeFlag = 2
	CompressZLIB       CompressionModeFlag = 3
	CompressZSTD       CompressionModeFlag = 4
)

type SystemMessageFlag uint32

func (i SystemMessageFlag) Int32() int32 {
	return int32(i)
}

func (i SystemMessageFlag) Uint32() uint32 {
	return uint32(i)
}

const (
	SystemMessageAddPeer    SystemMessageFlag = 0
	SystemMessageRemovePeer SystemMessageFlag = 1
)

type SystemChannelFlag int8

func (i SystemChannelFlag) Int8() int8 {
	return int8(i)
}

const (
	SystemChannelConfig     SystemChannelFlag = 0
	SystemChannelReliable   SystemChannelFlag = 1
	SystemChannelUnreliable SystemChannelFlag = 2
	SystemChannelMax        SystemChannelFlag = 3
)

type Packet struct {
	source uint32
	target int32
	data   []byte
}
type EnetBase struct {
	active            bool
	server            bool
	dtlsVerify        bool
	dtlsEnabled       bool
	serverRelay       bool
	alwaysOrdered     bool
	refuseConnections bool

	timeout      uint32
	uniqueId     uint32
	maxClients   uint32
	inBandwidth  uint32
	outBandwidth uint32
	channelCount SystemChannelFlag

	transferChannel SystemChannelFlag
	compressionMode CompressionModeFlag

	enetCompressor interface{}

	signals *signals.Signal

	chost *C.ENetHost

	peerMap map[uint32]*C.ENetPeer

	packetChannel chan *Packet
}

func newEnetBase() *EnetBase {
	return &EnetBase{
		active:            false,
		server:            false,
		dtlsVerify:        true,
		dtlsEnabled:       false,
		serverRelay:       false,
		alwaysOrdered:     false,
		refuseConnections: false,

		uniqueId:     0,
		timeout:      0,
		maxClients:   1024,
		inBandwidth:  0,
		outBandwidth: 0,
		channelCount: SystemChannelMax,

		transferChannel: -1,
		compressionMode: CompressNone,

		enetCompressor: nil,

		peerMap: make(map[uint32]*C.ENetPeer),

		signals: signals.New(),

		packetChannel: make(chan *Packet, 1024),
	}
}

func (b *EnetBase) On(name string, f interface{}) {
	b.signals.On(name, f)
}

func (b *EnetBase) Off(name string) {
	b.signals.Off(name)
}

func (b *EnetBase) SetPacketChannelSize(size int) {
	utils.IfPanic(b.active, "Server relaying can't be toggled while the multiplayer instance is active.")
	b.packetChannel = make(chan *Packet, size)
}

func (b *EnetBase) SetServerRelayEnabled(enabled bool) {
	utils.IfPanic(b.active, "Server relaying can't be toggled while the multiplayer instance is active.")
	b.serverRelay = enabled
}

func (b *EnetBase) EnableServerRelay() {
	utils.IfPanic(b.active, "Server relaying can't be toggled while the multiplayer instance is active.")
	b.serverRelay = true
}

func (b *EnetBase) DisableServerRelay() {
	utils.IfPanic(b.active, "Server relaying can't be toggled while the multiplayer instance is active.")
	b.serverRelay = false
}

func (b *EnetBase) SetMaxClients(v uint32) {
	utils.IfPanic(b.active, "Can set property while serving data")
	b.maxClients = v
}

func (b *EnetBase) SetInBandwidth(v uint32) {
	utils.IfPanic(b.active, "Can set property while serving data")
	b.inBandwidth = v
}

func (b *EnetBase) SetOutBandwidth(v uint32) {
	utils.IfPanic(b.active, "Can set property while serving data")
	b.outBandwidth = v
}

func (b *EnetBase) PutPacket() {

}

func (b *EnetBase) GetPacket() (uint32, int32, []byte) {
	packet := <-b.packetChannel
	return packet.source, packet.target, packet.data
}

func (b *EnetBase) ListenAndServe() {
	b.active = true

	var cevent C.ENetEvent
	for {
		ret := enet_host_service(b.chost, &cevent, b.timeout)

		if ret < 0 {
			// Error, do something?
			break
		} else if ret == 0 {
			continue
		}

		print(ret)

		switch EventType(cevent._type) {
		case EventTypeConnect:

			if b.server && b.refuseConnections {
				enet_peer_reset(cevent.peer)
				break
			}

			if _, ok := b.peerMap[uint32(cevent.data)]; b.server && (ok || uint32(cevent.data) < 2) {
				enet_peer_reset(cevent.peer)
				continue
			}

			newId := uint32(cevent.data)

			if newId == 0 {
				newId = 1
			}
			cevent.peer.data = unsafe.Pointer(&newId)

			b.peerMap[newId] = cevent.peer

			print("peer_connected", newId)
			b.signals.Emit("peer_connected", newId)

			if b.server {
				if !b.serverRelay {
					continue
				}

				var packet *C.ENetPacket
				for peerId, peer := range b.peerMap {
					if newId == peerId {
						continue
					}

					buffer := make([]byte, 0, 16)

					buffer = marshal.EncodeUint32(SystemMessageAddPeer.Uint32(), buffer)
					buffer = marshal.EncodeUint32(peerId, buffer)
					buffer = marshal.EncodeUint32(SystemMessageAddPeer.Uint32(), buffer)
					buffer = marshal.EncodeUint32(newId, buffer)

					packet = enet_packet_create(buffer[:8], PacketFlagReliable)
					enet_peer_send(cevent.peer, SystemChannelConfig, packet)

					packet = enet_packet_create(buffer[8:], PacketFlagReliable)
					enet_peer_send(peer, SystemChannelConfig, packet)
				}
			} else {
				b.signals.Emit("connection_succeeded")
			}
		case EventTypeDisconnect:

			id := *(*uint32)(cevent.peer.data)

			//TODO debug watch what is before we set id
			if id == 0 {
				if !b.server {
					b.signals.Emit("connection_failed")
				}
				break
			}

			if !b.server {
				// Client just disconnected from server.
				b.signals.Emit("server_disconnected")
				//close_connection();
				return
			} else if b.serverRelay {
				// Server just received a client disconnect and is in relay mode, notify everyone else.

				var packet *C.ENetPacket
				for peerId, peer := range b.peerMap {
					if id == peerId {
						continue
					}

					buffer := make([]byte, 0, 8)
					buffer = marshal.EncodeUint32(SystemMessageRemovePeer.Uint32(), buffer)
					buffer = marshal.EncodeUint32(id, buffer)

					packet = enet_packet_create(buffer, PacketFlagReliable)
					enet_peer_send(peer, SystemChannelConfig, packet)
				}
			}

			b.signals.Emit("peer_disconnected", id)
			delete(b.peerMap, id)
		case EventTypeReceive:

			if SystemChannelFlag(cevent.channelID) == SystemChannelConfig {
				if cevent.packet.dataLength < 8 {
					continue
				}

				if b.server {
					continue
				}

				data := C.GoBytes(unsafe.Pointer(cevent.packet.data), C.int(cevent.packet.dataLength))
				msg, data := marshal.DecodeUint32(data)
				id, _ := marshal.DecodeUint32(data)

				switch msg {
				case SystemMessageAddPeer.Uint32():
					{
						b.peerMap[id] = nil
						b.signals.Emit("peer_connected", id)

					}
				case SystemMessageRemovePeer.Uint32():
					{
						delete(b.peerMap, id)
						b.signals.Emit("peer_disconnected", id)
					}
				}

				enet_packet_destroy(cevent.packet)
			} else if SystemChannelFlag(cevent.channelID) < b.channelCount {

				if cevent.packet.dataLength < 8 {
					continue
				}

				packet := &Packet{}
				packet.data = C.GoBytes(unsafe.Pointer(cevent.packet.data), C.int(cevent.packet.dataLength))

				id := *(*uint32)(cevent.peer.data)

				packet.source, packet.data = marshal.DecodeUint32(packet.data)
				packet.target, packet.data = marshal.DecodeInt32(packet.data)

				if b.server {

					if id != packet.source {
						continue
					}

					if packet.target == 1 {

						b.packetChannel <- packet

					} else if !b.serverRelay {

						continue

					} else if packet.target == 0 {

						b.packetChannel <- packet

						for peerId, peer := range b.peerMap {
							if peerId == packet.source {
								continue
							}

							packet2 := enet_packet_create(C.GoBytes(unsafe.Pointer(cevent.packet.data), C.int(cevent.packet.dataLength)), PacketFlag(cevent.packet.flags))

							enet_peer_send(peer, SystemChannelFlag(cevent.channelID), packet2)
						}

					} else if packet.target < 0 {

						for peerId, peer := range b.peerMap {
							if peerId == packet.source || (packet.target < 0 && peerId == uint32(packet.target*-1)) {
								continue
							}

							packet2 := enet_packet_create(C.GoBytes(unsafe.Pointer(cevent.packet.data), C.int(cevent.packet.dataLength)), PacketFlag(cevent.packet.flags))

							enet_peer_send(peer, SystemChannelFlag(cevent.channelID), packet2)
						}

						if -packet.target == 1 {

							b.packetChannel <- packet

						} else {
							enet_packet_destroy(cevent.packet)
						}

					} else {

						if peer, ok := b.peerMap[uint32(packet.target)]; ok {
							enet_peer_send(peer, SystemChannelFlag(cevent.channelID), cevent.packet)
						}

					}

				} else {

					b.packetChannel <- packet

				}
			}
		}
	}
}
