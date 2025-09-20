package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	proxy "github.com/imatakatsu/proxylib"
)

var (
	proxyFile  = flag.String("p", "proxies.txt", "proxy file")
	threads    = flag.Int("t", 100, "thread amount for checking proxies")
	outputFile = flag.String("o", "valid.txt", "output proxy file")
)

func main() {
	flag.Parse()
	bytes, err := os.ReadFile(*proxyFile)
	if err != nil {
		fmt.Println("failed to read file: ", err)
		return
	}
	proxies := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(bytes)), "\r\n", "\n"), "\n")
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	var (
		valid_proxies []string
		mu2           sync.Mutex
	)

	for range *threads {
		wg.Add(1)
		go func() {
			for {
				mu.Lock()
				if len(proxies) < 1 {
					mu.Unlock()
					break
				}
				p := proxies[0]
				proxies = proxies[1:]
				mu.Unlock()

				conn, err := proxy.DialTimeout(proxy.HTTP, p, "google.com:443", time.Duration(5)*time.Second)
				if err != nil {
					continue
				}
				conn.Close()

				mu2.Lock()
				valid_proxies = append(valid_proxies, p)
				mu2.Unlock()
				fmt.Printf("VALID PROXY %s\r\nTOTAL VALID: %d\r\n", p, len(valid_proxies))
			}
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println("valid proxy amount: ", len(valid_proxies))
	os.WriteFile(*outputFile, []byte(strings.Join(valid_proxies, "\r\n")), 0644)
}
