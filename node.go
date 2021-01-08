package gogonet

import (
	"strings"

	"github.com/themrviper/gogonet/utils"
)

type PathSentCache struct {
	id              int
	confirmed_peers map[int]bool
}

type NodePath string

type NodeInfo struct {
	path     NodePath
	instance string
}

type PathGetCache struct {
	nodes map[int]NodeInfo
}

type Node struct {
	name string

	parent *Node
	childs []*Node

	rpc_handlers map[string]IRPC
}

var root_node = newNode("root")
var rpc_dummy_node = newNode("dummy")

func GetRootNode() *Node {
	return root_node
}

func newNode(name string) *Node {
	return &Node{
		name: name,

		childs: make([]*Node, 0, 0),

		rpc_handlers: make(map[string]IRPC),
	}
}

func (n *Node) NewNode(path string) (r *Node) {

	paths := strings.Split(path, "/")
	utils.IfPanic(len(paths) < 1, "Cannot create new node, invalid path")

	if paths[0] == "" {
		// absolute node, go to root and make path relative
		utils.IfPanic(len(paths) < 2, "Cannot create new node, invalid path")

		return n.GetRootNode().NewNode(strings.Join(paths[1:], "/"))
	}

	for len(paths) > 0 {

		utils.IfPanic(len(paths[0]) < 1, "Cannot create new node, invalid path")

		r = newNode(paths[0])
		r.parent = n
		n.AppendChild(r)

		n = r
		paths = paths[1:]
	}
	return
}

func (n *Node) GetOrNewNode(path string) *Node {
	if node := n.GetNode(path); node != nil {
		return node
	}

	return n.NewNode(path)
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) Path() (result string) {
	for n.parent != nil {
		result = "/" + n.name + result
		n = n.parent
	}

	result = "/" + n.name + result
	return
}

func (n *Node) GetNode(path string) *Node {
	paths := strings.Split(path, "/")

	if len(paths) <= 0 {
		panic("Invalid path " + path)
	}

	// If path is absolute, go to top, and find childs
	if paths[0] == "" {

		// also skip root node
		// so this paths will be the same
		// /root/Button
		// /Button
		if len(paths[1:]) > 1 && n.GetRootNode().Name() == paths[1:] {
			paths = paths[1:]
		}
		return n.GetRootNode().GetNode(strings.Join(paths[1:], "/"))
	}

	// If path is relative, search right here
	for _, node := range n.childs {
		if node.Name() == paths[0] {

			if len(paths) > 1 {
				return node.GetNode(strings.Join(paths[1:], "/"))
			}

			return node
		}
	}

	return nil
}

func (n *Node) GetRootNode() *Node {
	if n.parent == nil {
		return n
	}

	return n.parent
}

func (n *Node) AppendChild(node *Node) {
	if node == nil {
		panic("Node cant be nil")
	}

	if n.GetNode(node.name) != nil {
		panic("Node already added, cant duplicate node")
	}

	n.childs = append(n.childs, node)
}

func (n *Node) RPCHandler(name string, handler IRPC) {
	if _, ok := n.rpc_handlers[name]; ok {
		panic("[node] Handler " + name + " already registered, cant register it twice")
	}

	n.rpc_handlers[name] = handler
}

func (n *Node) GetRPCHandler(name string) (IRPC, bool) {
	if handler, ok := n.rpc_handlers[name]; ok {
		return handler.New(), true
	}

	return nil, false
}
