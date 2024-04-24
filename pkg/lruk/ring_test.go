package lruk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRing(t *testing.T) {
	r := NewRing[string](3)
	r.Append("pim")
	assert.Equal(t, 1, r.Length())
	assert.Equal(t, "pim", r.Last())
	r.Append("pam")
	assert.Equal(t, 2, r.Length())
	assert.Equal(t, "pam", r.Last())
	r.Append("poum")
	assert.Equal(t, 3, r.Length())
	r.Append("the captain")
	assert.Equal(t, 3, r.Length())
	h, err := r.Hist(1)
	assert.NoError(t, err)
	assert.Equal(t, "poum", h)
	_, err = r.Hist(3)
	assert.Error(t, err)
}
