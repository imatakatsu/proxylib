package main

import (
	"log"

	proxy "github.com/imatakatsu/proxylib"
)

func main() {
	conn, err := proxy.DialHTTP("tcp", "127.0.0.1:3128:admin:passwd", "google.com:443")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	// do something here
	return
}
