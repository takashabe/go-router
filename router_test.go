package router

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func dummyHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "hello, world")
}

func dummyHandlerWithParams(w http.ResponseWriter, req *http.Request, id int, name string) {
	fmt.Fprintf(w, "id=%d, name=%s", id, name)
}

type dummyValidationParam struct {
	raw string
}

func (v *dummyValidationParam) Validate(raw string) bool {
	v.raw = raw
	// allowed int value with 3 digit
	return regexp.MustCompile(`\A[0-9]{3}\z`).MatchString(raw)
}

func dummyHandlerWithValidationParams(w http.ResponseWriter, req *http.Request, v *dummyValidationParam) {
	fmt.Fprintf(w, "raw=%s", v.raw)
}

// for debug
func printValues(vs []reflect.Value) {
	for _, v := range vs {
		fmt.Printf("%#v\n", v)
	}
}

func TestParseParams(t *testing.T) {
	type invalidValidationParam struct{}

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
			nil,
			ErrNotFoundHandler,
		},
		{
			HandlerData{
				handler: dummyHandlerWithParams,
				params:  []interface{}{"hoge", "name"},
			},
			nil,
			ErrInvalidParam,
		},
		{
			HandlerData{
				handler: dummyHandlerWithValidationParams,
				params:  []interface{}{"100"},
			},
			[]reflect.Value{reflect.ValueOf(&dummyValidationParam{})},
			nil,
		},
		{
			HandlerData{
				handler: func(w http.ResponseWriter, req *http.Request, v *dummyValidationParam, id int) {},
				params:  []interface{}{"100", "10"},
			},
			[]reflect.Value{reflect.ValueOf(&dummyValidationParam{}), reflect.ValueOf(10)},
			nil,
		},
		{
			HandlerData{
				handler: func(w http.ResponseWriter, req *http.Request, v *invalidValidationParam) {},
				params:  []interface{}{"100"},
			},
			nil,
			ErrInvalidParam,
		},
	}
	for i, c := range cases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		router := NewRouter()
		result, err := router.parseParams(w, r, c.input)
		if errors.Cause(err) != c.expectError {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expectError, err)
		}
		if err != nil {
			continue
		}

		// complement a missing variable
		c.expectValues = append(c.expectValues, reflect.ValueOf(w))
		c.expectValues = append(c.expectValues, reflect.ValueOf(r))
		// compare to params
		if len(result) != len(c.expectValues) {
			t.Fatalf("#%d: want number of params %d, got %d", i, len(c.expectValues), len(result))
		}
		result = result[2:]
		for vi := 0; vi < len(result); vi++ {
			value := c.expectValues[vi].Interface()
			switch value.(type) {
			case ValidationParam:
				if _, ok := result[vi].Interface().(ValidationParam); !ok {
					t.Errorf("#%d-%d: want matched ValidationParam", i, vi)
				}
			default:
				if result[vi].Interface() != value {
					t.Errorf("#%d-%d: want result:%#v , got result:%#v ", i, vi, c.expectValues[vi], result[vi])
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
		router := NewRouter()
		err := router.callHandler(w, r, c.input)
		if errors.Cause(err) != c.expect {
			t.Errorf("#%d: want error:%#v , got error:%#v ", i, c.expect, err)
		}
	}
}

func TestServeHTTP(t *testing.T) {
	cases := []struct {
		serverPath    string
		serverHandler baseHandler
		inputMethod   string
		inputPath     string
		expectBody    string
		expectStatus  int
	}{
		{
			"/",
			dummyHandler,
			"GET",
			"/",
			"hello, world",
			200,
		},
		{
			"/:id/:name",
			dummyHandlerWithParams,
			"GET",
			"/10/hoge",
			"id=10, name=hoge",
			200,
		},
		{
			"/dummy/:id/dummy/:name",
			dummyHandlerWithParams,
			"GET",
			"/dummy/10/dummy/hoge",
			"id=10, name=hoge",
			200,
		},
		{
			"/",
			func(w http.ResponseWriter, req *http.Request) { fmt.Fprintf(w, "from post") },
			"POST",
			"/",
			"from post",
			200,
		},
		{
			"/dummy/",
			dummyHandler,
			"GET",
			"/dummy/10",
			"404 page not found\n", // http.NotFoundHandler used fmt.Fprintln()
			404,
		},
		{
			"/dummy/:id",
			func(w http.ResponseWriter, req *http.Request, id int) {},
			"GET",
			"/dummy/notint",
			"404 page not found\n", // http.NotFoundHandler used fmt.Fprintln()
			404,
		},
		{
			"/:id",
			dummyHandlerWithValidationParams,
			"GET",
			"/100",
			"raw=100",
			200,
		},
		{
			"/:id",
			dummyHandlerWithValidationParams,
			"GET",
			"/1000",
			"404 page not found\n", // http.NotFoundHandler used fmt.Fprintln()
			404,
		},
		{
			"/:id",
			func(w http.ResponseWriter, req *http.Request, v *dummyValidationParam) {
				fmt.Fprintf(w, "from ValidationParam with non-pointer")
			},
			"GET",
			"/100",
			"from ValidationParam with non-pointer",
			200,
		},
	}
	for i, c := range cases {
		r := NewRouter()
		ts := httptest.NewServer(r)
		defer ts.Close()

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
		defer res.Body.Close()

		if res.StatusCode != c.expectStatus {
			t.Errorf("#%d: want status code:%d, got status code:%d", i, c.expectStatus, res.StatusCode)
		}
		if body, _ := ioutil.ReadAll(res.Body); c.expectBody != string(body) {
			t.Errorf("#%d: want body:%s, got body:%s", i, c.expectBody, string(body))
		}
	}
}

func TestServeHTTPWithMultiplePath(t *testing.T) {
	cases := []struct {
		routes       []*Route
		inputMethod  string
		inputPath    string
		expectStatus int
	}{
		{
			[]*Route{
				&Route{method: "GET", path: "/dummy/:id", handler: dummyHandler},
				&Route{method: "GET", path: "/dummy/", handler: dummyHandler},
				&Route{method: "GET", path: "/dummy/", handler: dummyHandler},
				&Route{method: "GET", path: "/", handler: dummyHandler},
			},
			"GET",
			"/dummy/",
			200,
		},
		{
			[]*Route{
				&Route{method: "GET", path: "/", handler: dummyHandler},
				&Route{method: "GET", path: "/:id", handler: func(w http.ResponseWriter, req *http.Request, id int) {}},
			},
			"GET",
			"/",
			200,
		},
		{
			[]*Route{
				&Route{method: "GET", path: "/user/", handler: dummyHandler},
				&Route{method: "GET", path: "/user/:id", handler: func(w http.ResponseWriter, req *http.Request, id int) {}},
			},
			"GET",
			"/user/10",
			200,
		},
	}
	for i, c := range cases {
		r := NewRouter()
		ts := httptest.NewServer(r)
		defer ts.Close()

		for _, route := range c.routes {
			r.HandleFunc(route.method, route.path, route.handler)
		}
		req, err := http.NewRequest(c.inputMethod, ts.URL+c.inputPath, nil)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != c.expectStatus {
			t.Errorf("#%d: want status code:%d, got status code:%d", i, c.expectStatus, res.StatusCode)
		}
	}
}

func TestHandleFuncWithMethod(t *testing.T) {
	echoMethod := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, req.Method)
	}

	r := NewRouter()
	r.Get("/", echoMethod)
	r.Head("/", echoMethod)
	r.Post("/", echoMethod)
	r.Put("/", echoMethod)
	r.Patch("/", echoMethod)
	r.Delete("/", echoMethod)
	r.Options("/", echoMethod)
	ts := httptest.NewServer(r)
	defer ts.Close()

	cases := []struct {
		method string
	}{
		{"GET"},
		{"HEAD"},
		{"POST"},
		{"PUT"},
		{"PATCH"},
		{"DELETE"},
		{"OPTIONS"},
	}
	for i, c := range cases {
		req, err := http.NewRequest(c.method, ts.URL+"/", nil)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
		defer res.Body.Close()
		if body, _ := ioutil.ReadAll(res.Body); c.method != string(body) && string(body) != "" {
			t.Errorf("#%d: want body:%s, got body:%s", i, c.method, string(body))
		}
	}
}

