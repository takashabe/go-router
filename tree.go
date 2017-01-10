package router

import (
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

type Trie struct {
	Nodes []*Node
	Root  map[string]*Node
}

type Node struct {
	data  Data
	bros  *Node
	child *Node
}

type Data struct {
	// tree node key
	key string
	// origin URL path
	path    string
	handler baseHandler
}

func (t *Trie) Lookup(path string) (HandlerData, error) {
	for _, v := range t.Nodes {
		if v.data.path == path {
			return HandlerData{handler: v.data.handler, params: nil}, nil
		}
	}
	return HandlerData{}, errors.New(fmt.Sprintf("Not found path on tree,"))
}

func (t *Trie) Construct(routes []*Route) error {
	t.Nodes = make([]*Node, len(routes))
	pp.Printf("#1 t.Nodes: %v\n", t.Nodes)
	for i, v := range routes {
		pp.Printf("LOOP: #%d, %v\n", i, v)

		if string(v.path[0]) != "/" {
			return errors.New(fmt.Sprintf("invalid path. path must begin with '/'. path:%s\n", v.path))
		}
		parts := string.Split(v.path, "/")
		for k, p := range parts {
		}
	}
	return nil
}

func (t *Trie) insert(n Node) error {
	dst, ok := t.Root["GET"]
	if !ok {
		t.Root["GET"] = &Node{}
		dst = t.Root["GET"]
	}

	// traverse tree and find the insertion point of node
}

func (n *Node) getChild(key string) (Node, bool) {
	if n.child == nil {
		return nil, false
	}

	child := n.child
	for {
		if child.data.path == key {
			return child, true
		}
		child = child.bros
	}
	return nil, true
}

func (n *Node) setChild(node Node) error {
	if _, ok := n.getChild(node.data.path); ok {
		return errors.New(fmt.Sprintf("already registered node. path:%s\n", node.data.path))
	}
	n.child = node
	return nil
}
