package stores

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"scat"
	"scat/procs"
	"strconv"
)

const ddBsArg = "bs=1m"

type Dd struct {
	Dir     Dir
	Command commandFunc
}

type commandFunc func(string, ...string) *exec.Cmd

var _ Store = Dd{}

func (s Dd) Proc() procs.Proc {
	return procs.CmdInFunc(s.process)
}

func (s Dd) process(c *scat.Chunk) (*exec.Cmd, error) {
	path := s.Dir.FullPath(c.Hash())
	ofArg := "of=" + path
	name, args, env := "dd", []string{ofArg, ddBsArg}, []string{}
	if len(s.Dir.Part) > 0 {
		// Pass paths around making sure to avoid string concatenation at all cost
		// so as to not go through shell escaping hell...
		name = "sh"
		args = []string{
			"-c",
			`mkdir -p "$ddproc_dir" && dd "$ddproc_of" "$ddproc_bs"`,
		}
		env = []string{
			"ddproc_dir=" + filepath.Dir(path),
			"ddproc_of=" + ofArg,
			"ddproc_bs=" + ddBsArg,
		}
	}
	cmd := s.command(name, args...)
	cmd.Env = env
	return cmd, nil
}

func (s Dd) command(name string, args ...string) *exec.Cmd {
	fn := exec.Command
	if s.Command != nil {
		fn = s.Command
	}
	return fn(name, args...)
}

func (s Dd) Unproc() procs.Proc {
	return procs.CmdOutFunc(s.unprocess)
}

func (s Dd) unprocess(c *scat.Chunk) (*exec.Cmd, error) {
	path := s.Dir.FullPath(c.Hash())
	return s.command("dd", "if="+path, ddBsArg), nil
}

func (s Dd) Ls() ([]LsEntry, error) {
	return s.Dir.Ls(findDirLister(s.command))
}

type findDirLister commandFunc

func (fn findDirLister) Ls(dir string, depth int) <-chan DirLsRes {
	depthStr := fmt.Sprintf("%d", depth)
	cmd := fn(cmdGnuFind, dir,
		"-mindepth", depthStr, "-maxdepth", depthStr,
		"-type", "f",
		"-printf", `%s\0%p\0`,
	)
	errOut := &bytes.Buffer{}
	cmd.Stderr = errOut

	start := func() (out io.Reader, err error) {
		out, err = cmd.StdoutPipe()
		if err != nil {
			return
		}
		err = cmd.Start()
		return
	}

	sendResults := func(ch chan<- DirLsRes) error {
		out, err := start()
		if err != nil {
			return err
		}
		scan := bufio.NewScanner(out)
		scan.Split(byteSep('\000').scan)
		var size int64 = -1
		for scan.Scan() {
			if size < 0 {
				sz, err := strconv.ParseInt(scan.Text(), 10, 64)
				if err != nil {
					return err
				}
				if sz < 0 {
					// The sign of size is important for the control flow of this loop
					return errors.New("size can't be negative")
				}
				size = sz
			} else {
				name := filepath.Base(scan.Text())
				ch <- DirLsRes{Name: name, Size: size}
				size = -1
			}
		}
		return scan.Err()
	}

	ch := make(chan DirLsRes)
	go func() {
		defer close(ch)
		err := sendResults(ch)
		cmdErr := cmd.Wait()
		if cmdErr != nil {
			if exit, ok := cmdErr.(*exec.ExitError); ok {
				exit.Stderr = errOut.Bytes()
			}
			err = cmdErr
		}
		if err != nil {
			ch <- DirLsRes{Err: err}
		}
	}()
	return ch
}

type byteSep byte

func (s byteSep) scan(data []byte, _ bool) (n int, tok []byte, _ error) {
	n = bytes.IndexByte(data, byte(s))
	if n < 0 {
		n = 0 // load more
		return
	}
	tok = data[:n]
	n++
	return
}

func NewScp(host string, dir Dir) Store {
	return Dd{
		Dir: dir,
		Command: func(name string, args ...string) *exec.Cmd {
			args = append([]string{host, name}, args...)
			return exec.Command("ssh", args...)
		},
	}
}
