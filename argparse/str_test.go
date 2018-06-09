package argparse_test

import (
	"testing"

	"github.com/Roman2K/scat/argparse"
	assert "github.com/stretchr/testify/require"
)

func TestArgStr(t *testing.T) {
	tests := []struct {
		input    string
		expect   string
		consumed int
	}{
		{`abc`, "abc", 3},
		{` abc `, "", 0},
		{`  `, "", 0},
		{` "aoeu" `, "", 0},

		// Quotes accepted and removed
		{`"aoeu"`, `aoeu`, 6},

		// Malformed quotes aren't accepted
		// (reducing risks of something that used to work
		// continuing to work but doing something different)
		{`"aoeu`, `"aoeu`, 5},
		{`"aoeu"a`, `"aoeu"a`, 7},
		// No backslash escaping (otherwise windows paths get weird)
		{`"aoeu\""`, `"aoeu\""`, 8},
		{`"aoeu\" "`, `aoeu\`, 7},
	}
	for _, tt := range tests {
		res, n, err := argparse.ArgStr.Parse(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expect, res.(string))
		assert.Equal(t, tt.consumed, n)
	}
}
