package cpprocs

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"scat"
	"scat/procs"
	"scat/checksum"
	"scat/tmpdedup"
)

type rclone struct {
	remote string
	tmp    *tmpdedup.Dir
}

func NewRclone(remote string, tmp *tmpdedup.Dir) LsProcUnprocer {
	return rclone{remote, tmp}
}

func (rc rclone) Proc() procs.Proc {
	return procs.NewPathCmdIn(rc.procCmd, rc.tmp)
}

func (rc rclone) procCmd(_ *scat.Chunk, path string) (*exec.Cmd, error) {
	cmd := exec.Command("rclone", "copy", path, rc.remote, "-q")
	return cmd, nil
}

func (rc rclone) Unproc() procs.Proc {
	return procs.CmdOutFunc(rc.unprocess)
}

func (rc rclone) unprocess(c *scat.Chunk) (*exec.Cmd, error) {
	remote := fmt.Sprintf("%s/%x", rc.remote, c.Hash())
	cmd := exec.Command("rclone", "cat", remote)
	return cmd, nil
}

func (rc rclone) Ls() (entries []LsEntry, err error) {
	cmd := rcloneLs(rc.remote)
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

var rcloneLs = func(remote string) *exec.Cmd {
	return exec.Command("rclone", "ls", remote, "-q")
}
