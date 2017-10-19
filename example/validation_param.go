package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/takashabe/go-router"
)

// Allow3DigitNumber expect allow 3 digit number
// Require implements router.ValidateParam
type Allow3DigitNumber struct {
	raw string
}

// Validate is validate URL path parameters
func (v *Allow3DigitNumber) Validate(raw string) bool {
	v.raw = raw
	return regexp.MustCompile(`\A[0-9]{3}\z`).MatchString(raw)
}

func (v *Allow3DigitNumber) convert() (int, error) {
	return strconv.Atoi(v.raw)
}

func main() {
	r := router.NewRouter()
	r.Get("/:id", func(w http.ResponseWriter, req *http.Request, v *Allow3DigitNumber) {
		id, _ := v.convert()
		fmt.Printf("called get '/:id' with %d\n", id)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
