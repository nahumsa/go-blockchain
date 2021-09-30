package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/dgraph-io/badger/v3"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLenght = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *BlockChain
}

// Reindex clears out all outputs with utxoPrefix and rebuild the set inside the database
func (u UTXOSet) Reindex() {
	db := u.Blockchain.Database

	// clear db with utxoPrefix
	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.Blockchain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error {
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			key = append(utxoPrefix, key...)

			err = txn.Set(key, outs.Serialize())
			Handle(err)
		}

		return nil
	})

	Handle(err)

}

func (u *UTXOSet) Update(block *Block) {

	fmt.Println("PASS")
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)

					item, err := txn.Get(inID)
					Handle(err)

					var v []byte

					err = item.Value(func(val []byte) error {
						v = append([]byte{}, val...)
						return nil
					})

					Handle(err)

					outs := DeserializeOutputs(v)

					for outIDx, out := range outs.Outputs {
						if outIDx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}
					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	Handle(err)
}

// CountTransactions counts all UTXO transactions
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	Handle(err)

	return counter

}

// FindUTXO retuns a slice of Transaction for all unspent transactions
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	db := u.Blockchain.Database
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			var v []byte

			err := item.Value(func(val []byte) error {
				v = append([]byte{}, val...)
				return nil
			})

			Handle(err)
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	Handle(err)
	return UTXOs
}

// FindSpendableOutputs retuns the number of spendable outputs and a map for each address and the amount
// that can be expended
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			var v []byte

			err := item.Value(func(val []byte) error {
				v = append([]byte{}, val...)
				return nil
			})

			Handle(err)
			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	Handle(err)
	return accumulated, unspentOuts
}

// DeleteByPrefix will delete all keys that has utxoPrefix
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {

		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}

		return nil

	}

	collectSize := 100_000
	u.Blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}
