package router

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrPathNotFound      = errors.New("path not found")
	ErrInvalidPathFormat = errors.New("invalid path format")
)

type Trie struct {
	root map[string]*Node
}

type Node struct {
	data  *Data
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
	return HandlerData{}, ErrPathNotFound
}

func (t *Trie) Construct(routes []*Route) error {
	return nil
}

func (t *Trie) find(path string) (*Node, error) {
	if string(path[0]) != "/" {
		return nil, ErrInvalidPathFormat
	}
	return nil, ErrInvalidPathFormat
}

func (t *Trie) insert(path string, handler baseHandler) error {
	if string(path[0]) != "/" {
		return ErrPathNotFound
	}

	dst, ok := t.root["GET"]
	if !ok {
		t.root["GET"] = &Node{data: &Data{key: "/"}}
		dst = t.root["GET"]
	}

	// traverse tree and find the insertion point of node
	parts := strings.Split(path, "/")
	for i, part := range parts {
		log.Printf("dst.data.key=%s, part=%s\n", dst.data.key, part)
		if len(part) == 0 {
			continue
		}
		if dst.data.key == part {
			continue
		}
		if child, ok := dst.getChild(part); ok {
			dst = child
		}

		// param path
		if string(part[0]) == ":" {
			part = string(part[0])
		}
		data := Data{key: part}

		// if Leaf node, add params
		if len(parts) == i-1 {
			data.path = path
			data.handler = handler
		}
		node := Node{
			data:  &data,
			bros:  nil,
			child: nil,
		}
		dst.setChild(node)
	}
	return nil
}

func (n *Node) getChild(key string) (*Node, bool) {
	if n.child == nil {
		return nil, false
	}

	child := n.child
	if child.data.key == key {
		return child, true
	}
	if bros, ok := child.getBros(key); ok {
		return bros, true
	}

	return nil, false
}

func (n *Node) getBros(key string) (*Node, bool) {
	if n.bros == nil {
		return nil, false
	}

	bros := n.bros
	if bros.data.key == key {
		return bros, true
	}
	return bros.getBros(key)
}

func (n *Node) setChild(node Node) error {
	log.Printf("setChild: n.data.key=%s, node.data.key=%s\n", n.data.key, node.data.key)
	if _, ok := n.getChild(node.data.path); ok {
		return errors.New(fmt.Sprintf("already registered node. path:%s\n", node.data.path))
	}

	if n.child == nil {
		n.child.getLastBros().bros = &node
	} else {
		n.child = &node
	}
	return nil
}

func (n *Node) getLastBros() *Node {
	if n.bros == nil {
		return n
	}
	return n.bros.getLastBros()
}
