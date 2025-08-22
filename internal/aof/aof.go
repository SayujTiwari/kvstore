package aof

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/SayujTiwari/kvstore/internal/store"
)

type FsyncPolicy int

const (
	FsyncAlways FsyncPolicy = iota
	FsyncEverySec
)

type Logger struct {
	mu     sync.Mutex
	f      *os.File
	w      *bufio.Writer
	policy FsyncPolicy
	closed bool
	path   string
}

// New opens (or creates) the AOF file for appends.
func New(path string, policy FsyncPolicy) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	l := &Logger{f: f, w: bufio.NewWriter(f), policy: policy, path: path}

	// background fsync every second (if configured)
	if policy == FsyncEverySec {
		go func() {
			t := time.NewTicker(time.Second)
			defer t.Stop()
			for range t.C {
				l.mu.Lock()
				if l.closed {
					l.mu.Unlock()
					return
				}
				l.w.Flush()
				l.f.Sync()
				l.mu.Unlock()
			}
		}()
	}
	return l, nil
}

func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	if err := l.w.Flush(); err != nil {
		return err
	}
	if err := l.f.Close(); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.f = f
	l.w = bufio.NewWriter(f)
	return nil
}

func (l *Logger) AppendSet(k, v string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	if _, err := l.w.WriteString(fmt.Sprintf("SET %s %s\n", escape(k), escape(v))); err != nil {
		return err
	}
	if l.policy == FsyncAlways {
		l.w.Flush()
		return l.f.Sync()
	}
	return nil
}

func (l *Logger) AppendDel(k string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	if _, err := l.w.WriteString(fmt.Sprintf("DEL %s\n", escape(k))); err != nil {
		return err
	}
	if l.policy == FsyncAlways {
		l.w.Flush()
		return l.f.Sync()
	}
	return nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	l.w.Flush()
	return l.f.Close()
}

// Replay reads aofPath and replays commands into st.
func Replay(aofPath string, st *store.Store) error {
	f, err := os.Open(aofPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		switch strings.ToUpper(parts[0]) {
		case "SET":
			if len(parts) >= 3 {
				k := unescape(parts[1])
				v := unescape(strings.Join(parts[2:], " "))
				st.Set(k, v)
			}
		case "DEL":
			if len(parts) == 2 {
				st.Del(unescape(parts[1]))
			}
		}
	}
	return sc.Err()
}

// --- simple escaping so spaces/newlines in values survive ---
func escape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "\n", "\\n")
}
func unescape(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
