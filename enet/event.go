package enet

/*
#include <enet/enet.h>
*/
import "C"

//EventType stand for C.ENetEventType.
type EventType uint

//Constants for EventType.
const (
	EventTypeConnect    EventType = C.ENET_EVENT_TYPE_CONNECT    //on connect
	EventTypeDisconnect EventType = C.ENET_EVENT_TYPE_DISCONNECT //on disconnect
	EventTypeReceive    EventType = C.ENET_EVENT_TYPE_RECEIVE    //on packet receive
)

//Event simulates C.ENetEvent.
type Event struct {
	EventType EventType
	Peer      *Peer
	ChannelID uint8
	Data      uint32
	Packet    []byte
}

func toEvent(cevent *C.ENetEvent) *Event {
	event := Event{
		EventType: EventType(cevent._type),
		Peer:      nil,
		ChannelID: uint8(cevent.channelID),
		Data:      uint32(cevent.data),
		Packet:    fromCPacket(cevent.packet),
	}
	return &event
}
