package router

import (
	"fmt"

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
	handler RouteHandler
}

func (t *Trie) Lookup(path string) (RouteHandler, error) {
	for _, v := range t.Nodes {
		if v.data.path == path {
			return v.data.handler, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Not found path on tree,"))
}

func (t *Trie) Construct(routes []*Route) error {
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
