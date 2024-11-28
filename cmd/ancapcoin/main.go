package main

import (
	"example.com/ancapcoin/pkg/api"
	"example.com/ancapcoin/pkg/blockchain"
	"example.com/ancapcoin/pkg/transaction"
	"example.com/ancapcoin/pkg/wallet"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	currentWallet      *wallet.Wallet
	blockchainInstance *blockchain.Blockchain
	mempool            map[string]*transaction.Transaction // Mempool para transacciones no confirmadas
	lightningChannels  []*lightning.LightningChannel        // Canales de Lightning Network
	mu                 sync.Mutex                          // Para manejar concurrencia en mempool
	p2pNode            *p2p.Node                           // Nodo P2P
)


func main() {
	// Inicializar variables globales
	currentWallet = wallet.NewWallet()
	mempool = make(map[string]*transaction.Transaction)

	// Configurar nodo P2P
	p2pNode = p2p.NewNode("localhost:3000") // Dirección del nodo local

	// Configurar certificados para TLS
	certFile := "cert.pem"
	keyFile := "key.pem"
	go func() {
		log.Println("Starting P2P node with TLS...")
		p2pNode.StartTLS(certFile, keyFile)
	}()

	// Configurar descubrimiento dinámico
	p2pNode.DiscoveryURL = "http://discovery-service.local/nodes"
	go func() {
		for {
			p2pNode.DiscoverPeers()
			time.Sleep(30 * time.Second) // Descubrir nodos cada 30 segundos
		}
	}()

	// Configurar sincronización SPV
	go func() {
		peerAddress := "peer-address" // Reemplazar con la dirección de un nodo conocido
		log.Printf("Starting SPV synchronization with peer: %s", peerAddress)
		p2pNode.SyncSPV(peerAddress)
	}()

	// Solicitar el tipo de consenso
	var consensusType string
	fmt.Println("Welcome to AncapCoin!")
	fmt.Println("Select the consensus mechanism:")
	fmt.Println("1. Proof of Work (PoW)")
	fmt.Println("2. Proof of Stake (PoS)")
	fmt.Println("3. Delegated Proof of Stake (DPoS)")
	fmt.Println("4. Proof of Authority (PoA)")
	fmt.Print("Enter your choice: ")

	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil || choice < 1 || choice > 4 {
		log.Fatal("Invalid choice. Exiting.")
	}

	switch choice {
	case 1:
		consensusType = "pow"
	case 2:
		consensusType = "pos"
	case 3:
		consensusType = "dpos"
	case 4:
		consensusType = "poa"
	default:
		log.Fatal("Invalid consensus type. Exiting.")
	}

	// Crear el bloque génesis con el tipo de consenso seleccionado
	coinbase := transaction.NewCoinbaseTransaction(currentWallet.GetAddress(), "Genesis block")
	blockchainInstance = blockchain.NewBlockchain(coinbase, consensusType)

	// Iniciar el servidor API en una goroutine separada
	go startAPI()

	// Menú interactivo
	for {
		fmt.Println("\n========== AncapCoin ==========")
		fmt.Println("1. Create a new wallet")
		fmt.Println("2. Show the current wallet address")
		fmt.Println("3. Add a block to the blockchain")
		fmt.Println("4. Print the blockchain")
		fmt.Println("5. Create a transaction")
		fmt.Println("6. View pending transactions (Mempool)")
		fmt.Println("7. Register as a validator (PoS only)")
		fmt.Println("8. Vote for a delegate (DPoS only)")
		fmt.Println("9. Create a Lightning channel")
		fmt.Println("10. Send a Lightning payment")
		fmt.Println("11. Create a smart contract")
		fmt.Println("12. Execute a smart contract function")
		fmt.Println("13. Exit")
		fmt.Print("Select an option: ")

		var option int
		_, err := fmt.Scan(&option)
		if err != nil {
			log.Println("Invalid input. Please try again.")
			continue
		}

		switch option {
		case 1:
			createNewWallet()
		case 2:
			showCurrentWalletAddress()
		case 3:
			addBlockToBlockchain()
		case 4:
			printBlockchain()
		case 5:
			createTransaction() // Crear una nueva transacción
		case 6:
			viewPendingTransactions() // Ver transacciones pendientes en el mempool
		case 7:
			if consensusType == "pos" {
				registerValidator()
			} else {
				fmt.Println("This option is only available for Proof of Stake (PoS).")
			}
		case 8:
			if consensusType == "dpos" {
				voteForDelegate()
			} else {
				fmt.Println("This option is only available for Delegated Proof of Stake (DPoS).")
			}
		case 9:
			createLightningChannel()
		case 10:
			sendLightningPayment()
		case 11:
			createSmartContract()
		case 12:
			executeSmartContract()
		case 13:
			fmt.Println("Thank you for using AncapCoin!")
			os.Exit(0)
		case 14: // Retransmitir transacciones no confirmadas
			retransmitTransactions()
		
		case 15: // Crear una propuesta
			createNewProposal()
		
		case 16: // Votar por una propuesta
			voteForProposal()
		
		case 17: // Ver propuestas existentes
			viewProposals()
		
		case 18: // Cerrar una propuesta
			closeExistingProposal()
		default:
			fmt.Println("Invalid option. Please try again.")
		
	}
	}
}


