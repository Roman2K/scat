package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/restic/chunker"

	"ded3/meta"
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
	for {
		chunk, err := c.Next(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		meta := meta.Split{
			Size:   int64(len(chunk.Data)),
			Sha256: sha256.Sum256(chunk.Data),
		}
		err = binary.Write(out, binary.LittleEndian, meta)
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
