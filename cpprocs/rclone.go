package cpprocs

import (
	"os/exec"

	ss "secsplit"
	"secsplit/aprocs"
)

type rclone struct {
	remote string
}

func NewRcloneProc(remote, tmpDir string) (lsp LsProc, err error) {
	rc := rclone{remote: remote}
	proc, err := aprocs.NewPathCmdIn(tmpDir, rc.procCmd)
	lsp = NewLsProc(rc, proc)
	return
}

func (rc rclone) procCmd(_ *ss.Chunk, path string) (*exec.Cmd, error) {
	cmd := exec.Command("rclone", "copy", path, rc.remote)
	return cmd, nil
}

func (rc rclone) Ls() ([]LsEntry, error) {
	return nil, nil
}
