package enet

/*
#include <enet/enet.h>
*/
import "C"

func CreateServer(address string, port uint16) *EnetBase {
	server := newEnetBase()
	server.server = true
	server.serverRelay = false

	var caddr C.ENetAddress

	enet_address_set_host(&caddr, address)
	enet_address_set_port(&caddr, port)

	server.chost = enet_host_create(&caddr, server.maxClients, server.channelCount.Int8(), server.inBandwidth, server.outBandwidth)

	enet_host_compress(server.chost)

	return server
}
