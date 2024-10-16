package server

var links lib.Map[Link]

func GetLink(id string) *Link {
	return links.Load(id)
}
