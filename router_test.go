package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

func dummyHandler(w http.ResponseWriter, req *http.Request) {}

func dummyHandlerWithParams(w http.ResponseWriter, req *http.Request, id int, name string) {}

func printValues(vs []reflect.Value) {
	for _, v := range vs {
		fmt.Printf("%#v\n", v)
	}
}

func _TestServeHTTP(t *testing.T) {
	r := NewRouter()
	r.Get("/dummy/", dummyHandler)
	r.Get("/dummy/:id/dummy/:name", dummyHandlerWithParams)
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
	body, _ := ioutil.ReadAll(res.Body)
	pp.Println(string(body))
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
			[]reflect.Value{reflect.ValueOf("10"), reflect.ValueOf("name")},
			nil,
		},
		{
			HandlerData{
				handler: dummyHandlerWithParams,
				params:  []interface{}{"10"},
			},
			[]reflect.Value{reflect.ValueOf("10"), reflect.ValueOf("name")},
			ErrNotFoundHandler,
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
