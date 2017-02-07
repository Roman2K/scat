package stripe_test

import (
	"testing"

	"gitlab.com/Roman2K/scat/stripe"

	assert "github.com/stretchr/testify/require"
)

func TestStripeRR(t *testing.T) {
	seq := stripe.RR{}
	assert.Equal(t, nil, seq.Next())
	assert.Equal(t, nil, seq.Next())

	seq = stripe.RR{Items: []interface{}{
		1,
	}}
	assert.Equal(t, 1, seq.Next())
	assert.Equal(t, 1, seq.Next())

	seq = stripe.RR{Items: []interface{}{
		1,
		2,
	}}
	assert.Equal(t, 1, seq.Next())
	assert.Equal(t, 2, seq.Next())
	assert.Equal(t, 1, seq.Next())
}
