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

func NewTrie() *Trie {
	// cap num refers to: net/http/method.go
	// without "CONNECT", "TRACE"
	return &Trie{root: make(map[string]*Node, 7)}
}

func (t *Trie) Lookup(path string, method string) (HandlerData, error) {
	n, err := t.find(path, method)
	if err != nil {
		return HandlerData{}, errors.Wrapf(err, "failed lookup. path=%s method=%s", path, method)
	}
	return HandlerData{
		handler: n.data.handler,
		params:  n.exportParam(path),
	}, nil
}

func (t *Trie) find(path string, method string) (*Node, error) {
	parts, err := generateSplitPath(path)
	if err != nil {
		return nil, err
	}

	dst, ok := t.root[method]
	if !ok {
		return nil, ErrPathNotFound
	}

	for _, p := range parts {
		if n, ok := dst.getChild(p); ok {
			dst = n
		}
		if dst.pathEqual(path) {
			return dst, nil
		}
	}

	return nil, ErrPathNotFound
}

func (t *Trie) Insert(method, path string, handler baseHandler) error {
	parts, err := generateSplitPath(path)
	if err != nil {
		return errors.Wrapf(err, "failed insert. path=%s, method=%s", path, method)
	}

	dst, ok := t.root[method]
	if !ok {
		t.root[method] = &Node{data: &Data{key: "/"}}
		dst = t.root[method]
	}

	// insert "/"
	if dst.data.key == path {
		dst.data = &Data{
			key:     parts[0],
			path:    path,
			handler: handler,
		}
		return nil
	}

	// exclude "/"
	parts = parts[1:]
	for i, p := range parts {
		p = convertParamKey(p)
		if n, ok := dst.getChild(p); ok {
			if len(parts)-1 == i {
				// exist node, but yet registered path and handler
				if n.data.path == "" {
					n.data.path = path
					n.data.handler = handler
					return nil
				}
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
			return errors.Wrapf(err, "failed insert. path=%s, method=%s", path, method)
		}
	}
	return nil
}

func generateSplitPath(s string) ([]string, error) {
	s, err := validatePath(s)
	if err != nil {
		return nil, err
	}

	ds := strings.Split(s, "/")
	ds[0] = "/"
	return ds, nil
}

// check valid path. path is must be begin "/"
// if path "/foo/bar/.../" trim last "/"
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
	if isParamKey(s) {
		return ":"
	}
	return s
}

func isParamKey(s string) bool {
	return string(s[0]) == ":"
}

func (n *Node) getChild(key string) (*Node, bool) {
	if n.child == nil {
		return nil, false
	}

	// search match key node
	child := n.child
	if child.data.key == key {
		return child, true
	}
	if bros, ok := child.getBros(key); ok {
		return bros, true
	}

	// search param node
	if childParam, ok := n.getChildParam(); ok {
		return childParam, true
	}

	return nil, false
}

func (n *Node) getChildParam() (*Node, bool) {
	if n.child == nil {
		return nil, false
	}

	child := n.child
	if isParamKey(child.data.key) {
		return child, true
	}
	if bros, ok := child.getBrosParam(); ok {
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

func (n *Node) getBrosParam() (*Node, bool) {
	if n.bros == nil {
		return nil, false
	}

	bros := n.bros
	if isParamKey(bros.data.key) {
		return bros, true
	}
	return bros.getBrosParam()
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

func (n *Node) pathEqual(path string) bool {
	if len(n.data.path) == 0 {
		return false
	}
	a, err := generateSplitPath(n.data.path)
	if err != nil {
		return false
	}
	b, err := generateSplitPath(path)
	if err != nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		// param symbol is match any
		if len(v) != 0 && isParamKey(v) {
			continue
		}
		if v != string(b[i]) {
			return false
		}
	}
	return true
}

func (n *Node) exportParam(path string) []interface{} {
	p := []interface{}{}
	if len(n.data.path) == 0 {
		return p
	}
	a, err := generateSplitPath(n.data.path)
	if err != nil {
		return p
	}
	b, err := generateSplitPath(path)
	if err != nil {
		return p
	}
	if len(a) != len(b) {
		return p
	}

	for i, v := range a {
		if len(v) != 0 && isParamKey(v) {
			p = append(p, b[i])
		}
	}
	return p
}
