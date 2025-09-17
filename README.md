# proxylib
proxy library for golang. support only http proxies in current version

## example
```golang
package main

import (
	"fmt"
	"log"

	proxy "github.com/imatakatsu/proxylib"
)

func main() {
	conn, err := proxy.DialHTTP("tcp", "127.0.0.1:3128:admin:password", "ident.me:80")
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(
		"GET / HTTP/1.1\r\n" +
			"Host: ident.me\r\n" +
			"User-Agent: goProxy/0.1\r\n" +
			"Accept: */*\r\n\r\n",
	))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal()
	}
	fmt.Println(string(buf[:n]))
}
```
