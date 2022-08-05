package gogonet

type INativeMethod interface {
	New() INativeMethod
	SetOwnerNode(INode)

	Name() string

	Call()
	Unmarshal(*StreamReader)
}

func canCallNativeProcedure(node INode, procedureName string) bool {
	n := node.node()
	_, ok := n.nativeRpc[procedureName]
	return ok
}

func getNativeProcedure(node INode, procedureName string) INativeMethod {
	n := node.node()
	return n.nativeRpc[procedureName].New()
}
