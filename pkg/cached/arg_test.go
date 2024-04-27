package cached

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetArg(t *testing.T) {
	buff := &bytes.Buffer{}
	buff.Write([]byte{0, 0, 0, 9})
	buff.WriteString("plop")

	arg := &SetArg{
		Key:  "plop",
		Size: 9,
	}

	raw, err := arg.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, buff.Bytes(), raw)

	arg2 := &SetArg{}
	err = arg2.UnmarshalBinary(raw)
	assert.NoError(t, err)

	assert.Equal(t, arg, arg2)
}

func TestBool(t *testing.T) {
	b := Bool(false)
	raw, err := b.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0}, raw)
	err = b.UnmarshalBinary([]byte{1})
	assert.NoError(t, err)
	assert.True(t, bool(b))
}
