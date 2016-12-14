package main

import (
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

func start() (err error) {
	out := os.Stdout
	c := chunker.New(os.Stdin, chunker.Pol(0x3DA3358B4DC173))
	buf := make([]byte, chunker.MaxSize)
	lenBuf := make([]byte, binary.MaxVarintLen64)
	for {
		chunk, err := c.Next(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		written := binary.PutVarint(lenBuf, int64(len(chunk.Data)))
		_, err = out.Write(lenBuf[:written])
		if err != nil {
			return err
		}
		_, err = out.Write(chunk.Data)
		if err != nil {
			return err
		}
	}
	return nil
}
