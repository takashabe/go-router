package router

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrPathNotFound          = errors.New("path not found")
	ErrInvalidPathFormat     = errors.New("invalid path format")
	ErrAlreadyPathRegistered = errors.New("already path registered")
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

func NewTrie() Trie {
	// cap num refers to: net/http/method.go
	return Trie{root: make(map[string]*Node, 9)}
}

func (t *Trie) Lookup(path string, method string) (HandlerData, error) {
	return HandlerData{}, ErrPathNotFound
}

func (t *Trie) Construct(routes []*Route) error {
	for _, r := range routes {
		err := t.insert(r.path, r.method, r.handler)
		if err != nil {
			return errors.Wrap(err, "failed construct tree")
		}
	}
	return nil
}

func (t *Trie) find(path string, method string) (*Node, error) {
	path, err := validatePath(path)
	if err != nil {
		return nil, err
	}

	dst, ok := t.root[method]
	if !ok {
		return nil, ErrPathNotFound
	}

	parts := strings.Split(path, "/")
	// exclude "/"
	parts = parts[1:]
	for _, p := range parts {
		p = convertParamKey(p)
		if n, ok := dst.getChild(p); ok {
			dst = n
		}
		if dst.data.path == path {
			return dst, nil
		}
	}

	return nil, ErrPathNotFound
}

func (t *Trie) insert(path, method string, handler baseHandler) error {
	path, err := validatePath(path)
	if err != nil {
		return err
	}

	dst, ok := t.root[method]
	if !ok {
		t.root[method] = &Node{data: &Data{key: "/"}}
		dst = t.root[method]
	}

	parts := strings.Split(path, "/")
	// exclude "/"
	parts = parts[1:]
	for i, p := range parts {
		p = convertParamKey(p)
		if n, ok := dst.getChild(p); ok {
			if len(parts)-1 == i {
				return ErrAlreadyPathRegistered
			}
			dst = n
			continue
		}
		data := Data{key: p}
		// leaf node
		if len(parts)-1 == i {
			data.path = path
			data.handler = handler
		}
		dst, err = dst.setChild(Node{data: &data})
		if err != nil {
			return err
		}
	}
	return nil
}

func validatePath(s string) (string, error) {
	if string(s[0]) != "/" {
		return "", ErrInvalidPathFormat
	}
	if string(s[len(s)-1]) == "/" {
		return s[:len(s)-1], nil
	}
	return s, nil
}

func convertParamKey(s string) string {
	if string(s[0]) == ":" {
		return ":"
	}
	return s
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

func (n *Node) setChild(node Node) (*Node, error) {
	if _, ok := n.getChild(node.data.key); ok {
		return nil, ErrAlreadyPathRegistered
	}

	if n.child == nil {
		n.child = &node
	} else {
		n.child.getLastBros().bros = &node
	}
	return &node, nil
}

func (n *Node) getLastBros() *Node {
	if n.bros == nil {
		return n
	}
	return n.bros.getLastBros()
}
