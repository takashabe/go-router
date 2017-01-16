package router

import (
	"reflect"
	"testing"
)

// dummy Trie
var fixtureTrie Trie

// registered tree nodes. expect access from test code
var helperNodes map[string]*Node

// fixture means sample URL list:
//	"GET"
//		/user/list
//		/user/:userID
//		/user/:userID/follow
//		/shop/:shopID/detail
//		/shop/:shopID/:paymentID
func setupFixture() {
	fixtureTrie, helperNodes = generateFixture()
}

func generateFixture() (Trie, map[string]*Node) {
	// defined the sample URL in reverse order
	shop3b := &Node{
		data:  &Data{key: ":", path: "shop/:shopID/:paymentID", handler: nil},
		bros:  nil,
		child: nil,
	}
	shop3a := &Node{
		data:  &Data{key: "detail", path: "shop/:shopID/detail", handler: nil},
		bros:  shop3b,
		child: nil,
	}
	shop2a := &Node{
		data:  &Data{key: ":"},
		bros:  nil,
		child: shop3a,
	}
	shop1a := &Node{
		data:  &Data{key: "shop"},
		bros:  nil,
		child: shop2a,
	}

	user3a := &Node{
		data:  &Data{key: "follow", path: "/user/:userID/follow", handler: nil},
		bros:  nil,
		child: nil,
	}
	user2b := &Node{
		data:  &Data{key: ":", path: "/user/:userID", handler: nil},
		bros:  nil,
		child: user3a,
	}
	user2a := &Node{
		data:  &Data{key: "list", path: "/user/list", handler: nil},
		bros:  user2b,
		child: nil,
	}
	user1a := &Node{
		data:  &Data{key: "user"},
		bros:  shop1a,
		child: user2a,
	}

	root := &Node{
		data:  &Data{key: "/"},
		bros:  nil,
		child: user1a,
	}

	trie := Trie{root: map[string]*Node{"GET": root}}
	nodes := map[string]*Node{
		"root":   root,
		"user1a": user1a, "user2a": user2a, "user2b": user2b, "user3a": user3a,
		"shop1a": shop1a, "shop2a": shop2a, "shop3a": shop3a, "shop3b": shop3b,
	}
	return trie, nodes
}

func TestGetBros(t *testing.T) {
	setupFixture()

	cases := []struct {
		start      *Node
		input      string
		expectNode *Node
		expectBool bool
	}{
		{helperNodes["user2a"], ":", helperNodes["user2b"], true},
		{helperNodes["shop3a"], ":", helperNodes["shop3b"], true},
		{helperNodes["shop2a"], "detail", nil, false},
	}
	for i, c := range cases {
		result, ok := c.start.getBros(c.input)
		if result != c.expectNode || c.expectBool != ok {
			t.Errorf("#%d: want result:%#v ok:%t, got result:%#v ok:%t", i, c.expectNode, c.expectBool, result, ok)
		}
	}
}

func TestGetChild(t *testing.T) {
	setupFixture()

	cases := []struct {
		start      *Node
		input      string
		expectNode *Node
		expectBool bool
	}{
		{helperNodes["root"], "user", helperNodes["user1a"], true},
		{helperNodes["root"], "none", nil, false},
		{helperNodes["root"], ":", nil, false},
	}
	for i, c := range cases {
		result, ok := c.start.getChild(c.input)
		if result != c.expectNode || c.expectBool != ok {
			t.Errorf("#%d: want result:%#v ok:%t, got result:%#v ok:%t", i, c.expectNode, c.expectBool, result, ok)
		}
	}
}

func TestGetLastBros(t *testing.T) {
	setupFixture()

	cases := []struct {
		start      *Node
		expectNode *Node
	}{
		{helperNodes["user1a"], helperNodes["shop1a"]},
		{helperNodes["user3a"], helperNodes["user3a"]},
	}
	for i, c := range cases {
		result := c.start.getLastBros()
		if result != c.expectNode {
			t.Errorf("#%d: want result:%#v, got result:%#v", i, c.expectNode, result)
		}
	}
}

func TestInsert(t *testing.T) {
	setupFixture()

	expectTrie1, nodes := generateFixture()
	nodes["user3a"].bros = &Node{data: &Data{key: ":", path: "/user/:userID/:attrID", handler: nil}}

	cases := []struct {
		input        string
		expectResult error
		expectTree   Trie
	}{
		{"/user/:userID/:attrID", nil, expectTrie1},
		// {"user/:userID/:attrID", ErrInvalidPathFormat, expectNode1},
	}
	for i, c := range cases {
		result := fixtureTrie.insert(c.input, nil)
		if result != c.expectResult {
			t.Errorf("#%d: want result:%#v, got result:%#v", i, c.expectResult, result)
		}

		if !reflect.DeepEqual(c.expectTree, fixtureTrie) {
			t.Errorf("#%d: want tree:%#v, got tree:%#v", i, c.expectTree, fixtureTrie)
			// log.Println(pretty.Compare(fixtureTrie, c.expectTree))
			// pp.Println(fixtureTrie)
			// pp.Println(c.expectTree)
		}
	}
}
