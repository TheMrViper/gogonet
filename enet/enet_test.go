package enet

import (
	"errors"
	"net"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestConvAddr(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:9998")

	if err != nil {
		t.Error(err)
	}

	if addr.Port != 9998 {
		t.Error(addr)
	}

	caddr := toUDPAddr(addr)
	if caddr.port != 9998 {
		t.Error(caddr)
	}

	cbuf := [4]byte{}
	cbuf[0] = byte(caddr.host & 0xff)
	cbuf[1] = byte((caddr.host >> 8) & 0xff)
	cbuf[2] = byte((caddr.host >> 16) & 0xff)
	cbuf[3] = byte((caddr.host >> 24) & 0xff)

	buf := [...]byte{127, 0, 0, 1}

	if cbuf != buf {
		t.Errorf("%v != %v", cbuf, buf)
	}
}

func TestMainFlow(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:9201")
	if err != nil {
		t.Fatal(err)
	}

	client, err := CreateHost(nil, 1, 2, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	server, err := CreateHost(addr, 32, 2, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	//connect
	peer, err := client.Connect(addr, 2, 0)
	if err != nil {
		t.Fatal(err)
	}

	//service connecting
	group := sync.WaitGroup{}
	group.Add(2)
	go func() {
		_, err := server.Service(1 * time.Second)
		if err != nil {
			t.Error(err)
		}
		group.Done()
	}()
	go func() {
		evt, err := client.Service(1 * time.Second)
		if err != nil {
			t.Error(err)
		}
		if evt == nil || evt.EventType != EventTypeConnect {
			t.Error(errors.New("client do not get connect event"))
		}
		if evt.Peer != peer {
			t.Error(errors.New("event do not recover peer"))
		}
		group.Done()
	}()
	group.Wait()
	if t.Failed() {
		runtime.Goexit()
	}

	//send
	payload := []byte("hello,server!")
	err = peer.Send(0, payload, FlagReliable)
	if err != nil {
		t.Fatal(err)
	}
	client.Flush()

	//send receive
	var peer2 *Peer
	evt, err := server.Service(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if evt == nil || evt.EventType != EventTypeConnect {
		t.Fatal(errors.New("server not get connect event"))
	}
	peer2 = evt.Peer
	evt, err = server.Service(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if evt == nil || evt.EventType != EventTypeReceive {
		t.Fatal(errors.New("server not get receive event"))
	} else {
		if !reflect.DeepEqual(payload, evt.Packet) {
			t.Fatal(errors.New("payloads not the same"))
		}
		if peer2 != evt.Peer {
			t.Fatal("can't recover peer")
		}
	}

	//response
	payload2 := []byte("hello,client!")
	err = peer2.Send(1, payload2, FlagReliable)
	if err != nil {
		t.Fatal(err)
	}
	server.Flush()

	//response receive
	evt, err = client.Service(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if evt == nil || evt.EventType != EventTypeReceive {
		t.Fatal(errors.New("client not get receive event"))
	} else {
		if !reflect.DeepEqual(payload2, evt.Packet) {
			t.Fatal(errors.New("payload2 not the same"))
		}
	}

	//disconnect
	peer.Disconnect(123)
	client.Flush()
	evt, err = server.Service(1 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if evt == nil || evt.EventType != EventTypeDisconnect {
		t.Fatal(errors.New("server not get disconnect event"))
	}

}
