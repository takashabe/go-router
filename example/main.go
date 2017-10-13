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
	// Mapped to id=10, and "10" already checked valid int type.
	r.Get("/:id", func(w http.ResponseWriter, req *http.Request, id int) {
		fmt.Printf("called get '/:id' with %d\n", id)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
