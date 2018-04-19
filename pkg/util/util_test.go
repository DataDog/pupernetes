package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandStringBytesMaskImprSrc(t *testing.T) {
	var p, c string
	for i := 0; i < 10; i++ {
		c = RandStringBytesMaskImprSrc(20)
		assert.Len(t, c, 20)
		assert.NotEqual(t, c, p)
		p = c
	}
}
