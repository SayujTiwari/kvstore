package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/SayujTiwari/kvstore/internal/aof"
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

func main() {
	useAOF := getenvBool("KV_AOF", true)
	useSnap := getenvBool("KV_SNAPSHOT", true)

	st := store.New()
	const aofPath = "data.aof"
	const snapPath = "data.snap"

	start := time.Now()
	if useSnap {
		if err := snapshot.Load(snapPath, st); err != nil {
			log.Fatal("snapshot:", err)
		}
	}
	if useAOF {
		if err := aof.Replay(aofPath, st); err != nil {
			log.Fatal("replay:", err)
		}
	}
	elapsed := time.Since(start)
	log.Printf("startup_ms=%d (AOF=%v, SNAP=%v)\n", elapsed.Milliseconds(), useAOF, useSnap)
}