func registerValidator() {
    if blockchainInstance == nil {
        fmt.Println("Blockchain is not initialized.")
        return
    }

    var stake int64
    fmt.Print("Enter the amount of stake to register: ")
    _, err := fmt.Scan(&stake)
    if err != nil || stake <= 0 {
        fmt.Println("Invalid stake amount. Please try again.")
        return
    }

    err = blockchainInstance.RegisterValidator(currentWallet.GetAddress(), stake)
    if err != nil {
        fmt.Printf("Failed to register as validator: %v\n", err)
        return
    }

    fmt.Println("Successfully registered as a validator!")
}

// Create a new wallet
func createNewWallet() {
	currentWallet = wallet.NewWallet()
	fmt.Println("New wallet created.")
	fmt.Printf("Address of the new wallet: %s\n", currentWallet.GetAddress())
}

// Show the current wallet address
func showCurrentWalletAddress() {
	if currentWallet == nil {
		fmt.Println("No active wallet. Create one first.")
		return
	}
	fmt.Printf("Current wallet address: %s\n", currentWallet.GetAddress())
}

// Add a block to the blockchain with transactions from the mempool
func addBlockToBlockchain() {
	if blockchainInstance == nil {
		fmt.Println("Blockchain is not initialized.")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if len(mempool) == 0 {
		fmt.Println("No transactions in the mempool.")
		return
	}

	var transactions []*transaction.Transaction
	for _, tx := range mempool {
		transactions = append(transactions, tx)
	}
	blockchainInstance.AddBlock(transactions)

	// Clear mempool after mining
	mempool = make(map[string]*transaction.Transaction)

	fmt.Println("Block successfully mined and added to the blockchain.")
}

// Print the blockchain
func printBlockchain() {
	if blockchainInstance == nil {
		fmt.Println("Blockchain is not initialized.")
		return
	}

	fmt.Println("\n=== Blockchain ===")
	for _, block := range blockchainInstance.GetAllBlocks() {
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Previous Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Transactions:\n")
		for _, tx := range block.Transactions {
			fmt.Printf("\tID: %x\n", tx.ID)
			for _, input := range tx.Vin {
				fmt.Printf("\t\tInput: TxID: %x, Out: %d\n", input.Txid, input.Vout)
			}
			for _, output := range tx.Vout {
				fmt.Printf("\t\tOutput: Value: %d, Script: %s\n", output.Value, output.ScriptPubKey)
			}
		}
		fmt.Println("----------------------------")
	}
}

// Create a new transaction and add it to the mempool
func createTransaction() {
	if currentWallet == nil || blockchainInstance == nil {
		fmt.Println("Blockchain or wallet are not initialized.")
		return
	}

	var to string
	var amount int
	fmt.Print("Enter the destination address: ")
	_, err := fmt.Scan(&to)
	if err != nil {
		log.Println("Invalid input. Please try again.")
		return
	}

	fmt.Print("Enter the amount to transfer: ")
	_, err = fmt.Scan(&amount)
	if err != nil {
		log.Println("Invalid input. Please try again.")
		return
	}

	tx := transaction.NewTransaction(currentWallet.GetAddress(), to, amount, blockchainInstance.GetUTXOSet())
	mu.Lock()
	mempool[string(tx.ID)] = tx
	mu.Unlock()

	fmt.Println("Transaction successfully added to the mempool:")
	fmt.Printf("ID: %x\n", tx.ID)
	fmt.Printf("From: %s\n", currentWallet.GetAddress())
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Amount: %d\n", amount)
}


// View pending transactions in the mempool
func viewPendingTransactions() {
	mu.Lock()
	defer mu.Unlock()

	if len(mempool) == 0 {
		fmt.Println("No pending transactions.")
		return
	}

	fmt.Println("\n=== Pending Transactions ===")
	for _, tx := range mempool {
		fmt.Printf("ID: %x\n", tx.ID)
		fmt.Printf("From: %s\n", currentWallet.GetAddress())
		for _, output := range tx.Vout {
			fmt.Printf("To: %s, Amount: %d\n", output.ScriptPubKey, output.Value)
		}
		fmt.Println("----------------------------")
	}
}

// Start the API server
func startAPI() {
	apiServer := api.NewAPI(blockchainInstance)
	apiServer.Start()
}

func NewBlockchain(coinbase *transaction.Transaction, consensusType string) *Blockchain {
    var consensus Consensus
    switch consensusType {
    case "pow":
        consensus = NewProofOfWork(4) // Dificultad ajustada
    case "pos":
        consensus = NewProofOfStake()
    case "dpos":
        consensus = NewDelegatedProofOfStake()
    case "poa":
        consensus = NewProofOfAuthority()
    default:
        log.Panic("unsupported consensus type")
    }

    genesis := NewGenesisBlock(coinbase)
    return &Blockchain{
        blocks:    []*Block{genesis},
        consensus: consensus,
    }
}

func voteForDelegate() {
    if blockchainInstance == nil {
        fmt.Println("Blockchain is not initialized.")
        return
    }

    var delegateAddress string
    fmt.Print("Enter the address of the delegate to vote for: ")
    _, err := fmt.Scan(&delegateAddress)
    if err != nil {
        fmt.Println("Invalid input. Please try again.")
        return
    }

    err = blockchainInstance.VoteForDelegate(currentWallet.GetAddress(), delegateAddress)
    if err != nil {
        fmt.Printf("Failed to vote for delegate: %v\n", err)
        return
    }

    fmt.Println("Successfully voted for the delegate!")
}

var lightningChannels []*lightning.LightningChannel

func createLightningChannel() {
	var participant1, participant2 string
	var balance1, balance2 int64

	fmt.Print("Enter address of participant 1: ")
	fmt.Scan(&participant1)
	fmt.Print("Enter balance of participant 1: ")
	fmt.Scan(&balance1)
	fmt.Print("Enter address of participant 2: ")
	fmt.Scan(&participant2)
	fmt.Print("Enter balance of participant 2: ")
	fmt.Scan(&balance2)

	channel := lightning.NewLightningChannel(participant1, participant2, balance1, balance2)
	lightningChannels = append(lightningChannels, channel)

	fmt.Println("Lightning channel created successfully!")
}

func sendLightningPayment() {
	var from, to string
	var amount int64

	fmt.Print("Enter your address: ")
	fmt.Scan(&from)
	fmt.Print("Enter recipient address: ")
	fmt.Scan(&to)
	fmt.Print("Enter amount to send: ")
	fmt.Scan(&amount)

	for _, channel := range lightningChannels {
		if (channel.Participants[0] == from && channel.Participants[1] == to) ||
			(channel.Participants[1] == from && channel.Participants[0] == to) {
			err := channel.SendPayment(from, amount)
			if err != nil {
				fmt.Printf("Payment failed: %v\n", err)
			} else {
				fmt.Println("Payment sent successfully!")
			}
			return
		}
	}

	fmt.Println("No channel found between the given addresses.")
}

func createToken() {
	var name, symbol string
	var totalSupply int64
	fmt.Print("Enter token name: ")
	fmt.Scan(&name)
	fmt.Print("Enter token symbol: ")
	fmt.Scan(&symbol)
	fmt.Print("Enter total supply: ")
	fmt.Scan(&totalSupply)

	token := token.NewToken(name, symbol, totalSupply, currentWallet.GetAddress())
	fmt.Printf("Token '%s' (%s) created successfully!\n", token.Name, token.Symbol)
}

func transferToken() {
	var tokenName, to string
	var amount int64
	fmt.Print("Enter token name: ")
	fmt.Scan(&tokenName)
	fmt.Print("Enter recipient address: ")
	fmt.Scan(&to)
	fmt.Print("Enter amount to transfer: ")
	fmt.Scan(&amount)

	err := token.Transfer(currentWallet.GetAddress(), to, amount)
	if err != nil {
		fmt.Printf("Token transfer failed: %v\n", err)
		return
	}
	fmt.Println("Token transfer successful!")
}

func createSmartContract() {
	var code string
	fmt.Print("Enter contract code: ")
	fmt.Scan(&code)

	contract := blockchain.NewSmartContract(code, currentWallet.GetAddress())
	fmt.Println("Smart contract created successfully!")
}

func executeSmartContract() {
	var address, functionName string
	fmt.Print("Enter contract address: ")
	fmt.Scan(&address)
	fmt.Print("Enter function name: ")
	fmt.Scan(&functionName)

	// Lógica para pasar argumentos a la función
	args := []interface{}{}
	// (Implementa lógica para capturar argumentos dinámicos)

	result, err := contract.ExecuteFunction(functionName, args...)
	if err != nil {
		fmt.Printf("Failed to execute function: %v\n", err)
		return
	}
	fmt.Printf("Function executed successfully. Result: %v\n", result)
}

func startAPIServer() {
	http.HandleFunc("/wallet", getWalletInfo)
	http.HandleFunc("/sendTransaction", sendTransactionAPI)
	http.HandleFunc("/createSmartContract", createSmartContractAPI)
	http.HandleFunc("/executeSmartContract", executeSmartContractAPI)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getWalletInfo(w http.ResponseWriter, r *http.Request) {
	walletInfo := map[string]interface{}{
		"address": currentWallet.GetAddress(),
		"balance": blockchainInstance.GetBalance(currentWallet.GetAddress()),
	}
	json.NewEncoder(w).Encode(walletInfo)
}

func sendTransactionAPI(w http.ResponseWriter, r *http.Request) {
	var txData struct {
		To     string `json:"to"`
		Amount int    `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&txData)

	tx := transaction.NewTransaction(currentWallet.GetAddress(), txData.To, txData.Amount, blockchainInstance.GetUTXOSet())
	blockchainInstance.AddTransaction(tx)
	w.WriteHeader(http.StatusOK)
}

func createSmartContractAPI(w http.ResponseWriter, r *http.Request) {
	var contractData struct {
		Code string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&contractData)

	contract := blockchain.NewSmartContract(contractData.Code, currentWallet.GetAddress())
	blockchainInstance.AddContract(contract)
	w.WriteHeader(http.StatusOK)
}

func executeSmartContractAPI(w http.ResponseWriter, r *http.Request) {
	var execData struct {
		ContractAddress string        `json:"contractAddress"`
		FunctionName    string        `json:"functionName"`
		Args            []interface{} `json:"args"`
	}
	json.NewDecoder(r.Body).Decode(&execData)

	contract := blockchainInstance.GetContract(execData.ContractAddress)
	result, err := contract.ExecuteFunction(execData.FunctionName, execData.Args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
}

func startExplorer(blockchainInstance *blockchain.Blockchain) {
	explorer := explorer.Explorer{Blockchain: blockchainInstance}
	http.HandleFunc("/explorer/blocks", explorer.GetLatestBlocks)
	http.HandleFunc("/explorer/block", explorer.GetBlockDetails)
}

func retransmitTransactions() {
	mu.Lock()
	defer mu.Unlock()

	if len(mempool) == 0 {
		fmt.Println("No unconfirmed transactions to retransmit.")
		return
	}

	for _, pt := range mempool {
		p2pNode.BroadcastTransaction(pt.Tx)
	}
	fmt.Println("All unconfirmed transactions have been retransmitted.")
}