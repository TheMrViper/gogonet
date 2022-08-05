package gogonet

import (
	"fmt"
	"sync/atomic"

	"github.com/TheMrViper/gogonet/marshal"
	"github.com/TheMrViper/gogonet/signals"
	"github.com/TheMrViper/gogonet/utils"
)

type NetworkCommand uint8

func (i NetworkCommand) Uint8() uint8 {
	return uint8(i)
}

const (
	CommandRemoteCall   NetworkCommand = 0
	CommandRemoteSet    NetworkCommand = 1
	CommandSimplifyPath NetworkCommand = 2
	CommandConfirmPath  NetworkCommand = 3
	CommandRaw          NetworkCommand = 4
)

type INetworkPeer interface {
	signals.ISubscribeable

	GetPacket() (source uint32, target int32, data []byte)
	PutPacket()
	ListenAndServe()
}

type SentPathCache struct {
	id             uint32
	confirmedPeers map[uint32]bool
}

type MultiplayerAPI struct {
	active bool

	signals *signals.Signal

	networkPeer INetworkPeer

	connectedPeers map[uint32]bool

	lastSendCacheId uint32

	recvPathCache map[uint32]map[uint32]string
	sentPathCache map[string]*SentPathCache
}

func (m *MultiplayerAPI) SetNetworkPeer(peer INetworkPeer) {
	utils.IfPanic(m.active, "Cannot change peer when server is running")

	if m.networkPeer != nil {
		m.networkPeer.Off("peer_connected")
		m.networkPeer.Off("peer_disconnected")
		m.networkPeer.Off("connection_succeeded")
		m.networkPeer.Off("connection_failed")
		m.networkPeer.Off("server_disconnected")
	}

	if peer != nil {
		m.networkPeer = peer
		m.networkPeer.On("peer_connected", m.addPeer)
		m.networkPeer.On("peer_disconnected", m.deletePeer)
		m.networkPeer.On("connection_succeeded", m.connectionSucceeded)
		m.networkPeer.On("connection_failed", m.connectionFailed)
		m.networkPeer.On("server_disconnected", m.serverDisconnected)
	}
}

func (m *MultiplayerAPI) ListenAndServe() {
	utils.IfPanic(m.networkPeer == nil, "NetworkPeer cannot be nil")

	go m.networkPeer.ListenAndServe()

	for {
		source, target, data := m.networkPeer.GetPacket()
		m.processPacket(source, target, data)
	}
}

func (m *MultiplayerAPI) addPeer(id uint32) {
	_, ok := m.connectedPeers[id]
	utils.IfPanic(ok, "Duplicate peer id, how its happend???")

	m.connectedPeers[id] = true
	m.recvPathCache[id] = make(map[uint32]string)
}

func (m *MultiplayerAPI) deletePeer(id uint32) {

	delete(m.connectedPeers, id)
	delete(m.recvPathCache, id)

	for path, cache := range m.sentPathCache {

		for confirmedPeerId := range cache.confirmedPeers {

			if confirmedPeerId == id {
				delete(m.sentPathCache[path].confirmedPeers, id)
			}
		}
	}

	m.signals.Emit("network_peer_disconnected", id)
}

func (m *MultiplayerAPI) connectionSucceeded() {
	m.signals.Emit("connection_succeeded")
}
func (m *MultiplayerAPI) connectionFailed() {
	m.signals.Emit("connection_failed")
}
func (m *MultiplayerAPI) serverDisconnected() {
	m.signals.Emit("server_disconnected")
}

func (m *MultiplayerAPI) processGetNode(source uint32, nodeCachedId uint32, data []byte) INode {
	//defer utils.Recover("process_get_node")

	if nodeCachedId&0x80000000 > 0 {
		ofs := (nodeCachedId & 0x7FFFFFFF)

		// 1 - packet type size, 4 - nodeCachedId size
		utils.IfPanic(ofs-1-4 >= uint32(len(data)), "Invalid packet received. Size smaller than declared.")

		path, _ := marshal.DecodeCString(data[ofs-1-4:])

		fmt.Println(path)
		return GetTree().GetNode(path)
	} else {
		if nodes, ok := m.recvPathCache[source]; ok {
			if path, ok := nodes[nodeCachedId]; ok {

				return GetTree().GetNode(path)
			}

			utils.Panic("Invalid packet received. Unabled to find requested cached node.")
		}

		utils.Panic("Invalid packet received. Requests invalid peer cache.")
	}

	return nil
}

func (m *MultiplayerAPI) sendPacket(node INode, target int32, unreliable bool, procedureName string, v ...interface{}) {

}

