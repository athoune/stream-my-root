package local

import (
	"os"
	"testing"

	"github.com/athoune/stream-my-root/pkg/blocks"
	"github.com/stretchr/testify/assert"
)

func FuzzReader(f *testing.F) {
	recipe_f, err := os.Open("../../../fixtures/gcr.io_distroless_static-debian12.img.recipe")
	assert.NoError(f, err)
	recipe, err := blocks.ReadRecipe(recipe_f)
	assert.NoError(f, err)
	r, err := NewLocalReader(&LocalReaderOpts{
		CacheDirectory: "../../../fixtures/chunks",
		Tainted:        false,
	})
	assert.NoError(f, err)
	f.Add(uint64(2048), uint64(2048))
	f.Fuzz(func(t *testing.T, buffer_size, off uint64) {
		buffer := make([]byte, buffer_size)
		n, err := blocks.Read(buffer, int64(off), recipe, r)
		assert.NoError(t, err)
		if len(buffer) == 0 {
			assert.Equal(t, 0, n)
		} else {
			assert.True(t, n > 0, n)
		}
	})
}
