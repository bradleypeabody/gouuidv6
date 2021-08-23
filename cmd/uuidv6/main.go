package main

import (
	"log"

	"github.com/kai5263499/gouuidv6"
)

func main() {
	id1 := gouuidv6.New()
	id1Str := id1.String()
	log.Printf(
		"id1 id=%s gouuid.node=%d node=%x",
		id1Str,
		gouuidv6.GetNode(),
		id1.Node(),
	)

	gouuidv6.AlwaysRandomizeNode()
	id2 := gouuidv6.New()

	log.Printf(
		"id2 id=%s node=%x",
		id2.String(),
		id2.Node())

	id3 := gouuidv6.New()
	log.Printf(
		"id3 id=%s time=%s node=%x",
		id3.String(),
		id3.Time(),
		id3.Node())
}
