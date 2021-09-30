package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/nahumsa/go-blockchain/blockchain"
	"github.com/nahumsa/go-blockchain/network"
	"github.com/nahumsa/go-blockchain/wallet"
)

// CommandLine has all the command line interface for the blockchain
type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADRESS - get balance for the adress")
	fmt.Println(" createblockchain -address ADRESS - creates a block chain in that adress")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send amount from an adress to another adress. Then -mine flag mines the node")
	fmt.Println(" createwallet - Creates a new wallet")
	fmt.Println(" listaddresses - List the addresses in our wallet file")
	fmt.Println(" reindexutxo - Rebuilds the UTXO set")
	fmt.Println(" startnode -miner ADDESS - Start a node with ID Specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}

	network.StartServer(nodeID, minerAddress)
}

func (cli *CommandLine) reindexUTXO(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()

	iter := chain.Iterator()
	for {
		block := iter.Next()

		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address, nodeID string) {
	if !wallet.MakeWallet().ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.InitBlockchain(address, nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	UTXOSet.Reindex()
	chain.Database.Close()

	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address, nodeID string) {
	if !wallet.MakeWallet().ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(from) {
		log.Panic("From address is not valid")
	}

	if !wallet.ValidateAddress(to) {
		log.Panic("To address is not valid")
	}

	chain := blockchain.ContinueBlockChain(nodeID)
	UTXOSet := blockchain.UTXOSet{Blockchain: chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	blockchain.Handle(err)

	wallet := wallets.GetWallet(from)
	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("Send tx")
	}

	fmt.Printf("successfully tranfered %d from: %s to: %s\n", amount, from, to)
}

func (cli *CommandLine) listAddresses(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet()

	wallets.SaveFile(nodeID)

	fmt.Printf("New address is: %s\n", address)
}

// Run runs the CLI
func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env is not set!")
		runtime.Goexit()
	}

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	starteNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send the genesis block")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	startNodeMiner := sendCmd.String("Miner", "", "Enable Mining mode and send reward to miner")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "startnode":
		err := starteNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}

		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if starteNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			starteNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
}
