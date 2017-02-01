package stores

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"scat"
	"scat/checksum"
	"scat/procs"
	"scat/tmpdedup"
)

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
	return procs.CmdOutFunc(rc.unprocess)
}

func (rc Rclone) unprocess(c *scat.Chunk) (*exec.Cmd, error) {
	remote := fmt.Sprintf("%s/%x", rc.Remote, c.Hash())
	cmd := exec.Command("rclone", "cat", remote)
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

var rcloneLs = func(remote string) *exec.Cmd {
	return exec.Command("rclone", "ls", remote, "-q")
}
