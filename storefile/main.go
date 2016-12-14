package main

import (
	"bufio"
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
	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, chunker.MaxSize)

	// TMP
	fnum := uint(0)
	// TMP

	for {
		chunkLen, err := binary.ReadVarint(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if chunkLen > int64(len(buf)) {
			return io.ErrShortBuffer
		}
		readBuf := buf[:chunkLen]
		err = fill(readBuf, r)
		if err != nil {
			return err
		}
		_, err = write(fnum, readBuf)
		if err != nil {
			return err
		}

		// TEMP
		fnum++
		// TEMP
	}
	return nil
}

func write(fnum uint, buf []byte) (n int, err error) {
	f, err := os.Create(fmt.Sprintf("%03d", fnum))
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
