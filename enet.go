package gogonet

import (
	"log"
	"net"
	"time"

	"github.com/themrviper/gogonet/enet"
)

type IRPC interface {
	New() IRPC

	Process()

	Marshal(StreamWriter)
	Unmarshal(StreamReader)
}

func CreateServer(address string, port string) {
	err := enet.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer enet.Deinitialize()

	addr, err := net.ResolveUDPAddr("udp", address+":"+port)
	if err != nil {
		log.Fatalf("Invalid udp address: '%s'", err)
	}

	host, err := enet.CreateHost(addr, 2, 2, 0, 0)
	if err != nil {
		log.Fatalf("Failed to create host: '%s'", err)
	}
	defer host.Close()
	log.Printf("Start on: '%s'", addr)

	multiplayer_api := &MultiplayerAPI{
		connected_peers: make(map[int32]*enet.Peer),

		recv_path_cache: make(map[int32]map[uint32]string),
		sent_path_cache: make(map[string]*SentPathCache),
	}

	for {
		event, err := host.Service(3 * time.Second)
		if err != nil {
			log.Fatal(err)
		}

		if event == nil {
			continue
		}

		switch event.EventType {
		case enet.EventTypeConnect:
			log.Println("new connection: ", event.Data)

			multiplayer_api.OnConnected(int32(event.Data), event.Peer)
		case enet.EventTypeDisconnect:
			log.Println("disconnection: ", event.Data)

			multiplayer_api.OnDisconnected(int32(event.Data))

		case enet.EventTypeReceive:
			log.Printf("Pure packet %x\n", event.Packet)
			peer_id, packet := decode_int32(event.Packet)
			multiplayer_api.OnMessage(peer_id, packet[4:])
		}
	}
}
