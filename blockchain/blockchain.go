package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v3"
)

const (
	dbFile      = "/MANIFEST"
	genesisData = "First Transaction From Genesis"
)

// BlockChain ist the main struct for the blockchain package. It keeps the last hash and
// the badger database.
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// BCIterator is an auxiliary struct that works for iterating on the blockchain.
type BCIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// InitBlockchain initializes the blockchain and check if it already exists, if it does
// it will exit.
func InitBlockchain(dbPath, address string) *BlockChain {
	var lastHash []byte

	if DBexists(dbPath) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	db := initDB(dbPath)

	err := db.Update(func(txn *badger.Txn) error {
		coinbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(coinbtx)

		fmt.Println("Genesis block created")
		err := txn.Set(genesis.Hash, genesis.Serialize())

		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain

}

// ContinueBlockChain reinitializes the blockchain for a given address
func ContinueBlockChain(dbPath, address string) *BlockChain {
	if !DBexists(dbPath) {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	db := initDB(dbPath)

	err := db.Update(func(txn *badger.Txn) error {

		item, err := txn.Get([]byte("lh"))
		Handle(err)

		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func initDB(dbPath string) *badger.DB {
	opts := badger.DefaultOptions(dbPath).WithLogger(nil)

	db, err := badger.Open(opts)
	Handle(err)

	return db
}

// DBexists returns true if a database exists on the dbPath folder
func DBexists(dbPath string) bool {
	if _, err := os.Stat(dbPath + dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// AddBlock adds a block given a slice of transactions
func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)

		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

		return err
	})

	Handle(err)

	newBlock := CreateBlock(transactions, lastHash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

// Iterator generates a iterator for the blockchain
func (chain *BlockChain) Iterator() *BCIterator {
	iter := &BCIterator{chain.LastHash, chain.Database}

	return iter
}

// Next returns the next block on the blockchain
func (iter *BCIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		var encodedBlock []byte

		item, err := txn.Get(iter.CurrentHash)

		Handle(err)

		err = item.Value(func(val []byte) error {
			encodedBlock = append([]byte{}, val...)
			return nil
		})

		block = block.Deserialize(encodedBlock)

		return err
	})

	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}

// FindUnspentTransactions retuns a slice of Transaction for all unspent transactions
func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTxs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

	Outputs:
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			for outID, out := range tx.Outputs {
				if spentTxs[txID] != nil {
					for _, spentOut := range spentTxs[txID] {
						if spentOut == outID {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTxs[inTxID] = append(spentTxs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

// FindUTXO returns a slice of all unspended transactions
func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var unspentTxs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				unspentTxs = append(unspentTxs, out)
			}
		}
	}

	return unspentTxs
}

// FindSpendableOutputs retuns the number of spendable outputs and a map for each address and the amount
// that can be spended
func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outID, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outID)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) bool {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
