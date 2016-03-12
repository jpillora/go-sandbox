package main

import (
	"os"

	"github.com/jpillora/go-sandbox/lib"
)

//run it
func main() {
	s := sandbox.New()
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	s.ListenAndServe(":" + port)
}
