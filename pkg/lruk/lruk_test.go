package lruk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUK(t *testing.T) {
	l := New[string, string](2, time.Second, 16, func(s string) int {
		return len(s)
	}, nil)
	now := time.Now()
	l.AddAt("pim", "stole a pie", now)
	_, ok := l.Get("pam")
	assert.False(t, ok)
	_, ok = l.GetAt("pim", now)
	assert.True(t, ok)
	assert.Equal(t, now, l.history["pim"].Last())
	assert.Equal(t, 1, l.history["pim"].Length())
	_, ok = l.GetAt("pim", now.Add(2*time.Second))
	assert.True(t, ok)
	assert.Equal(t, 2, l.history["pim"].Length())
	l.AddAt("the captain", "lost his pipe", now)
}
