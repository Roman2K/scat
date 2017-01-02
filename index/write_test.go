package index_test

import (
	"bytes"
	"secsplit/checksum"
	"secsplit/index"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	var (
		hex  = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
		hash = checksum.Hash{
			44, 242, 77, 186, 95, 176, 163, 14, 38, 232, 59, 42, 197, 185, 226, 158,
			27, 22, 30, 92, 31, 167, 66, 94, 115, 4, 51, 98, 147, 139, 152, 36,
		}
	)
	buf := &bytes.Buffer{}
	index.Write(buf, hash, 123)
	assert.Equal(t, hex+" 123\n", string(buf.Bytes()))
}
