package argproc_test

import (
	"testing"

	"gitlab.com/Roman2K/scat/argproc"
)

func TestNew(t *testing.T) {
	// just test that it compiles
	argproc.New(nil, nil)
}
