package main

import (
	"bufio"
	"log"
	"net"
	"strings"

	"github.com/SayujTiwari/kvstore/internal/proto"
	"github.com/SayujTiwari/kvstore/internal/store"
)

func handleConn(c net.Conn, st *store.Store) {
	defer c.Close()
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	defer w.Flush()

	for {
		cmd, args, err := proto.ReadCommand(r)
		if err != nil {
			// client closed or error; end this connection
			return
		}
		switch cmd {
		case "PING":
			proto.WriteString(w, "+PONG\n")
		case "SET":
			if len(args) < 2 {
				proto.WriteString(w, "-ERR SET needs key and value\n")
				continue
			}
			key := args[0]
			val := strings.Join(args[1:], " ") // allow spaces in value
			st.Set(key, val)
			proto.WriteString(w, "+OK\n")
		case "GET":
			if len(args) != 1 {
				proto.WriteString(w, "-ERR GET needs key\n")
				continue
			}
			if v, ok := st.Get(args[0]); ok {
				proto.WriteString(w, "$"+v+"\n") // simple bulk response
			} else {
				proto.WriteString(w, "$(nil)\n")
			}
		case "DEL":
			if len(args) != 1 {
				proto.WriteString(w, "-ERR DEL needs key\n")
				continue
			}
			if st.Del(args[0]) {
				proto.WriteString(w, ":1\n")
			} else {
				proto.WriteString(w, ":0\n")
			}
		default:
			proto.WriteString(w, "-ERR unknown command\n")
		}
		w.Flush()
	}
}

func main() {
	ln, err := net.Listen("tcp", ":6380") // Redis-ish, but 6380
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Println("kv-server listening on :6380")

	st := store.New()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go handleConn(conn, st) // goroutine per client
	}
}
