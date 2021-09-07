package main

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/nahumsa/go-blockchain/blockchain"
)

func main() {
	chain := blockchain.InitBlockchain()

	chain.AddBlock("1st block")
	chain.AddBlock("2nd Block")
	chain.AddBlock("3rd block")

	for _, block := range chain.Blocks {

		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

	}

	fmt.Println(reflect.TypeOf(blockchain.ToHex(10)))
	fmt.Println(blockchain.ToHex(20))
}
