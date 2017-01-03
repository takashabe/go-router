package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	r := New()
	http.Handle("/", r)
	ts := httptest.NewServer(r)

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}
}