func TestServeHTTPWithPost(t *testing.T) {
	r := NewRouter()
	r.Post("/login", func(w http.ResponseWriter, req *http.Request) {})
	ts := httptest.NewServer(r)
	defer ts.Close()

	_, err := http.PostForm(ts.URL+"/login", url.Values{
		"id": []string{"foo", "bar"}},
	)
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestServeDir(t *testing.T) {
	cases := []struct {
		definePath string
		defineRoot http.FileSystem
		input      string
		expectCode int
		expectBody string // see "/testdata/*"
	}{
		{
			"/testdata/*filepath",
			http.Dir("testdata"),
			"/testdata/foo",
			200,
			"hello from testdata/foo\n",
		},
		{
			"/testdata/dir/*filepath",
			http.Dir("testdata"),
			"/testdata/dir/dir/bar",
			200,
			"hello from testdata/dir/bar\n",
		},
		{
			"/testdata/*filepath",
			http.Dir("testdata"),
			"/testdata/../router_test.go",
			404,
			"",
		},
		{
			"/foo/*filepath",
			http.Dir("testdata"),
			"/foo/foo",
			200,
			"hello from testdata/foo\n",
		},
	}
	for i, c := range cases {
		r := NewRouter()
		r.ServeDir(c.definePath, c.defineRoot)
		ts := httptest.NewServer(r)
		defer ts.Close()

		res, err := http.Get(ts.URL + c.input)
		if err != nil {
			t.Errorf("#%d: want no error, got %v", i, err)
		}
		defer res.Body.Close()
		if res.StatusCode != c.expectCode {
			t.Errorf("#%d: want %d, got %d", i, c.expectCode, res.StatusCode)
		}
		if c.expectBody != "" {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("#%d: want no error, got %v", i, err)
			}
			if string(body) != c.expectBody {
				t.Errorf("#%d: want %s, got %s", i, c.expectBody, string(body))
			}
		}
	}
}

