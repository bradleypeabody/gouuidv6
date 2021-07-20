package main

import (
	"log"

	"github.com/kai5263499/gouuidv6"
)

func main() {
	id1 := gouuidv6.New()
	id1Str := id1.String()
	log.Printf(
		"id1 id=%s\ntime=%s\nbytes=%b\nnode=%d\ngouuid.node=%d\ngouuid.node=%b",
		id1Str,
		id1.Time(),
		id1.Bytes(),
		id1.Node(),
		gouuidv6.GetNode(),
		gouuidv6.GetNode(),
	)

	id2, err := gouuidv6.Parse(id1Str)
	if err != nil {
		log.Fatalf("unable to parse id1 from string err=%s", err)
	}

	log.Printf("id2 id=%s time=%s node=%d", id2.String(), id2.Time(), id2.Node())

	id3, err := gouuidv6.ParseBytes(id1.Bytes())
	if err != nil {
		log.Fatalf("unable to parse id1 from string err=%s", err)
	}
	log.Printf("id3 id=%s time=%s node=%d", id3.String(), id3.Time(), id3.Node())
}
