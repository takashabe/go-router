package router

import (
	"testing"

	"github.com/k0kubun/pp"
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
	// defined the sample URL in reverse order
	shop3b := &Node{
		data:  &Data{key: ":", path: "shop/:shopID/:paymentID", handler: func() {}},
		bros:  nil,
		child: nil,
	}
	shop3a := &Node{
		data:  &Data{key: "detail", path: "shop/:shopID/detail", handler: func() {}},
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
		data:  &Data{key: "follow", path: "/user/:userID/follow", handler: func() {}},
		bros:  nil,
		child: nil,
	}
	user2b := &Node{
		data:  &Data{key: ":", path: "/user/:userID", handler: func() {}},
		bros:  nil,
		child: user3a,
	}
	user2a := &Node{
		data:  &Data{key: "list", path: "/user/list", handler: func() {}},
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

	fixtureTrie = Trie{root: map[string]*Node{"GET": root}}
	helperNodes = map[string]*Node{
		"root":   root,
		"user1a": user1a, "user2a": user2a, "user2b": user2b, "user3a": user3a,
		"shop1a": shop1a, "shop2a": shop2a, "shop3a": shop3a, "shop3b": shop3b,
	}
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

	a := &fixtureTrie

	b := *a
	b.insert("/user/:id/:hoge", func() {})
	pp.Println(a)
	pp.Println(b)

	cases := []struct {
		input        string
		expectResult error
		// ポインタ使いすぎかも。メソッドの返り値とか。値型で良いところはもっとそうした方が良さそう
		// でないとTrie全体の比較するのつらくなる
		expectTree *Trie
	}{
		{"/user/:userID/:attrID", nil, nil},
		// {"user/:userID/:attrID", ErrInvalidPathFormat, expectNode1},
	}
	for i, c := range cases {
		result := fixtureTrie.insert(c.input, func() {})
		// pp.Printf("#%d %v", i, fixtureTrie)
		// pp.Printf("Origin=== %v", i, c.expectTree)
		if result != c.expectResult {
			t.Errorf("#%d: want result:%#v, got result:%#v", i, c.expectResult, result)
		}
	}
}
