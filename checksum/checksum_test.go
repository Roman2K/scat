package checksum_test

import (
	"fmt"
	"scat/checksum"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestChecksum(t *testing.T) {
	const (
		valid = "348df4eb47f9230bfe89637afe7409bec883424d822257b6cbbce93ee780d992"
	)
	var h checksum.Hash
	err := h.LoadSlice([]byte{1, 2, 3})
	assert.Error(t, err)
	slice := []byte{}
	_, err = fmt.Sscanf(valid, "%x", &slice)
	assert.NoError(t, err)
	err = h.LoadSlice(slice)
	assert.NoError(t, err)
	assert.Equal(t, valid, fmt.Sprintf("%x", h))
}
