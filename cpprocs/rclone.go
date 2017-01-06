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

type rclone struct {
	remote string
	tmp    *tmpdedup.Dir
}

func NewRclone(remote string, tmp *tmpdedup.Dir) LsProcUnprocer {
	return rclone{remote, tmp}
}

func (rc rclone) Proc() aprocs.Proc {
	return aprocs.NewPathCmdIn(rc.procCmd, rc.tmp)
}

func (rc rclone) procCmd(_ *ss.Chunk, path string) (*exec.Cmd, error) {
	cmd := exec.Command("rclone", "copy", path, rc.remote, "-q")
	return cmd, nil
}

func (rc rclone) Unproc() aprocs.Proc {
	return nil
}

func (rc rclone) Ls() (entries []LsEntry, err error) {
	cmd := rcloneLs(rc.remote)
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

var rcloneLs = func(remote string) *exec.Cmd {
	return exec.Command("rclone", "ls", remote, "-q")
}
