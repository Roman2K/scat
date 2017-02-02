package index_test

import (
	"bytes"
	"testing"

	"gitlab.com/Roman2K/scat/index"
	"gitlab.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	index.Write(buf, testutil.Hash1.Hash, 123)
	assert.Equal(t, testutil.Hash1.Hex+" 123\n", string(buf.Bytes()))
}
