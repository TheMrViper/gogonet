package enet

/*
#cgo pkg-config: libenet
#include <enet/enet.h>
*/
import "C"

import "errors"

func init() {
	err := Initialize()
	if err != nil {
		panic(err)
	}
}

//Initialize must be called before use enet.
func Initialize() error {
	if C.enet_initialize() != C.int(0) {
		return errors.New("ENet failed to initialize")
	}
	return nil
}

//Deinitialize must be called after use enet.
func Deinitialize() {
	C.enet_deinitialize()
}
