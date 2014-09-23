package main

import "github.com/jpillora/go-sandbox"

func main() {
	s := sandbox.New()
	s.ListenAndServe(3000)
}
