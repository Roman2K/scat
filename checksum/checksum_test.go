package checksum_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Roman2K/scat/checksum"
	assert "github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
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

func TestSum(t *testing.T) {
	const (
		data = "abc"
		hex  = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	)
	h, err := checksum.Sum(strings.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, hex, fmt.Sprintf("%x", h))
}
