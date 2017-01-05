package cpprocs

import (
	"bufio"
	"fmt"
	"os/exec"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/tmpdedup"
)

func NewRcloneLsProc(remote string, tmp *tmpdedup.Dir) LsProc {
	lser := NewRcloneLister(remote)
	proc := NewRcloneProc(remote, tmp)
	return NewLsProc(lser, proc)
}

func NewRcloneProc(remote string, tmp *tmpdedup.Dir) aprocs.Proc {
	newCmd := func(_ *ss.Chunk, path string) (*exec.Cmd, error) {
		cmd := exec.Command("rclone", "copy", path, remote, "-q")
		return cmd, nil
	}
	return aprocs.NewPathCmdIn(newCmd, tmp)
}

type rcloneLister struct {
	remote string
}

func NewRcloneLister(remote string) Lister {
	return rcloneLister{remote: remote}
}

var rcloneLs = func(remote string) *exec.Cmd {
	return exec.Command("rclone", "ls", remote, "-q")
}

func (rcl rcloneLister) Ls() (entries []LsEntry, err error) {
	cmd := rcloneLs(rcl.remote)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	scan := bufio.NewScanner(out)
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
	err = cmd.Wait()
	return
}
