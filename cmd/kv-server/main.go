package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/SayujTiwari/kvstore/internal/aof"
	"github.com/SayujTiwari/kvstore/internal/proto"
	"github.com/SayujTiwari/kvstore/internal/snapshot"
	"github.com/SayujTiwari/kvstore/internal/store"
)

func getenvBool(k string, def bool) bool {
	v := strings.ToLower(os.Getenv(k))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "on" || v == "yes"
}
func getenvOneOf(k, def string, allowed ...string) string {
	v := strings.ToLower(os.Getenv(k))
	if v == "" {
		return def
	}
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	return def
}

func handleConn(c net.Conn, st *store.Store, logAOF *aof.Logger) {
	defer c.Close()
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	defer w.Flush()

	for {
		cmd, args, err := proto.ReadCommand(r)
		if err != nil {
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
			val := strings.Join(args[1:], " ")
			st.Set(key, val)
			if logAOF != nil {
				_ = logAOF.AppendSet(key, val)
			}
			proto.WriteString(w, "+OK\n")
		case "GET":
			if len(args) != 1 {
				proto.WriteString(w, "-ERR GET needs key\n")
				continue
			}
			if v, ok := st.Get(args[0]); ok {
				proto.WriteString(w, "$"+v+"\n")
			} else {
				proto.WriteString(w, "$(nil)\n")
			}
		case "DEL":
			if len(args) != 1 {
				proto.WriteString(w, "-ERR DEL needs key\n")
				continue
			}
			if st.Del(args[0]) {
				if logAOF != nil {
					_ = logAOF.AppendDel(args[0])
				}
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
	// --- config via env ---
	addr := getenvOneOf("KV_ADDR", ":6380") // host:port
	useAOF := getenvBool("KV_AOF", true)    // on/off
	useSnap := getenvBool("KV_SNAPSHOT", true)
	fsyncMode := getenvOneOf("KV_FSYNC", "everysec", "always", "everysec", "off")

	st := store.New()
	const aofPath = "data.aof"
	const snapPath = "data.snap"

	// Load snapshot first (if enabled), then AOF.
	if useSnap {
		if err := snapshot.Load(snapPath, st); err != nil {
			log.Fatal("snapshot load:", err)
		}
	}
	if useAOF {
		if err := aof.Replay(aofPath, st); err != nil {
			log.Fatal("replay:", err)
		}
	}

	// Open AOF logger if enabled
	var logAOF *aof.Logger
	if useAOF {
		pol := aof.FsyncEverySec
		if fsyncMode == "always" {
			pol = aof.FsyncAlways
		}
		var err error
		logAOF, err = aof.New(aofPath, pol)
		if err != nil {
			log.Fatal("aof:", err)
		}
		defer logAOF.Close()
	}

	// Background snapshots (if enabled)
	if useSnap {
		go func() {
			for {
				time.Sleep(30 * time.Second)
				if err := snapshot.Save(snapPath, st); err != nil {
					log.Println("snapshot save:", err)
					continue
				}
				// compact the AOF so restart only replays a tiny tail
				if useAOF && logAOF != nil {
					if err := logAOF.Rotate(); err != nil {
						log.Println("aof rotate:", err)
					}
				}
			}
		}()
	}

	// --- network server ---
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	log.Println("kv-server listening on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go handleConn(conn, st, logAOF)
	}
}
