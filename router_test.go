package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func dummyHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "hello, world")
}

func dummyHandlerWithParams(w http.ResponseWriter, req *http.Request, id int, name string) {
	fmt.Fprintf(w, "id=%d, name=%s", id, name)
}

// for debug
func printValues(vs []reflect.Value) {
	for _, v := range vs {
		fmt.Printf("%#v\n", v)
	}
}

func TestParseParams(t *testing.T) {
	cases := []struct {
		input        HandlerData
		expectValues []reflect.Value
		expectError  error
	}{
		{
			HandlerData{
				handler: dummyHandlerWithParams,
				params:  []interface{}{"10", "name"},
			},
			[]reflect.Value{reflect.ValueOf(10), reflect.ValueOf("name")},
			nil,
		},
		{
			HandlerData{
				handler: dummyHandlerWithParams,
				params:  []interface{}{"10"},
			},
			[]reflect.Value{reflect.ValueOf(10), reflect.ValueOf("name")},
			ErrNotFoundHandler,
		},
		{
			HandlerData{
				handler: dummyHandlerWithParams,
				params:  []interface{}{"hoge", "name"},
			},
			[]reflect.Value{reflect.ValueOf(10), reflect.ValueOf("name")},
			ErrInvalidParam,
		},
	}
	for i, c := range cases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		result, err := parseParams(w, r, c.input)
		if errors.Cause(err) != c.expectError {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expectError, err)
		}
		if err == nil {
			// complement a missing variable
			c.expectValues = append(c.expectValues, reflect.ValueOf(w))
			c.expectValues = append(c.expectValues, reflect.ValueOf(r))
			// compare to params
			result = result[2:]
			for vi := 0; vi < len(result); vi++ {
				if result[vi].Interface() != c.expectValues[vi].Interface() {
					t.Errorf("#%d-%d: want result:%#v , got result:%#v ", i, vi, c.expectValues, result)
				}
			}
		}
	}
}

func TestCallHandler(t *testing.T) {
	cases := []struct {
		input  HandlerData
		expect error
	}{
		{
			HandlerData{
				handler: dummyHandlerWithParams,
				params:  []interface{}{"10", "name"},
			},
			nil,
		},
		{
			HandlerData{
				handler: "not func",
				params:  []interface{}{"10", "name"},
			},
			ErrInvalidHandler,
		},
	}
	for i, c := range cases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		err := callHandler(w, r, c.input)
		if errors.Cause(err) != c.expect {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expect, err)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	r := NewRouter()
	http.Handle("/", r)
	ts := httptest.NewServer(r)

	cases := []struct {
		serverPath    string
		serverHandler baseHandler
		inputMethod   string
		inputPath     string
		expectBody    string
	}{
		{
			"/dummy/",
			dummyHandler,
			"GET",
			"/dummy",
			"hello, world",
		},
		{
			"/dummy/:id/dummy/:name",
			dummyHandlerWithParams,
			"GET",
			"/dummy/10/dummy/hoge",
			"id=10, name=hoge",
		},
		{
			"/dummy/",
			func(w http.ResponseWriter, req *http.Request) { fmt.Fprintf(w, "from post") },
			"POST",
			"/dummy/",
			"from post",
		},
	}
	for i, c := range cases {
		r.HandleFunc(c.inputMethod, c.serverPath, c.serverHandler)
		req, err := http.NewRequest(c.inputMethod, ts.URL+c.inputPath, nil)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
		body, _ := ioutil.ReadAll(res.Body)
		if c.expectBody != string(body) {
			t.Errorf("#%d: want body:%s, got body:%s", i, c.expectBody, string(body))
		}
		res.Body.Close()
	}
}
