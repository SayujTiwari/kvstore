package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

func runClient(addr string, n int, wg *sync.WaitGroup, ops *int64) {
	defer wg.Done()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	r := bufio.NewReader(conn)

	for i := 0; i < n; i++ {
		k := fmt.Sprintf("k%d", rand.Intn(100000))
		v := "x"
		fmt.Fprintf(conn, "SET %s %s\n", k, v)
		if _, err := r.ReadString('\n'); err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(conn, "GET %s\n", k)
		if _, err := r.ReadString('\n'); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	addr := "localhost:6380"
	if a := os.Getenv("KV_ADDR"); a != "" {
		addr = a
	}

	clients := 200
	itersPerClient := 1000

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(clients)
	for i := 0; i < clients; i++ {
		go runClient(addr, itersPerClient, &wg, nil)
	}
	wg.Wait()
	d := time.Since(start)
	total := clients * itersPerClient * 2 // SET+GET per iter
	opsPerSec := float64(total) / d.Seconds()
	fmt.Printf("total ops=%d  duration=%v  throughput=%.0f ops/sec\n", total, d, opsPerSec)
}
