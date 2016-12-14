package main

import (
	"ded3/meta"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/restic/chunker"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	r := os.Stdin
	buf := make([]byte, chunker.MaxSize)
	var meta meta.Split
	for {
		err := binary.Read(r, binary.LittleEndian, &meta)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if meta.Size > int64(len(buf)) {
			return io.ErrShortBuffer
		}
		readBuf := buf[:meta.Size]
		err = fill(readBuf, r)
		if err != nil {
			return err
		}
		_, err = write(meta.Sha256[:], readBuf)
		if err != nil {
			return err
		}
	}
	return nil
}

func write(csum, buf []byte) (n int, err error) {
	f, err := os.Create(fmt.Sprintf("%x", csum))
	if err != nil {
		return
	}
	defer f.Close()
	return f.Write(buf)
}

func fill(buf []byte, r io.Reader) error {
	offset := 0
	for {
		read, err := r.Read(buf[offset:])
		if err != nil {
			return err
		}
		offset += read
		bufLen := len(buf)
		if offset > bufLen {
			return io.ErrShortBuffer
		}
		if offset == bufLen {
			return nil
		}
	}
}
