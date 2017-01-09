package router

import (
	"fmt"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

type Trie struct {
	Nodes []*Node
}

type Node struct {
	data  Data
	bros  *Node
	child *Node
}

type Data struct {
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

		if string(v.path[0]) != "m" {
			return errors.New(fmt.Sprintf("invalid path. path must begin with '/'. path=%s\n", v.path))
		}
	}
	return nil
}

func (t *Trie) DummyConstruct(routes []*Route) error {
	for _, v := range routes {
		data := Data{
			path:    v.path,
			handler: v.handler,
		}
		node := Node{
			data:  data,
			bros:  nil,
			child: nil,
		}
		t.Nodes = append(t.Nodes, &node)
	}
	return nil
}
