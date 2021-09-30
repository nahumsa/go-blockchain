package blockchain

// import (
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/nahumsa/go-blockchain/wallet"
// 	"github.com/stretchr/testify/assert"
// )

// const (
// 	dbTestPath  = "test_blocks"
// 	from        = "A"
// 	to          = "B"
// 	amount      = 5
// 	initBalance = 100
// )

// func setupDB(t *testing.T) func(t *testing.T) {
// 	err := os.Mkdir(dbTestPath, 0755)
// 	Handle(err)

// 	return func(t *testing.T) {
// 		os.RemoveAll(dbTestPath)
// 	}
// }

// // TestBlockchain tests initializes, add a block to the blockchain and see if
// // those blocks are valid, check the balance of the sender and the receiver
// func TestBlockchain(t *testing.T) {

// 	teardownDB := setupDB(t)
// 	defer teardownDB(t)

// 	t.Run("Test the initialization of the blockchain", testInitBlockchain)
// 	t.Run("Test the value of the initial adress", testBalanceBeforeTransaction)
// 	t.Run("Test adding a block on the blockchain", testAddBlock)
// 	t.Run("Test the value of the sender after the transaction", testBalanceSender)
// 	t.Run("Test the value of the receiver after the transaction", testBalanceReceiver)
// }

// func testInitBlockchain(t *testing.T) {

// 	chain := InitBlockchain(dbTestPath, from)
// 	defer chain.Database.Close()

// 	iter := chain.Iterator()

// 	for {
// 		block := iter.Next()

// 		pow := NewProof(block)

// 		assert.True(t, pow.Validate(), "Genesis block not valid")

// 		if len(block.PrevHash) == 0 {
// 			break
// 		}
// 	}

// }

// func testAddBlock(t *testing.T) {

// 	chain := ContinueBlockChain(dbTestPath, from)
// 	defer chain.Database.Close()

// 	tx := NewTransaction(from, to, amount, chain)
// 	chain.AddBlock([]*Transaction{tx})

// 	iter := chain.Iterator()
// 	i := 1

// 	for {
// 		block := iter.Next()

// 		pow := NewProof(block)
// 		assert.True(t, pow.Validate(), fmt.Sprintf("Block %d not valid", i))

// 		if len(block.PrevHash) == 0 {
// 			break
// 		}

// 		i--
// 	}

// }

// func getBalance(address string) int {
// 	chain := ContinueBlockChain(dbTestPath, address)
// 	defer chain.Database.Close()

// 	balance := 0

// 	pubKeyHash := wallet.Base58Decode([]byte(address))
// 	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
// 	UTXOs := chain.FindUTXO(pubKeyHash)

// 	for _, out := range UTXOs {
// 		balance += out.Value
// 	}

// 	return balance
// }

// func testBalanceBeforeTransaction(t *testing.T) {
// 	balanceA := getBalance(from)

// 	assert.Equal(t, initBalance, balanceA, "Balance for the initial block doesn't match")
// }

// func testBalanceSender(t *testing.T) {
// 	balanceSender := getBalance(from)

// 	assert.Equal(t, initBalance-amount, balanceSender, "Balance for the sender doesn't match")
// }

// func testBalanceReceiver(t *testing.T) {
// 	balanceReceiver := getBalance(to)

// 	assert.Equal(t, amount, balanceReceiver, "Balance for the receiver doesn't match")
// }
