package gogonet

import (
	"github.com/themrviper/gogonet/enet"
	"github.com/themrviper/gogonet/utils"
)

const (
	NETWORK_COMMAND_REMOTE_CALL = iota
	NETWORK_COMMAND_REMOTE_SET
	NETWORK_COMMAND_SIMPLIFY_PATH
	NETWORK_COMMAND_CONFIRM_PATH
	NETWORK_COMMAND_RAW
)

type SentPathCache struct {
	id              uint32
	confirmed_peers map[int32]bool
}

type MultiplayerAPI struct {
	connected_peers map[int32]*enet.Peer

	last_send_cache_id uint32

	recv_path_cache map[int32]map[uint32]string
	sent_path_cache map[string]*SentPathCache
}

func (m *MultiplayerAPI) process_packet(from int32, packet []byte) {
	utils.IfPanic(len(packet) < 1, "Invalid packet received. Size too small.")

	packetType, packet := decode_uint8(packet)

	switch packetType {
	case NETWORK_COMMAND_SIMPLIFY_PATH:
		m.process_simplify_path(from, packet)
	case NETWORK_COMMAND_CONFIRM_PATH:
		m.process_confirm_path(from, packet)
	case NETWORK_COMMAND_REMOTE_CALL:
		fallthrough
	case NETWORK_COMMAND_REMOTE_SET:
		//utils.IfPanic(packetLen < 6, "Invalid packet received. Size too small.")

		target, packet := decode_uint32(packet)
		node := m.process_get_node(from, target, packet)

		if node == nil {
			// process custom solution
			// if rpc func just added without node, set node to dummy
			node = rpc_dummy_node
		}

		name, packet := decode_cstring(packet)

		if packetType == NETWORK_COMMAND_REMOTE_CALL {
			m.process_rpc(node, name, from, packet)
		} else {
			m.process_rset(node, name, from, packet)
		}
	case NETWORK_COMMAND_RAW:
		break
	}
}

func (m *MultiplayerAPI) process_get_node(from int32, target uint32, packet []byte) *Node {

	if target&0x80000000 > 0 {
		ofs := (target & 0x7FFFFFFF)

		// 1 - packet type size, 4 - target size
		utils.IfPanic(ofs-1-4 >= uint32(len(packet)), "Invalid packet received. Size smaller than declared.")

		path, _ := decode_cstring(packet[ofs-1-4:])

		return GetRootNode().GetNode(path)
	} else {
		if nodes, ok := m.recv_path_cache[from]; ok {
			if path, ok := nodes[target]; ok {

				return GetRootNode().GetNode(path)
			}

			utils.Panic("Invalid packet received. Unabled to find requested cached node.")
		}

		utils.Panic("Invalid packet received. Requests invalid peer cache.")
	}

	return nil
}

func (m *MultiplayerAPI) send_simplify_path(node *Node, to int32) (cache *SentPathCache, has_all_peers bool) {
	ok := false
	m.last_send_cache_id += 1

	if cache, ok = m.sent_path_cache[node.Path()]; !ok {
		cache := &SentPathCache{
			id: m.last_send_cache_id,

			confirmed_peers: make(map[int32]bool),
		}

		m.sent_path_cache[node.Path()] = cache
	}

	has_all_peers = true

	for peer_id, _ := range m.connected_peers {

		if to < 0 && peer_id == -to {
			continue // Continue, excluded.
		}

		if to > 0 && peer_id != to {
			continue // Continue, not for this peer.
		}

		if confirmed, ok := cache.confirmed_peers[peer_id]; !confirmed || !ok {

			//TODO generate packet

			has_all_peers = false
		}
	}

	return cache, has_all_peers
}
func (m *MultiplayerAPI) process_simplify_path(from int32, packet []byte) {
	utils.IfPanic(len(packet) < 5, "Invalid packet received. Size too small.")

	id, packet := decode_uint32(packet)
	path, packet := decode_cstring(packet)

	if _, ok := m.recv_path_cache[from]; !ok {
		m.recv_path_cache[from] = make(map[uint32]string)
	}

	m.recv_path_cache[from][id] = path

	m.send_confirm_path(from, path)
}

func (m *MultiplayerAPI) send_confirm_path(to int32, path string) {
	packet := make([]byte, 0, len(path)+2+8)

	packet = encode_int32(1, packet)
	packet = encode_int32(to, packet)
	packet = encode_int8(int8(NETWORK_COMMAND_CONFIRM_PATH), packet)
	packet = encode_cstring(packet, path)

	m.SendPacket(to, packet, enet.FlagReliable)
}

func (m *MultiplayerAPI) process_confirm_path(from int32, packet []byte) {
	utils.IfPanic(len(packet) < 2, "Invalid packet received. Size too small.")

	decode_uint8(packet)
	path, _ := decode_cstring(packet)

	cache, ok := m.sent_path_cache[path]
	utils.IfPanic(!ok, "Invalid packet received. Tries to confirm a path which was not found in cache.")

	_, ok = cache.confirmed_peers[from]
	utils.IfPanic(!ok, "Invalid packet received. Source peer was not found in cache for the given path.")

	m.sent_path_cache[path].confirmed_peers[from] = true
}

func (m *MultiplayerAPI) process_rpc(node *Node, procedureName string, from int32, packet []byte) {
	utils.IfLog(!canCallProcedure(node, procedureName), 6, "Uknnown procedure call ", procedureName, node.Path())

	procedureVariables = decodePacketVariables(packet)

	go reflectProcedureCall(node, procedureName, procedureVariables)
}

func (m *MultiplayerAPI) process_rset(node *Node, variableName string, from int32, packet []byte) {
	utils.Log(6, "variable set", variableName, from)
	utils.Logf(6, "params data: %x\n", packet)

	utils.Panic("Not implemented")
}

func (m *MultiplayerAPI) OnConnected(peer_id int32, peer *enet.Peer) {
	_, ok := m.connected_peers[peer_id]
	utils.IfPanic(ok, "Duplicate peer id, how its happend???")

	m.connected_peers[peer_id] = peer
	m.recv_path_cache[peer_id] = make(map[uint32]string)
}

func (m *MultiplayerAPI) OnDisconnected(peer_id int32) {

	delete(m.connected_peers, peer_id)
	delete(m.recv_path_cache, peer_id)

	for path, cache := range m.sent_path_cache {

		for confirmed_peer_id := range cache.confirmed_peers {

			if confirmed_peer_id == peer_id {
				delete(m.sent_path_cache[path].confirmed_peers, peer_id)
			}
		}
	}
}

func (m *MultiplayerAPI) OnMessage(peer_id int32, packet []byte) {
	//defer utils.Recover("OnMessage")

	utils.Logf(6, "OnMessage from %d packet %x \n", peer_id, packet)
	m.process_packet(peer_id, packet)
}

func (m *MultiplayerAPI) SendPacket(to int32, packet []byte, flags enet.Flag) error {

	utils.Logf(6, "send: %d\n", to)
	if peer, ok := m.connected_peers[to]; ok {
		utils.Logf(6, "Send: to %d packet %x\n", to, packet)
		utils.Log(6, "Send result: ", peer.Send(1, packet, flags))
	}

	return nil
}
