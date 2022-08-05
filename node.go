package gogonet

import (
	"log"
	"strings"
	"sync/atomic"
	"time"

	"./signals"
	"./utils"
)

type Node struct {
	name string

	instanceId uint32

	parent INode
	childs map[uint32]INode

	nativeRpc map[string]INativeMethod

	multiplayerAPI *MultiplayerAPI
}

//
// tree stuff
//

var tree INode
var lastInstanceId uint32

func init() {
	tree = NewNode("/root")

	tree.(*Node).multiplayerAPI = &MultiplayerAPI{
		signals:        signals.New(),
		connectedPeers: make(map[uint32]bool),
		recvPathCache:  make(map[uint32]map[uint32]string),

		sentPathCache: make(map[string]*SentPathCache),
	}
}

type ITree interface {
	INode

	SetScene(scene INode)
	SetNetworkPeer(peer INetworkPeer)
	ListenAndServe()
}

func GetTree() ITree {
	return tree.(*Node)
}

func (t *Node) SetScene(scene INode) {
	t.childs = make(map[uint32]INode)
	t.childs[scene.InstanceID()] = scene
}

func (t *Node) SetNetworkPeer(peer INetworkPeer) {
	t.multiplayerAPI.SetNetworkPeer(peer)
}

func (t *Node) ListenAndServe() {
	go t.multiplayerAPI.ListenAndServe()

	fps := 60
	prev := time.Now()

	ticker := time.NewTicker(time.Second / time.Duration(fps))

	for now := range ticker.C {

		prev = now
	}
}

//
// scene stuff
//

var sceneRepo = NewNode("/scene_repo")

type IScene interface {
	INode

	Instance() INode
}

func GetScene(name string) IScene {
	return sceneRepo.GetNode(name).(*Node)
}

func NewScene(name string) IScene {
	return sceneRepo.NewNode(name).(*Node)
}

func GetOrNewScene(name string) IScene {
	return sceneRepo.GetOrNewNode(name).(*Node)
}

// Lets think that, /root is scene tree, like in godot, so we can call instance() method
func (n *Node) Instance() INode {
	// TODO
	utils.Log(6, n.parent)
	utils.IfPanic(n.parent != sceneRepo, "Cannot create instance of this scene, maybe its node?")

	node := n.clone()
	node.instanceId = nodeGenerateId()

	return node
}

func (n *Node) clone() *Node {
	node := NewNode(n.name).node()

	node.nativeRpc = n.nativeRpc
	for id, child := range n.childs {
		node.childs[id] = child.node().clone()
	}

	return node
}

//
// node stuff
//

type INode interface {
	node() *Node
	Name() string
	Path() string
	InstanceID() uint32

	NewNode(path string) INode
	GetNode(path string) INode
	GetOrNewNode(path string) INode

	AppendChild(node INode)
	AddNativeRPCMethod(method INativeMethod)

	Rpc(procedureName string, params ...interface{})
	RpcId(id int32, procedureName string, params ...interface{})

	RpcUnreliable(procedureName string, params ...interface{})
	RpcUnreliableId(id int32, procedureName string, params ...interface{})
}

func NewNode(name string) INode {
	return &Node{
		name: name,

		instanceId: nodeGenerateId(),

		childs: make(map[uint32]INode),

		nativeRpc: make(map[string]INativeMethod),
	}
}

func nodeGenerateId() uint32 {
	atomic.AddUint32(&lastInstanceId, 1)
	return lastInstanceId
}

func (n *Node) node() *Node {
	return n
}

func (n *Node) Name() string {
	return n.node().name
}

func (n *Node) Path() (result string) {

	log.Println(n.parent)
	for n.parent != nil {
		result = "/" + n.name + result
		n = n.parent.node()
	}

	return
}

func (n *Node) InstanceID() uint32 {
	return n.instanceId
}

// Create child node, or if path is absolute, create node for this path
func (n *Node) NewNode(path string) (r INode) {

	paths := strings.Split(path, "/")
	utils.IfPanic(len(paths) < 1, "Cannot create new node, invalid path")

	if paths[0] == "" {
		// absolute node, go to root and make path relative
		utils.IfPanic(len(paths) < 2, "Cannot create new node, invalid path")

		if len(paths[1:]) > 1 && GetTree().Name() == paths[1] {
			paths = paths[1:]
		}
		return GetTree().NewNode(strings.Join(paths[1:], "/"))
	}

	for len(paths) > 0 {

		utils.IfPanic(len(paths[0]) < 1, "Cannot create new node, invalid path")

		r = NewNode(paths[0])
		r.node().parent = n
		n.AppendChild(r)

		n = r.node()
		paths = paths[1:]
	}
	return
}

func (n *Node) GetNode(path string) INode {
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
		if len(paths[1:]) > 1 && GetTree().Name() == paths[1] {
			paths = paths[1:]
		}
		return GetTree().GetNode(strings.Join(paths[1:], "/"))
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

func (n *Node) GetOrNewNode(path string) INode {
	if node := n.GetNode(path); node != nil {
		return node
	}

	return n.NewNode(path)
}

func (n *Node) AppendChild(node INode) {
	if node == nil {
		panic("Node cant be nil")
	}

	if _, ok := n.childs[node.InstanceID()]; ok {
		return
	}

	node.node().parent = n
	node.node().multiplayerAPI = n.multiplayerAPI
	n.childs[node.InstanceID()] = node
}

//
// rpc stuff
//

func (n *Node) AddNativeRPCMethod(method INativeMethod) {
	n.nativeRpc[method.Name()] = method
}

func (n *Node) Rpc(procedureName string, params ...interface{})             {}
func (n *Node) RpcId(id int32, procedureName string, params ...interface{}) {}

func (n *Node) RpcUnreliable(procedureName string, params ...interface{})             {}
func (n *Node) RpcUnreliableId(id int32, procedureName string, params ...interface{}) {}
