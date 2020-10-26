package main

import (
	"log"
	"os"

	sandbox "github.com/jpillora/go-sandbox/lib"
)

//run it
func main() {
	s := sandbox.New()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(s.ListenAndServe(":" + port))
}
