package snapshot

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/SayujTiwari/kvstore/internal/store"
)

const magic = "KVS1" // file header

// Save writes a compact binary snapshot: MAGIC, then [keyLen,varint][key][valLen,varint][val]...
func Save(path string, st *store.Store) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	// header
	if _, err := w.Write([]byte(magic)); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	// stream entries
	var buf [binary.MaxVarintLen64]byte
	st.ForEach(func(k, v string) {
		// key
		n := binary.PutUvarint(buf[:], uint64(len(k)))
		_, _ = w.Write(buf[:n])
		_, _ = w.Write([]byte(k))
		// value
		n = binary.PutUvarint(buf[:], uint64(len(v)))
		_, _ = w.Write(buf[:n])
		_, _ = w.Write([]byte(v))
	})

	if err := w.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// Load restores the store from a binary snapshot.
func Load(path string, st *store.Store) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	hdr := make([]byte, len(magic))
	if _, err := io.ReadFull(r, hdr); err != nil {
		return err
	}
	if string(hdr) != magic {
		return errors.New("invalid snapshot header")
	}

	for {
		klen, err := binary.ReadUvarint(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		kb := make([]byte, klen)
		if _, err := io.ReadFull(r, kb); err != nil {
			return err
		}

		vlen, err := binary.ReadUvarint(r)
		if err != nil {
			return err
		}
		vb := make([]byte, vlen)
		if _, err := io.ReadFull(r, vb); err != nil {
			return err
		}

		st.Set(string(kb), string(vb))
	}
	return nil
}
