package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	addr := "localhost:6380"
	if a := os.Getenv("KV_ADDR"); a != "" {
		addr = a // allow KV_ADDR=host:port overrides
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: kv-cli COMMAND [args...]")
		fmt.Println("Examples:")
		fmt.Println("  kv-cli PING")
		fmt.Println("  kv-cli SET name sayuj")
		fmt.Println("  kv-cli GET name")
		fmt.Println("  kv-cli DEL name")
		os.Exit(1)
	}

	// Join args into a single command line
	cmd := strings.Join(os.Args[1:], " ")

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("dial %s: %v", addr, err)
	}
	defer conn.Close()

	// send line-terminated command
	if _, err := fmt.Fprintf(conn, "%s\n", cmd); err != nil {
		log.Fatalf("write: %v", err)
	}

	// read one-line response
	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	fmt.Print(resp)
}
