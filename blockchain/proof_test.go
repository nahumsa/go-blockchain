package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHex(t *testing.T) {
	assert.Equal(t, ToHex(10), []uint8{0, 0, 0, 0, 0, 0, 0, 10})
}
