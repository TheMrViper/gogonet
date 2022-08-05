package gogonet

import (
	"log"
	"reflect"

	"./marshal"
)

func canCallReflectProcedure(object interface{}, procedureName string) bool {
	return reflect.ValueOf(object).MethodByName(procedureName) != reflect.Value{}
}

func reflectProcedureCall(object interface{}, procedureName string, params []reflect.Value) {
	reflect.ValueOf(object).MethodByName(procedureName).Call(params)
}

func reflectDecodePacketVariables(data []byte) []reflect.Value {

	var paramsCount uint8
	paramsCount, data = marshal.DecodeUint8(data)

	var p interface{}
	result := make([]reflect.Value, paramsCount)

	for i := uint8(0); i < paramsCount; i++ {
		log.Printf("remain: %x\n", data)

		var paramType int32
		paramType, data = marshal.DecodeInt32(data)

		switch paramType {
		case INT:
			p, data = marshal.DecodeInt32(data)
		case BOOL:
			var v int32
			v, data = marshal.DecodeInt32(data)
			p = v > 0
		case STRING:
			p, data = marshal.DecodeString(data)

		}

		result[i] = reflect.ValueOf(p)
	}

	return result
}
