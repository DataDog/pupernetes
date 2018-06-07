package setup

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMajorMinorVersion(t *testing.T) {
	for _, tc := range []struct {
		expected string
		given    string
	}{
		{
			"1.3",
			"1.3.5",
		},
		{
			"1.10",
			"1.10.2",
		},
		{
			"1.11",
			"1.11.0",
		},
		{
			"1.11",
			"1.11",
		},
	} {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.expected, getMajorMinorVersion(tc.given))
		})
	}
}
