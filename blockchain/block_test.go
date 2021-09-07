package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBlock(t *testing.T) {
	input := "test1"
	block := CreateBlock(input, []byte{})

	assert.Equal(t, string(block.Data), input, "Data doesn't match")

}
