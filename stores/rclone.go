package stores

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"scat"
	"scat/checksum"
	"scat/procs"
	"scat/tmpdedup"
)

var (
	rcloneNotFoundRe   *regexp.Regexp
	errRcloneZeroBytes = errors.New("downloaded 0 bytes")
)

func init() {
	rcloneNotFoundRe = regexp.MustCompile(`\b(?i:not found)\b`)
}

type Rclone struct {
	Remote string
	Tmp    *tmpdedup.Dir
}

func (rc Rclone) Proc() procs.Proc {
	return procs.NewPathCmdIn(rc.procCmd, rc.Tmp)
}

func (rc Rclone) procCmd(_ *scat.Chunk, path string) (*exec.Cmd, error) {
	cmd := exec.Command("rclone", "copy", path, rc.Remote, "-q")
	return cmd, nil
}

func (rc Rclone) Unproc() procs.Proc {
	return procs.Filter{
		Proc: procs.CmdOutFunc(rc.loadCmd),
		Filter: func(res procs.Res) procs.Res {
			if err := rcloneDownloadErr(res); err != nil {
				res.Err = err
			}
			return res
		},
	}
}

//
// A download error manifests itself in different manners for different cloud
// providers. For instance, `rclone cat` for a missing file:
//
// * Drive: exit=1 stdout="" stderr="directory not found"
// * Dropbox: exit=0 stdout="" stderr=""
//
// So the most universal way of detecting a failed download is seeing 0 bytes on
// stdout.
//
func rcloneDownloadErr(res procs.Res) error {
	if res.Err != nil {
		return procs.MissingDataError{res.Err}
	}
	getSize := func() (sz int, err error) {
		c := res.Chunk
		if c == nil {
			err = errors.New("no chunk")
			return
		}
		data := c.Data()
		if sizer, ok := data.(scat.Sizer); ok {
			sz = sizer.Size()
			return
		}
		b, err := data.Bytes()
		if err != nil {
			return
		}
		sz = len(b)
		return
	}
	sz, err := getSize()
	if err == nil && sz <= 0 {
		err = errRcloneZeroBytes
	}
	if err != nil {
		return procs.MissingDataError{err}
	}
	return nil
}

func (rc Rclone) loadCmd(c *scat.Chunk) (*exec.Cmd, error) {
	remote := fmt.Sprintf("%s/%x", rc.Remote, c.Hash())
	cmd := rcloneCat(remote)
	return cmd, nil
}

func (rc Rclone) Ls() (entries []LsEntry, err error) {
	cmd := rcloneLs(rc.Remote)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	scan := bufio.NewScanner(bytes.NewReader(out))
	entries = make([]LsEntry, 0, bytes.Count(out, []byte{'\n'}))
	var (
		buf   = make([]byte, checksum.Size)
		entry LsEntry
	)
	for scan.Scan() {
		n, err := fmt.Sscanf(scan.Text(), "%d %x", &entry.Size, &buf)
		if err != nil || n != 2 {
			continue
		}
		err = entry.Hash.LoadSlice(buf)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return
}

// vars for tests
var (
	rcloneLs = func(remote string) *exec.Cmd {
		return exec.Command("rclone", "ls", remote, "-q")
	}
	rcloneCat = func(remote string) *exec.Cmd {
		return exec.Command("rclone", "cat", remote)
	}
)
