package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
)

func main() {
	addr := "localhost:6380"
	if v := os.Getenv("KV_ADDR"); v != "" {
		addr = v
	}

	// Args: N [concurrency] [pipeline]
	N := 1_000_000
	if len(os.Args) > 1 {
		if x, err := strconv.Atoi(os.Args[1]); err == nil {
			N = x
		}
	}
	concurrency := 50
	if len(os.Args) > 2 {
		if x, err := strconv.Atoi(os.Args[2]); err == nil {
			concurrency = x
		}
	}
	pipeline := 8
	if len(os.Args) > 3 {
		if x, err := strconv.Atoi(os.Args[3]); err == nil {
			pipeline = x
		}
	}

	type job struct{ i int }
	jobs := make(chan job, 1024)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	for w := 0; w < concurrency; w++ {
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()
			r := bufio.NewReader(conn)

			buf := make([]byte, 0, 4096)
			send := func(s string) {
				buf = append(buf, s...)
				if len(buf) > 1<<15 { // flush if buffer is large
					if _, err := conn.Write(buf); err != nil {
						log.Fatal(err)
					}
					buf = buf[:0]
				}
			}

			count := 0
			for j := range jobs {
				k := j.i
				send(fmt.Sprintf("SET k%d v%d\n", k, rand.Int()))
				count++
				if count%pipeline == 0 {
					if len(buf) > 0 {
						if _, err := conn.Write(buf); err != nil {
							log.Fatal(err)
						}
						buf = buf[:0]
					}
					// read 'pipeline' replies
					for i := 0; i < pipeline; i++ {
						if _, err := r.ReadString('\n'); err != nil {
							log.Fatal(err)
						}
					}
				}
			}
			// flush remaining
			if len(buf) > 0 {
				if _, err := conn.Write(buf); err != nil {
					log.Fatal(err)
				}
				buf = buf[:0]
			}
		}()
	}

	for i := 0; i < N; i++ {
		jobs <- job{i}
	}
	close(jobs)
	wg.Wait()
	fmt.Printf("loaded %d keys with %d workers, pipeline=%d\n", N, concurrency, pipeline)
}
