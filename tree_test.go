package router

import "testing"

var fixtureMap map[string]*Node

// fixture means sample URL list:
//	"GET"
//		/user/list
//		/user/:userID
//		/user/:userID/follow
//		/shop/:shopID/detail
//		/shop/:shopID/:paymentID
//
//	return teadown func
func setupFixture() func() {
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

	// TODO: Trie.Rootの代わりをどうするか考える
	get := &Node{
		data:  &Data{key: "GET"},
		bros:  nil,
		child: root,
	}

	fixtureMap = map[string]*Node{
		"root":   root,
		"user1a": user1a, "user2a": user2a, "user2b": user2b, "user3a": user3a,
		"shop1a": shop1a, "shop2a": shop2a, "shop3a": shop3a, "shop3b": shop3b,
	}

	// expect restructure fixture
	return func() { setupFixture() }
}

func TestGetBros(t *testing.T) {
	teardown := setupFixture()
	defer teardown()

	cases := []struct {
		start      *Node
		input      string
		expectNode *Node
		expectBool bool
	}{
		{fixtureMap["user2a"], ":", fixtureMap["user2b"], true},
		{fixtureMap["shop3a"], ":", fixtureMap["shop3b"], true},
		{fixtureMap["shop2a"], "detail", nil, false},
	}
	for i, c := range cases {
		result, ok := c.start.getBros(c.input)
		if result != c.expectNode || c.expectBool != ok {
			t.Errorf("#%d: want result:%#v ok:%t, got result:%#v ok:%t", i, c.expectNode, c.expectBool, result, ok)
		}
	}
}

func TestGetChild(t *testing.T) {
	teardown := setupFixture()
	defer teardown()

	cases := []struct {
		start      *Node
		input      string
		expectNode *Node
		expectBool bool
	}{
		{fixtureMap["root"], "user", fixtureMap["user1a"], true},
		{fixtureMap["root"], "none", nil, false},
		{fixtureMap["root"], ":", nil, false},
	}
	for i, c := range cases {
		result, ok := c.start.getChild(c.input)
		if result != c.expectNode || c.expectBool != ok {
			t.Errorf("#%d: want result:%#v ok:%t, got result:%#v ok:%t", i, c.expectNode, c.expectBool, result, ok)
		}
	}
}

func TestGetLastBros(t *testing.T) {
	teardown := setupFixture()
	defer teardown()

	cases := []struct {
		start      *Node
		expectNode *Node
	}{
		{fixtureMap["user1a"], fixtureMap["shop1a"]},
		{fixtureMap["user3a"], fixtureMap["user3a"]},
	}
	for i, c := range cases {
		result := c.start.getLastBros()
		if result != c.expectNode {
			t.Errorf("#%d: want result:%#v, got result:%#v", i, c.expectNode, result)
		}
	}
}
