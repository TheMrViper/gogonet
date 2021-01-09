package gogonet

import "reflect"

func decodePacketVariables(packet []byte) []interface{} {

	paramsCount, packet := decode_uint8(packet)

}

func canCallProcedure(object interface{}, procedureName string) bool {
	return reflect.ValueOf(object).MethodByName(procedureName) != nil
}

func reflectProcedureCall(object interface{}, procedureName string, params []interface{}) {
	reflectedParams = reflect_procedure_variables(params)
	reflect.ValueOf(object).MethodByName(procedureName).Call(reflectedParams)
}

func reflectProcedureVariables(v []interface{}) []reflect.Value {
	r := make([]reflect.Value, len(v))
	for i, _ := range v {
		r[i] = reflect.ValueOf(v[i])
	}

	return r
}
