package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp    int64
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
	Height       int
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	tree := NewMerkleTree(txHashes)

	return tree.RootNode.Data
}

func CreateBlock(txs []*Transaction, PrevHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), []byte{}, txs, PrevHash, 0, height}

	pow := NewProof(block)

	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	// something is wrong here
	err := decoder.Decode(&block)
	Handle(err)

	return &block
}

func Handle(err error) {

	if err != nil {
		log.Panic(err)
	}
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0)
}
