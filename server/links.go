package server

import "github.com/god-jason/bucket/lib"

var links lib.Map[Link]

func GetLink(id string) *Link {
	return links.Load(id)
}
