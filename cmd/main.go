package main

import (
	"errors"
	"fmt"
	"os"

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
		// case "join":
		// 	return cmdJoin()
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

func cmdSplit() (err error) {
	splitter := split.NewSplitter(os.Stdin)
	// index := procs.NewIndex(os.Stdout)
	chain := procs.Chain{
		procs.Checksum{}.Proc(),
		procs.NewDedup(),
		procs.Split(),
		procs.Checksum{}.Proc(),
		// (&procs.Compress{}).Proc(),
		// (&paritySplit{data: 2, parity: 1}).Process,
		(&procs.LocalStore{"out"}).Proc(),
		// index.Process,
	}
	ppool := procs.NewPool(8, chain)
	defer ppool.Finish()
	err = procs.Process(splitter, ppool)
	if err != nil {
		return
	}
	return chain.Finish()
}

// func cmdJoin() error {
// 	w := os.Stdout
// 	iter := newIndexScanner(os.Stdin)
// 	// TODO proc pool, respect order from index iterator
// 	process := procChain{
// 		inplace((&localStore{"out"}).UnprocessInplace).Process,
// 		// inplace((&compress{}).UnprocessInplace).Process,
// 		inplace(verify).Process,
// 		inplace((&out{w}).UnprocessInplace).Process,
// 	}.Process
// 	for iter.Next() {
// 		res := process(iter.Chunk())
// 		if e := res.err; e != nil {
// 			return e
// 		}
// 	}
// 	return iter.Err()
// }

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
