package main

import (
	"errors"
	"fmt"
	"os"

	"secsplit/indexscan"
	"secsplit/procs"
	"secsplit/split"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	args := os.Args[1:]
	if len(args) != 1 {
		return errors.New("usage: split|join")
	}
	cmd := args[0]
	switch cmd {
	case "split":
		return cmdSplit()
	case "join":
		return cmdJoin()
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

func cmdSplit() (err error) {
	splitter := split.NewSplitter(os.Stdin)
	index := procs.NewIndex(os.Stdout)
	parity, err := procs.Parity(2, 1)
	if err != nil {
		return
	}
	chain := procs.NewChain([]procs.Proc{
		procs.Checksum{}.Proc(),
		index,
		procs.NewDedup(),
		parity.Proc(),
		procs.Checksum{}.Proc(),
		// (&procs.Compress{}).Proc(),
		(&procs.LocalStore{"out"}).Proc(),
	})
	ppool := procs.NewPool(8, chain)
	defer ppool.Finish()
	err = procs.Process(splitter, ppool)
	if err != nil {
		return
	}
	return chain.Finish()
}

func cmdJoin() (err error) {
	scan := indexscan.NewScanner(os.Stdin)
	out := procs.WriteTo(os.Stdout)
	parity, err := procs.Parity(2, 1)
	if err != nil {
		return
	}
	chain := procs.NewChain([]procs.Proc{
		(&procs.LocalStore{"out"}).Unproc(),
		(&procs.Group{parity.NShards}).Proc(),
		parity.Unproc(),
		out,
	})
	// TODO proc pool, respect order from index iterator
	for scan.Next() {
		res := chain.Process(scan.Chunk())
		if e := res.Err; e != nil {
			return e
		}
	}
	return scan.Err()
}

// type paritySplit struct {
// 	rs reedsolomon.Encoder
// }

// func newParitySplit(data, parity int) *paritySplit {
// 	return &paritySplit{rs: reedsolomon.New(data, parity)}
// }

// func (ps *paritySplit) Process(c *Chunk) outChunk {
// 	shards, err := rs.Split(c.Data)
// 	if err != nil {
// 		return outChunk{err: err}
// 	}
// 	out := make([]*Chunk, len(shards))
// 	for i, shard := range shards {
// 		out[i] = shard
// 	}
// 	return outChunk{out: out}
// }
