package tmpdedup_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat/tmpdedup"
)

func TestMan(t *testing.T) {
	const data = "ok"

	tmpf, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	tmp := tmpf.Name()
	defer os.Remove(tmp)
	tmpExists := func() bool {
		_, err := os.Stat(tmp)
		return err == nil
	}
	readTmp := func() string {
		data, err := ioutil.ReadFile(tmp)
		assert.NoError(t, err)
		return string(data)
	}

	man := tmpdedup.NewMan()
	created := 0
	create := func() (err error) {
		created++
		f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		_, err = f.Write([]byte(data))
		return f.Close()
	}

	_, err = man.Get(tmp, create)
	assert.Error(t, err)
	assert.True(t, strings.HasSuffix(err.Error(), "file exists"))
	assert.Equal(t, 1, created)
	assert.Equal(t, 0, man.Len())

	os.Remove(tmp)

	wg1, err := man.Get(tmp, create)
	assert.NoError(t, err)
	assert.True(t, tmpExists())
	assert.Equal(t, data, readTmp())
	assert.Equal(t, 2, created)
	assert.Equal(t, 1, man.Len())

	wg2, err := man.Get(tmp, create)
	assert.NoError(t, err)
	assert.True(t, tmpExists())
	assert.Equal(t, data, readTmp())
	assert.Equal(t, 2, created)
	assert.Equal(t, 1, man.Len())

	for i := 0; i < 100; i++ {
		wg, err := man.Get(tmp, create)
		assert.NoError(t, err)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()
	}
	go wg1.Done()
	go wg2.Done()
	man.Wait()
	assert.False(t, tmpExists())
	assert.Equal(t, 2, created)
	assert.Equal(t, 0, man.Len())

	wg, err := man.Get(tmp, create)
	assert.True(t, tmpExists())
	assert.Equal(t, data, readTmp())
	assert.Equal(t, 3, created)
	assert.Equal(t, 1, man.Len())

	wg.Done()
	man.Wait()
	assert.Equal(t, 0, man.Len())
	assert.False(t, tmpExists())
}
