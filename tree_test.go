package router

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
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
		if !reflect.DeepEqual(result, c.expectNode) || c.expectBool != ok {
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
		{helperNodes["user1a"], "10", helperNodes["user2b"], true},
	}
	for i, c := range cases {
		result, ok := c.start.getChild(c.input)
		if !reflect.DeepEqual(result, c.expectNode) || c.expectBool != ok {
			t.Errorf("#%d: want result:%#v ok:%t, got result:%#v ok:%t", i, c.expectNode, c.expectBool, result, ok)
		}
	}
}

func TestGetBrosParam(t *testing.T) {
	setupFixture()

	cases := []struct {
		start      *Node
		expectNode *Node
		expectBool bool
	}{
		{helperNodes["user2a"], helperNodes["user2b"], true},
		{helperNodes["shop3a"], helperNodes["shop3b"], true},
		{helperNodes["shop2a"], nil, false},
	}
	for i, c := range cases {
		result, ok := c.start.getBrosParam()
		if !reflect.DeepEqual(result, c.expectNode) || c.expectBool != ok {
			t.Errorf("#%d: want result:%#v ok:%t, got result:%#v ok:%t", i, c.expectNode, c.expectBool, result, ok)
		}
	}
}

func TestGetChildParam(t *testing.T) {
	setupFixture()

	cases := []struct {
		start      *Node
		expectNode *Node
		expectBool bool
	}{
		{helperNodes["shop1a"], helperNodes["shop2a"], true},
		{helperNodes["shop2a"], helperNodes["shop3b"], true},
		{helperNodes["root"], nil, false},
		{helperNodes["shop3b"], nil, false},
	}
	for i, c := range cases {
		result, ok := c.start.getChildParam()
		if !reflect.DeepEqual(result, c.expectNode) || c.expectBool != ok {
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
		if !reflect.DeepEqual(result, c.expectNode) {
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

func TestPathEqual(t *testing.T) {
	cases := []struct {
		baseNode *Node
		input    string
		expect   bool
	}{
		{
			&Node{data: &Data{path: "/user/:userID/follow"}},
			"/user/10/follow/",
			true,
		},
		{
			&Node{data: &Data{path: "/user/:userID/follow"}},
			"user/10/follow/",
			false,
		},
		{
			&Node{data: &Data{path: "/user/:userID/follow"}},
			"/user/10/follow/10",
			false,
		},
		{
			&Node{data: &Data{path: "/user/:userID/follow"}},
			"/user/10/dummy/",
			false,
		},
	}
	for i, c := range cases {
		result := c.baseNode.pathEqual(c.input)
		if result != c.expect {
			t.Errorf("#%d: want:%#v , got:%#v ", i, c.expect, result)
		}
	}
}

func TestFind(t *testing.T) {
	setupFixture()

	cases := []struct {
		inputPath   string
		inputMethod string
		expectNode  *Node
		expectError error
	}{
		{"/user/1/follow", "GET", helperNodes["user3a"], nil},
		{"/user/1/follow/none", "GET", nil, ErrPathNotFound},
		{"/shop/1/detail", "POST", nil, ErrPathNotFound},
		{"shop/1/detail", "GET", nil, ErrInvalidPathFormat},
	}
	for i, c := range cases {
		result, err := fixtureTrie.find(c.inputPath, c.inputMethod)
		if err != c.expectError {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expectError, err)
		}
		if !reflect.DeepEqual(result, c.expectNode) {
			t.Errorf("#%d: want result:%#v , got result:%#v ", i, c.expectNode, result)
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

	expectTrie, _ := generateFixture()

	cases := []struct {
		inputPath    string
		inputMethod  string
		expectResult error
		expectTree   Trie
	}{
		{"/user/:userID/dummy", "GET", nil, expectTrie1},
		{"/shop/:shopID/:paymentID/:dummyID", "GET", nil, expectTrie2},
		{"/post", "POST", nil, expectTrie3},
		{"/user/:userID", "GET", ErrAlreadyPathRegistered, expectTrie},
		{"user", "GET", ErrInvalidPathFormat, expectTrie},
	}
	for i, c := range cases {
		setupFixture()
		result := fixtureTrie.Insert(c.inputMethod, c.inputPath, nil)
		if result != c.expectResult {
			t.Errorf("#%d: want result:%#v, got result:%#v", i, c.expectResult, result)
		}
		if !reflect.DeepEqual(c.expectTree, fixtureTrie) {
			t.Errorf("#%d: want tree:%#v, got tree:%#v", i, c.expectTree, fixtureTrie)
		}
	}
}

func TestConstruct(t *testing.T) {
	setupFixture()

	input1 := []*Route{
		&Route{path: "/user/list", method: "GET"},
		&Route{path: "/user/:userID", method: "GET"},
		&Route{path: "/user/:userID/follow", method: "GET"},
		&Route{path: "/shop/:shopID/detail", method: "GET"},
		&Route{path: "/shop/:shopID/:paymentID", method: "GET"},
	}

	input2 := []*Route{
		&Route{path: "user/list", method: "GET"},
	}

	cases := []struct {
		input       []*Route
		expectError error
		expectTree  *Trie
	}{
		{input1, nil, &fixtureTrie},
		{input2, ErrInvalidPathFormat, NewTrie()},
	}
	for i, c := range cases {
		trie := NewTrie()
		err := trie.Construct(c.input)
		if errors.Cause(err) != c.expectError {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expectError, err)
		}
		if !reflect.DeepEqual(c.expectTree, trie) {
			t.Errorf("#%d: want result:%#v , got result:%#v ", i, c.expectTree, trie)
		}
	}
}

func TestExportParam(t *testing.T) {
	cases := []struct {
		baseNode *Node
		input    string
		expect   []interface{}
	}{
		{
			&Node{data: &Data{path: "/user/:userID/follow/:attrID"}},
			"/user/10/follow/20/",
			[]interface{}{"10", "20"},
		},
		{
			&Node{data: &Data{path: "/user/list"}},
			"/user/list",
			[]interface{}{},
		},
		{
			&Node{data: &Data{path: "/user/:userID/follow"}},
			"/user/10/follow/10",
			[]interface{}{},
		},
		{
			&Node{data: &Data{path: "/user/:userID/follow"}},
			"user/10/follow/10",
			[]interface{}{},
		},
		{
			&Node{data: &Data{}},
			"/user/10/follow/10",
			[]interface{}{},
		},
	}
	for i, c := range cases {
		result := c.baseNode.exportParam(c.input)
		if !reflect.DeepEqual(result, c.expect) {
			t.Errorf("#%d: want:%#v , got:%#v ", i, c.expect, result)
		}
	}
}

func TestLookup(t *testing.T) {
	setupFixture()

	cases := []struct {
		inputPath   string
		inputMethod string
		expectData  HandlerData
		expectError error
	}{
		{
			"/shop/10/20/",
			"GET",
			HandlerData{handler: nil, params: []interface{}{"10", "20"}},
			nil,
		},
		{
			"/shop/10/20/30/",
			"GET",
			HandlerData{},
			ErrPathNotFound,
		},
	}
	for i, c := range cases {
		result, err := fixtureTrie.Lookup(c.inputPath, c.inputMethod)
		if errors.Cause(err) != c.expectError {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expectError, err)
		}
		if !reflect.DeepEqual(result, c.expectData) {
			t.Errorf("#%d: want result:%#v , got result:%#v ", i, c.expectData, result)
		}
	}
}
