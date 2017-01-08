package router

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k0kubun/pp"
)

func TestServeHTTP(t *testing.T) {
	r := New()
	r.HandleFunc("/user", userHandler)
	r.HandleFunc("/shop", shopHandler)
	r.HandleFunc("/login", loginHandler)
	r.Construct()

	http.Handle("/", r)
	ts := httptest.NewServer(r)

	res, err := http.Get(ts.URL + "/login")
	defer res.Body.Close()
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}
	body, _ := ioutil.ReadAll(res.Body)
	pp.Println(string(body))
}

func userHandler()  {}
func shopHandler()  {}
func loginHandler() {}
