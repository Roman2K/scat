package cpprocs

import (
	"os/exec"
	"secsplit/checksum"
)

type PathCmdSpawner interface {
	NewPathCmd(checksum.Hash, string) (*exec.Cmd, error)
	Finish() error
}
