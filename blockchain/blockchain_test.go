package blockchain

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dbTestPath = "test_blocks"
)

func TestInitBlockChain(t *testing.T) {

	err := os.Mkdir(dbTestPath, 0755)
	Handle(err)
	defer os.RemoveAll(dbTestPath)

	chain := InitBlockchain(dbTestPath)
	defer chain.Database.Close()

	iter := chain.Iterator()
	for {
		block := iter.Next()

		assert.Equal(t, string(block.Data), "Genesis", "Genesis data doesn't match")

		if len(block.PrevHash) == 0 {
			break
		}
	}

}

func TestInitAddBlock(t *testing.T) {
	err := os.Mkdir(dbTestPath, 0755)
	Handle(err)
	defer os.RemoveAll(dbTestPath)

	chain := InitBlockchain(dbTestPath)
	defer chain.Database.Close()

	chain.AddBlock("test block")

	iter := chain.Iterator()
	i := 1
	expectedValues := []string{"Genesis", "test block"}

	for {
		block := iter.Next()

		assert.Equal(t, string(block.Data), expectedValues[i], fmt.Sprintf("Block %d data doesn't match", i))

		if len(block.PrevHash) == 0 {
			break
		}

		i--
	}

}