func (m *MultiplayerAPI) processPacket(source uint32, target int32, data []byte) {
	//defer utils.Recover("process_packet")
	utils.IfPanic(len(data) < 1, "Invalid packet received. Size too small.")

	packetType, data := marshal.DecodeUint8(data)

	switch packetType {
	case CommandSimplifyPath.Uint8():
		m.processSimplifyPath(source, data)
	case CommandConfirmPath.Uint8():
		m.processConfirmPath(source, data)
	case CommandRemoteCall.Uint8():
		fallthrough
	case CommandRemoteSet.Uint8():
		//utils.IfPanic(packetLen < 6, "Invalid packet received. Size too small.")

		nodeCachedId, data := marshal.DecodeUint32(data)

		node := m.processGetNode(source, nodeCachedId, data)

		utils.IfPanic(node == nil, "Unknown node for rpc call")

		name, data := marshal.DecodeCString(data)

		if packetType == CommandRemoteCall.Uint8() {
			m.processRpc(node, name, source, data)
		} else {
			m.processRset(node, name, source, data)
		}
	case CommandRemoteSet.Uint8():
		m.processRaw(source, data)
	}
}

func (m *MultiplayerAPI) sendRaw(data []byte) {

}

func (m *MultiplayerAPI) processRaw(source uint32, data []byte) {

}

func (m *MultiplayerAPI) sendSimplifyPath(node INode, target int32) (cache *SentPathCache, has_all_peers bool) {
	//defer utils.Recover("send_simplify_path")
	ok := false

	atomic.AddUint32(&m.lastSendCacheId, 1)

	if cache, ok = m.sentPathCache[node.Path()]; !ok {
		cache := &SentPathCache{
			id: m.lastSendCacheId,

			confirmedPeers: make(map[uint32]bool),
		}

		m.sentPathCache[node.Path()] = cache
	}

	has_all_peers = true

	for peerId, _ := range m.connectedPeers {

		if target < 0 && peerId == uint32(target*-1) {
			continue // Continue, excluded.
		}

		if target > 0 && peerId != uint32(target) {
			continue // Continue, not for this peer.
		}

		if confirmed, ok := cache.confirmedPeers[peerId]; !confirmed || !ok {

			//TODO generate packet

			has_all_peers = false
		}
	}

	return cache, has_all_peers
}
func (m *MultiplayerAPI) processSimplifyPath(source uint32, packet []byte) {
	//defer utils.Recover("process_simplify_path")
	utils.IfPanic(len(packet) < 5, "Invalid packet received. Size too small.")

	id, packet := marshal.DecodeUint32(packet)
	path, packet := marshal.DecodeCString(packet)

	if _, ok := m.recvPathCache[source]; !ok {
		m.recvPathCache[source] = make(map[uint32]string)
	}

	m.recvPathCache[source][id] = path

	m.sendConfirmPath(source, path)
}

func (m *MultiplayerAPI) sendConfirmPath(target uint32, path string) {
	//defer utils.Recover("send_confirm_path")
	packet := make([]byte, 0, len(path)+2+8)

	packet = marshal.EncodeUint32(1, packet)
	packet = marshal.EncodeUint32(target, packet)
	packet = marshal.EncodeUint8(CommandConfirmPath.Uint8(), packet)
	packet = marshal.EncodeCString(path, packet)

	//m.sendPacket(to, packet, enet.FlagReliable)
}

func (m *MultiplayerAPI) processConfirmPath(source uint32, packet []byte) {
	//defer utils.Recover("process_confirm_path")
	utils.IfPanic(len(packet) < 2, "Invalid packet received. Size too small.")

	path, _ := marshal.DecodeCString(packet)

	cache, ok := m.sentPathCache[path]
	utils.IfPanic(!ok, "Invalid packet received. Tries to confirm a path which was not found in cache.")

	_, ok = cache.confirmedPeers[source]
	utils.IfPanic(!ok, "Invalid packet received. Source peer was not found in cache for the given path.")

	m.sentPathCache[path].confirmedPeers[source] = true
}

func (m *MultiplayerAPI) processRpc(node INode, procedureName string, source uint32, data []byte) {
	//defer utils.Recover("process_rpc")

	if canCallNativeProcedure(node, procedureName) {
		procedure := getNativeProcedure(node, procedureName)

		streamReader := NewStreamReader(data)
		procedure.SetOwnerNode(node)
		procedure.Unmarshal(streamReader)

		go procedure.Call()
	} else if canCallReflectProcedure(node, procedureName) {

		procedureVariables := reflectDecodePacketVariables(data)

		go reflectProcedureCall(node, procedureName, procedureVariables)
	} else {
		utils.Panic("Uknnown procedure call")
	}
}

func (m *MultiplayerAPI) processRset(node INode, variableName string, source uint32, data []byte) {
	//defer utils.Recover("process_rset")
	utils.Log(6, "variable set", variableName, source)
	utils.Logf(6, "params data: %x\n", data)

	utils.Panic("Not implemented")
}
