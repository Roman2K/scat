package cpprocs

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestRcloneLister(t *testing.T) {
	lsOrig := rcloneLs
	defer func() {
		rcloneLs = lsOrig
	}()

	out := "" +
		`       27 b35c29e433130160c9e0fddebdc6a705b86cbe657f516efc149520884bdfd899
       27 Copy of c0528a15b8504c39ac55735705155466a5991b5f260784c03c187dca8d50b969
       27 a646cf8e18d00b01c654f7e2c85834491ba0a4fec44e5a630d53a3ef15fc2ea4
       27 a/c0528a15b8504c39ac55735705155466a5991b5f260784c03c187dca8d50b969
       27 a/b35c29e433130160c9e0fddebdc6a705b86cbe657f516efc149520884bdfd899
       27 a/a646cf8e18d00b01c654f7e2c85834491ba0a4fec44e5a630d53a3ef15fc2ea4
       27 b/c0528a15b8504c39ac55735705155466a5991b5f260784c03c187dca8d50b969
       27 b/a646cf8e18d00b01c654f7e2c85834491ba0a4fec44e5a630d53a3ef15fc2ea4
       27 b/b35c29e433130160c9e0fddebdc6a705b86cbe657f516efc149520884bdfd899
`
	e0h := "b35c29e433130160c9e0fddebdc6a705b86cbe657f516efc149520884bdfd899"
	e1h := "a646cf8e18d00b01c654f7e2c85834491ba0a4fec44e5a630d53a3ef15fc2ea4"

	rcl := NewRcloneLister("xxx:yyy")
	rcloneLs = func(string) *exec.Cmd {
		cmd := exec.Command("cat")
		cmd.Stdin = strings.NewReader(out)
		return cmd
	}

	entries, err := rcl.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entries))
	assert.Equal(t, int64(27), entries[0].Size)
	assert.Equal(t, e0h, fmt.Sprintf("%x", entries[0].Hash))
	assert.Equal(t, int64(27), entries[1].Size)
	assert.Equal(t, e1h, fmt.Sprintf("%x", entries[1].Hash))
}
