package stores

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"scat"
	"scat/procs"
	"strconv"
	"strings"
)

const ddBsArg = "bs=1048576" // most universal ("1M" on GNU, "1m" on macOS)

var (
	cmdGnuFind   = "find" // var for tests
	noSuchFileRe *regexp.Regexp
)

func init() {
	noSuchFileRe = regexp.MustCompile(`\b(?i:no such file)\b`)
}

type Dd struct {
	Dir        Dir
	Command    commandFunc
	StrCommand strCommandFunc
}

type commandFunc func(string, ...string) *exec.Cmd
type strCommandFunc func(env, string) *exec.Cmd

var _ Store = Dd{}

func (s Dd) Proc() procs.Proc {
	return procs.CmdInFunc(s.process)
}

func (s Dd) process(c *scat.Chunk) (*exec.Cmd, error) {
	path := s.Dir.FullPath(c.Hash())
	ofArg := "of=" + path
	return s.outCommand(path, ofArg), nil
}

func (s Dd) outCommand(path, ofArg string) *exec.Cmd {
	if len(s.Dir.Part) > 0 {
		// Pass paths around making sure to avoid string concatenation at all cost
		// so as to not go through shell escaping hell...
		env := env{
			"ddproc_dir=" + filepath.Dir(path),
			"ddproc_of=" + ofArg,
			"ddproc_bs=" + ddBsArg,
		}
		export := "true"
		if exports := env.exports(); len(exports) > 0 {
			export = "export " + strings.Join(exports, " ")
		}
		cmd := s.strCommand(env, export+
			` && mkdir -p "$ddproc_dir"`+
			` && dd "$ddproc_of" "$ddproc_bs"`)
		return cmd
	}
	return s.command("dd", ofArg, ddBsArg)
}

func (s Dd) command(name string, args ...string) *exec.Cmd {
	fn := exec.Command
	if s.Command != nil {
		fn = s.Command
	}
	return fn(name, args...)
}

func (s Dd) strCommand(env env, str string) *exec.Cmd {
	fn := defaultStrCommand
	if s.StrCommand != nil {
		fn = s.StrCommand
	}
	return fn(env, str)
}

func defaultStrCommand(env env, str string) (cmd *exec.Cmd) {
	cmd = exec.Command("sh", "-c", str)
	cmd.Env = env
	return
}

func (s Dd) Unproc() procs.Proc {
	return procs.Filter{
		Proc: procs.CmdOutFunc(s.loadCmd),
		Filter: func(res procs.Res) procs.Res {
			if exit, ok := res.Err.(*exec.ExitError); ok {
				if noSuchFileRe.Match(exit.Stderr) {
					res.Err = procs.ErrMissingData
				}
			}
			return res
		},
	}
}

func (s Dd) loadCmd(c *scat.Chunk) (*exec.Cmd, error) {
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
				// The sign of size is important for the control flow of this loop
				if sz < 0 {
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

type env []string

func (env env) exports() (names []string) {
	names = make([]string, 0, len(env))
	for _, s := range env {
		if idx := strings.IndexRune(s, '='); idx > -1 {
			names = append(names, s[:idx])
		}
	}
	return
}

func NewScp(host string, dir Dir) Store {
	return Dd{
		Dir: dir,
		Command: func(name string, args ...string) *exec.Cmd {
			args = append([]string{host, name}, args...)
			return exec.Command("ssh", args...)
		},
		StrCommand: func(env env, str string) *exec.Cmd {
			args := append(
				append([]string{host}, env...),
				str,
			)
			return exec.Command("ssh", args...)
		},
	}
}
