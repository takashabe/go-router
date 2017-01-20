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
		data:  &Data{key: ":", path: "/shop/:shopID/:paymentID", handler: nil},
		bros:  nil,
		child: nil,
	}
	shop3a := &Node{
		data:  &Data{key: "detail", path: "/shop/:shopID/detail", handler: nil},
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

func TestSetChild(t *testing.T) {
	expectTrie1, nodes1 := generateFixture()
	inputNode1 := Node{data: &Data{key: ":", path: "/user/:userID/:attrID"}}
	nodes1["user3a"].bros = &inputNode1

	expectTrie2, nodes2 := generateFixture()
	inputNode2 := Node{data: &Data{key: ":", path: "/shop/:shopID/:paymentID/:dummyID"}}
	nodes2["shop3b"].child = &inputNode2

	expectTrie3, _ := generateFixture()
	inputNode3 := Node{data: &Data{key: ":", path: "/user/:userID"}}

	cases := []struct {
		start       string
		input       Node
		expectError error
		expectTree  Trie
	}{
		{"user2b", inputNode1, nil, expectTrie1},
		{"shop3b", inputNode2, nil, expectTrie2},
		{"user1a", inputNode3, ErrAlreadyPathRegistered, expectTrie3},
	}
	for i, c := range cases {
		setupFixture()
		_, err := helperNodes[c.start].setChild(c.input)
		if err != c.expectError {
			t.Errorf("#%d: want error:%v, got error:%v", i, c.expectError, err)
		}
		if !reflect.DeepEqual(c.expectTree, fixtureTrie) {
			t.Errorf("#%d: want tree:%#v, got tree:%#v", i, c.expectTree, fixtureTrie)
		}
	}
}

func TestFind(t *testing.T) {
	setupFixture()

	cases := []struct {
		input       string
		expectNode  *Node
		expectError error
	}{
		{"/user/:userID/follow", helperNodes["user3a"], nil},
		{"/user/:userID/follow/none", nil, ErrPathNotFound},
		{"/shop/:shopID/detail", helperNodes["shop3a"], nil},
	}
	for i, c := range cases {
		result, err := fixtureTrie.find(c.input, "GET")
		if result != c.expectNode {
			t.Errorf("#%d: want result:%#v , got result:%#v ", i, c.expectNode, result)
		}
		if err != c.expectError {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expectError, err)
		}
	}
}

func TestInsert(t *testing.T) {
	expectTrie1, nodes1 := generateFixture()
	nodes1["user3a"].bros = &Node{data: &Data{key: "dummy", path: "/user/:userID/dummy"}}

	expectTrie2, nodes1 := generateFixture()
	nodes1["shop3b"].child = &Node{data: &Data{key: ":", path: "/shop/:shopID/:paymentID/:dummyID"}}

	expectTrie3, nodes1 := generateFixture()
	node3 := &Node{
		data:  &Data{key: "/"},
		child: &Node{data: &Data{key: "post", path: "/post"}},
	}
	expectTrie3.root["POST"] = node3

	cases := []struct {
		inputPath    string
		inputMethod  string
		expectResult error
		expectTree   Trie
	}{
		{"/user/:userID/dummy", "GET", nil, expectTrie1},
		{"/shop/:shopID/:paymentID/:dummyID", "GET", nil, expectTrie2},
		{"/post", "POST", nil, expectTrie3},
	}
	for i, c := range cases {
		setupFixture()
		result := fixtureTrie.insert(c.inputPath, c.inputMethod, nil)
		if result != c.expectResult {
			t.Errorf("#%d: want result:%#v, got result:%#v", i, c.expectResult, result)
		}
		if !reflect.DeepEqual(c.expectTree, fixtureTrie) {
			t.Errorf("#%d: want tree:%#v, got tree:%#v", i, c.expectTree, fixtureTrie)
		}
	}
}