func TestServeFile(t *testing.T) {
	cases := []struct {
		definePath string
		defineRoot string
		input      string
		expectCode int
		expectBody string // see "/testdata/*"
	}{
		{
			"/",
			"./testdata/foo",
			"/",
			200,
			"hello from testdata/foo\n",
		},
		{
			"/",
			"./testdata/../testdata/foo",
			"/",
			400,
			"",
		},
	}
	for i, c := range cases {
		r := NewRouter()
		r.ServeFile(c.definePath, c.defineRoot)
		ts := httptest.NewServer(r)
		defer ts.Close()

		res, err := http.Get(ts.URL + c.input)
		if err != nil {
			t.Errorf("#%d: want no error, got %v", i, err)
		}
		defer res.Body.Close()
		if res.StatusCode != c.expectCode {
			t.Errorf("#%d: want %d, got %d", i, c.expectCode, res.StatusCode)
		}
		if c.expectBody != "" {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("#%d: want no error, got %v", i, err)
			}
			if string(body) != c.expectBody {
				t.Errorf("#%d: want %s, got %s", i, c.expectBody, string(body))
			}
		}
	}
}

func TestPrintRoutes(t *testing.T) {
	router := NewRouter()
	router.Get("/", dummyHandler)
	router.Get("/bar/:id/:id", func(w http.ResponseWriter, r *http.Request, a, b int) {})
	router.Post("/", dummyHandler)

	var buf bytes.Buffer
	router.PrintRoutes(&buf)
	want := `[GET] "/" -> github.com/takashabe/go-router.dummyHandler
[GET] "/bar/:id/:id" -> github.com/takashabe/go-router.TestPrintRoutes.func1
[POST] "/" -> github.com/takashabe/go-router.dummyHandler`
	if strings.TrimSpace(buf.String()) != want {
		t.Errorf("want:\n%s\ngot:\n%s", want, buf.String())
	}
}
