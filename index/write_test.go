package index_test

import (
	"bytes"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat/index"
	"github.com/Roman2K/scat/testutil"
)

func TestWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	index.Write(buf, testutil.Hash1.Hash, 123)
	assert.Equal(t, testutil.Hash1.Hex+" 123\n", string(buf.Bytes()))
}
