package stores

import (
	"fmt"
	"os/exec"
	"scat"
	"scat/procs"
	"scat/testutil"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestRcloneMissingData(t *testing.T) {
	origCat := rcloneCat
	defer func() {
		rcloneCat = origCat
	}()

	exitCode, out, errOut := 0, "", ""
	rcloneCat = func(string) *exec.Cmd {
		cmd := exec.Command("bash", "-c", fmt.Sprintf(
			`cat; echo -n %q >&2; exit %d`, errOut, exitCode,
		))
		cmd.Stdin = strings.NewReader(out)
		return cmd
	}
	rc := Rclone{}

	exitCode, out, errOut = 0, "foo", "some benign err"
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(rc.Unproc().Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))
	b, err := chunks[0].Data().Bytes()
	assert.NoError(t, err)
	assert.Equal(t, out, string(b))

	getErr := func() error {
		c := scat.NewChunk(0, nil)
		chunks, err := testutil.ReadChunks(rc.Unproc().Process(c))
		assert.Error(t, err)
		assert.Equal(t, 1, len(chunks))
		b, bytesErr := chunks[0].Data().Bytes()
		assert.NoError(t, bytesErr)
		assert.Equal(t, 0, len(b))
		return err
	}

	exitCode, out, errOut = 1, "bar", "2017/02/01 10:17:01 directory not found"
	err = getErr()
	missErr, ok := err.(procs.MissingDataError)
	assert.True(t, ok)
	assert.IsType(t, &exec.ExitError{}, missErr.Err)

	exitCode, out, errOut = 1, "", ""
	err = getErr()
	missErr, ok = err.(procs.MissingDataError)
	assert.True(t, ok)
	assert.IsType(t, &exec.ExitError{}, missErr.Err)

	exitCode, out, errOut = 0, "", ""
	err = getErr()
	missErr, ok = err.(procs.MissingDataError)
	assert.True(t, ok)
	assert.Equal(t, errRcloneZeroBytes, missErr.Err)
}

func TestRcloneLs(t *testing.T) {
	origLs := rcloneLs
	defer func() {
		rcloneLs = origLs
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

	rcloneLs = func(string) *exec.Cmd {
		cmd := exec.Command("cat")
		cmd.Stdin = strings.NewReader(out)
		return cmd
	}

	rc := Rclone{}
	entries, err := rc.Ls()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entries))
	assert.Equal(t, int64(27), entries[0].Size)
	assert.Equal(t, e0h, fmt.Sprintf("%x", entries[0].Hash))
	assert.Equal(t, int64(27), entries[1].Size)
	assert.Equal(t, e1h, fmt.Sprintf("%x", entries[1].Hash))
}
