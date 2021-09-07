package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitBlockChain(t *testing.T) {

	blockchain := InitBlockchain()
	genesisBlock := blockchain.Blocks[0]
	assert.Equal(t, string(genesisBlock.Data), "Genesis", "Genesis data doesn't match")

}

func TestAddBlock(t *testing.T) {
	blockchain := InitBlockchain()
	blockchain.AddBlock("test1")
	lastBlock := blockchain.Blocks[1]

	assert.Equal(t, string(lastBlock.Data), "test1", "Added block data doesn't match")

}
