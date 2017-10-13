# go-router

Provide routing for HTTP request. It implemented `http.handler` interface, thereby easily introducing.
Support URL path parameters and can be mapped to argument of handler method.

## Installation

```
go get -u github.com/takashabe/go-router
```

## Usage

Basic usage:

```go
package main

import (
  "fmt"
  "log"
  "net/http"

  "github.com/takashabe/go-router"
)

func main() {
  r := router.NewRouter()
  r.Get("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Println("called get '/'")
  })
  // Path "/" can be registered each HTTP methods
  r.Post("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Println("called post '/'")
  })
  // ex. receive query "/10":
  // Mapped to id=10, and "10" already checked valid int type
  r.Get("/:id", func(w http.ResponseWriter, req *http.Request, id int) {
    fmt.Printf("called get '/:id' with %d\n", id)
  })

  log.Fatal(http.ListenAndServe(":8080", r))
}
```

For Static files:

```go
func Routes() *router.Router {
  r := router.NewRouter()
  r.Get("/", getIndex)

  // Query "css/*" mapped to "./static/css/*"
  r.ServeDir("/css/*filepath", http.Dir("static/css"))

  return r
}
```

For SPA application:

```go
func Routes() *router.Router {
  r := router.NewRouter()

  // Routing of the backend API
  r.Get("/api/login", ...)
  r.Post("/api/login", ...)

  // Routing of the frontend
  r.ServeFile("/", "./public/index.html")
  r.ServeFile("/bundle.js", "./public/bundle.js")
  r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    http.ServeFile(w, req, "./public/index.html")
  })

  return r
}
```
