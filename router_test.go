package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k0kubun/pp"
)

func TestServeHTTP(t *testing.T) {
	r := New()
	r.HandleFunc("/user", UserHandler)
	r.HandleFunc("/shop", ShopHandler)
	r.HandleFunc("/login", LoginHandler)
	err := r.Construct()
	if err != nil {
		t.Errorf(err.Error())
	}

	http.Handle("/", r)
	ts := httptest.NewServer(r)

	res, err := http.Get(ts.URL + "/user")
	defer res.Body.Close()
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}
	pp.Println(res.Header)
	body, _ := ioutil.ReadAll(res.Body)
	pp.Println(string(body))
}

// Dummy handlers
func UserHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Im user")
}
func ShopHandler(w http.ResponseWriter, req *http.Request, id int) {
	fmt.Fprintf(w, "Im shop. id:%d", id)
}
func LoginHandler(w http.ResponseWriter, req *http.Request, signature string, password string) {
	fmt.Fprintf(w, "Im login. signature:%s, password:%s", signature, password)
}
