package blockchain

import (
	"example.com/ancapcoin/pkg/transaction"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	GenesisData     = "19/11/2023 the day we reclaimed freedom"
	Subsidy         = 50
	HalvingInterval = 210000
)

const InitialReward = int64(50) // Recompensa inicial por bloque
const HalvingInterval = 210000 // Número de bloques para reducir a la mitad
const MaxSupply = int64(21000000) // Máximo suministro de monedas

var TotalSupply = int64(0

const (
    BaseFee      = 10         // Tarifa base mínima
    CongestionFactor = 2      // Multiplicador por congestión
)

type DAO struct {
    Balance      int64
    Proposals    map[string]*Proposal
    ApprovedLogs []string
}

type Proposal struct {
    ID        string
    Title     string
    Amount    int64
    Recipient string
    Votes     map[string]bool // Votos de los usuarios
    Approved  bool
}

var CommunityFund = &DAO{
    Balance:   0,
    Proposals: make(map[string]*Proposal),
}

type PriorityTransaction struct {
    Tx  *Transaction // Transacción
    Fee int64        // Tarifa asociada
}

type Mempool struct {
    transactions []*PriorityTransaction // Usamos un heap para manejar prioridades
}

type CrossChainEvent struct {
    OriginChain      string `json:"origin_chain"`
    DestinationChain string `json:"destination_chain"`
    Asset            string `json:"asset"`
    Amount           int64  `json:"amount"`
    Sender           string `json:"sender"`
    Receiver         string `json:"receiver"`
    Timestamp        int64  `json:"timestamp"`
    Signature        string `json:"signature"` // Opcional, para seguridad adicional
}

type CrossChainState struct {
    Events []CrossChainEvent `json:"events"`
}

type Proposal struct {
    ID          string
    Title       string
    Description string
    Author      string
    Votes       map[string]int
    CreatedAt   int64
    Status      string
}

var Proposals = make(map[string]*Proposal)
var proposalMutex sync.Mutex

// Block represents a single block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*transaction.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type TokenTransaction struct {
	Token   string 
	From    string 
	To      string 
	Amount  int64  
}

type Shard struct {
	ID        int
	Blockchain *Blockchain
}

type ShardedBlockchain struct {
	Shards map[int]*Shard // Mapa de ID de shard a fragmento de blockchain
	mu     sync.RWMutex   // Mutex para acceso seguro
}

// Blockchain maintains the chain of blocks with concurrency safety
type Blockchain struct {
	mu     sync.RWMutex // Mutex for thread-safe access
	blocks []*Block
}

var CrossChainStates = make(map[string]*CrossChainState)


// NewBlock creates a new block with transactions and executes Proof of Work
func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock creates the genesis block with a coinbase transaction
func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{})
}

// NewBlockchain initializes a blockchain with the genesis block
func NewBlockchain(coinbase *transaction.Transaction, consensusType string) *Blockchain {
	var consensus Consensus
	switch consensusType {
	case "pow":
		consensus = NewProofOfWork(4) // Usa dificultad 4 por defecto
	case "pos":
		consensus = NewProofOfStake()
	case "dpos":
		consensus = NewDelegatedProofOfStake([]string{"delegate1", "delegate2"})
	case "poa":
		consensus = NewProofOfAuthority([]string{"authority1", "authority2"})
	default:
		log.Panic("unsupported consensus type")
	}

	genesis := NewGenesisBlock(coinbase)
	return &Blockchain{
		blocks:    []*Block{genesis},
		consensus: consensus,
	}
}

// AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(transactions []*transaction.Transaction) error {
	if len(transactions) == 0 {
		return errors.New("no transactions provided for the new block")
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(transactions, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
	return nil
}

// Validate checks the integrity of the entire blockchain
func (bc *Blockchain) Validate() error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.blocks); i++ {
		currentBlock := bc.blocks[i]
		prevBlock := bc.blocks[i-1]

		if !bytes.Equal(currentBlock.PrevBlockHash, prevBlock.Hash) {
			return errors.New("invalid previous block hash")
		}

		pow := NewProofOfWork(currentBlock)
		if !pow.Validate() {
			return errors.New("invalid proof of work")
		}
	}
	return nil
}

// FindTransaction retrieves a transaction by ID
func (bc *Blockchain) FindTransaction(ID []byte) (*transaction.Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return tx, nil
			}
		}
	}
	return nil, errors.New("transaction not found")
}

func (b *Blockchain) AddBlock(transactions []*Transaction) {
    height := len(b.Blocks)
    reward := CalculateBlockReward(height)
    if TotalSupply+reward > MaxSupply {
        reward = MaxSupply - TotalSupply // Ajusta si el suministro alcanza el límite
    }

    coinbaseTx := NewCoinbaseTransaction(minerAddress, reward)
    transactions = append([]*Transaction{coinbaseTx}, transactions...)

    newBlock := CreateBlock(transactions, b.GetLatestBlock().Hash)
    b.Blocks = append(b.Blocks, newBlock)
    TotalSupply += reward
}

func NewTransaction(from, to string, amount, fee int64, utxoSet map[string][]TXOutput) (*Transaction, error) {
    minFee := AdjustTransactionFee(len(Mempool))
    if fee < minFee {
        return nil, errors.New("transaction fee too low")
    }
}

func AdjustTransactionFee(mempoolSize int) int64 {
    baseFee := int64(100) // Tarifa base
    increment := int64(10) // Incremento por cada 10 transacciones en el mempool
    return baseFee + (increment * int64(mempoolSize/10))
}

// CalculateBlockReward computes the mining reward based on the height
func CalculateBlockReward(height int) int64 {
    halvings := height / HalvingInterval
    reward := InitialReward >> halvings // Reduce la recompensa a la mitad
    if reward <= 0 {
        return 0 // Detén la emisión cuando sea demasiado baja
    }
    return reward
}

func NewCoinbaseTransaction(to string, height int) *Transaction {
	reward := CalculateBlockReward(height)
	txin := TXInput{Txid: []byte{}, Vout: -1, Signature: nil, PubKey: []byte("Coinbase")}
	txout := TXOutput{Value: int64(reward), ScriptPubKey: to}
	tx := Transaction{Vin: []TXInput{txin}, Vout: []TXOutput{txout}}
	tx.ID = tx.Hash()
	return &tx
}

// Serialize converts a block into a byte slice
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(b); err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// DeserializeBlock reconstructs a block from a byte slice
func DeserializeBlock(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&block); err != nil {
		log.Panic(err)
	}
	return &block
}

// GetLatestBlock retrieves the most recent block
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blocks[len(bc.blocks)-1]
}

// GetBlockchainHeight returns the current height of the blockchain
func (bc *Blockchain) GetBlockchainHeight() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.blocks)
}

// GetAllBlocks retrieves all blocks in the blockchain
func (bc *Blockchain) GetAllBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return append([]*Block{}, bc.blocks...) // Copy for safety
}

// SerializeBlockchain serializes the entire blockchain
func (bc *Blockchain) SerializeBlockchain() []byte {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(bc.blocks); err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// DeserializeBlockchain reconstructs a blockchain from serialized data
func DeserializeBlockchain(data []byte) *Blockchain {
	var blocks []*Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&blocks); err != nil {
		log.Panic(err)
	}
	return &Blockchain{blocks: blocks}
}

func (pos *ProofOfStake) AddBlock(transactions []*transaction.Transaction, prevHash []byte, height int) (*Block, error) {
	validator := pos.selectValidator()
	log.Printf("Selected validator for block %d: %s", height, validator)
	block := NewBlock(transactions, prevHash, height)
	block.Hash = util.HashSHA256([]byte(validator))
	return block, nil
}

func (bc *Blockchain) SyncBlocks(peerBlocks []*Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for _, block := range peerBlocks {
		if block.Validate(bc.blocks[len(bc.blocks)-1]) == nil {
			bc.blocks = append(bc.blocks, block)
		}
	}
}

// NewShardedBlockchain crea una blockchain shardeada con un número específico de shards
func NewShardedBlockchain(numShards int, genesisTx *transaction.Transaction) *ShardedBlockchain {
	shardedBC := &ShardedBlockchain{
		Shards: make(map[int]*Shard),
	}

	for i := 0; i < numShards; i++ {
		genesisBlock := NewGenesisBlock(genesisTx)
		shardedBC.Shards[i] = &Shard{
			ID:         i,
			Blockchain: &Blockchain{blocks: []*Block{genesisBlock}},
		}
	}

	return shardedBC
}

// AddTransactionToShard añade una transacción a un shard específico basado en una función hash
func (sbc *ShardedBlockchain) AddTransactionToShard(tx *transaction.Transaction) error {
	shardID := hashToShard(tx.ID, len(sbc.Shards))
	shard := sbc.Shards[shardID]

	if shard == nil {
		return fmt.Errorf("shard %d not found", shardID)
	}

	return shard.Blockchain.AddBlock([]*transaction.Transaction{tx})
}

// hashToShard determina a qué shard pertenece una transacción
func hashToShard(txID []byte, numShards int) int {
	hash := sha256.Sum256(txID)
	return int(hash[0]) % numShards
}

func (b *Block) AddTokenTransaction(tx *TokenTransaction) {
	b.Transactions = append(b.Transactions, &transaction.Transaction{
		Type:   "token",
		Token:  tx.Token,
		From:   tx.From,
		To:     tx.To,
		Amount: tx.Amount,
	})
}

func (b *Block) AddContractTransaction(tx *transaction.ContractTransaction) {
	// Convierte la transacción a un formato generalizado y la añade al bloque
	b.Transactions = append(b.Transactions, &transaction.Transaction{
		Type:             "contract",
		ContractAddress:  tx.ContractAddress,
		Function:         tx.Function,
		Args:             tx.Args,
	})
}

func (m *Mempool) AddTransaction(tx *Transaction) {
    pt := &PriorityTransaction{Tx: tx, Fee: tx.Fee}
    heap.Push(m, pt) // Añadir la transacción al heap
}

func (m *Mempool) GetTopTransactions(limit int) []*Transaction {
    var topTransactions []*Transaction
    for i := 0; i < limit && m.Len() > 0; i++ {
        pt := heap.Pop(m).(*PriorityTransaction)
        topTransactions = append(topTransactions, pt.Tx)
    }
    return topTransactions
}

// Implementación del heap para manejar la prioridad
func (m *Mempool) Len() int { return len(m.transactions) }
func (m *Mempool) Less(i, j int) bool {
    // Mayor tarifa tiene mayor prioridad
    return m.transactions[i].Fee > m.transactions[j].Fee
}
func (m *Mempool) Swap(i, j int) {
    m.transactions[i], m.transactions[j] = m.transactions[j], m.transactions[i]
}
func (m *Mempool) Push(x interface{}) {
    m.transactions = append(m.transactions, x.(*PriorityTransaction))
}
func (m *Mempool) Pop() interface{} {
    old := m.transactions
    n := len(old)
    item := old[n-1]
    m.transactions = old[0 : n-1]
    return item
}

func createProposal(author, title, description string) string { /* ... */ }

func voteOnProposal(proposalID, voter string, vote int) error { /* ... */ }

func closeProposal(proposalID string) error { /* ... */ }


func SerializeProposals() []byte {
    proposalMutex.Lock()
    defer proposalMutex.Unlock()

    var buffer bytes.Buffer
    encoder := gob.NewEncoder(&buffer)
    err := encoder.Encode(Proposals)
    if err != nil {
        log.Panic(err)
    }
    return buffer.Bytes()
}

func DeserializeProposals(data []byte) {
    proposalMutex.Lock()
    defer proposalMutex.Unlock()

    decoder := gob.NewDecoder(bytes.NewReader(data))
    err := decoder.Decode(&Proposals)
    if err != nil {
        log.Panic(err)
    }
}

func (bc *Blockchain) AddBlockWithReward(transactions []*transaction.Transaction, minerAddress string) error {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    height := len(bc.blocks)
    reward := CalculateBlockReward(height)
    if reward > 0 {
        coinbase := transaction.NewCoinbaseTransaction(minerAddress, reward)
        transactions = append([]*transaction.Transaction{coinbase}, transactions...)
    }

    prevBlock := bc.blocks[len(bc.blocks)-1]
    newBlock := NewBlock(transactions, prevBlock.Hash)
    bc.blocks = append(bc.blocks, newBlock)

    return nil
}

func AdjustFees(mempoolSize int) int64 {
    if mempoolSize < 100 {
        return BaseFee
    }
    return int64(BaseFee + CongestionFactor*(mempoolSize/100))
}

// IncludeHighPriorityTransactions selects the highest fee transactions
func IncludeHighPriorityTransactions(mempool []*transaction.Transaction, limit int) []*transaction.Transaction {
    // Ordenar transacciones por tarifa (descendente)
    sort.Slice(mempool, func(i, j int) bool {
        return mempool[i].Fee > mempool[j].Fee
    })

    // Seleccionar las transacciones con mayor tarifa
    if len(mempool) > limit {
        return mempool[:limit]
    }
    return mempool
}

func (bc *Blockchain) MineBlock(minerAddress string, mempool []*transaction.Transaction) error {
    mempoolSize := len(mempool)
    adjustedFee := AdjustFees(mempoolSize)

    // Incluir transacciones de alta prioridad
    selectedTxs := IncludeHighPriorityTransactions(mempool, 100)

    for _, tx := range selectedTxs {
        tx.Fee += adjustedFee
    }

    // Agregar bloque con recompensa y transacciones
    return bc.AddBlockWithReward(selectedTxs, minerAddress)
}

func AddTransactionToMempool(tx *blockchain.Transaction) {
    Mempool = append(Mempool, tx)
}

func EmitCrossChainEvent(event CrossChainEvent) error {
    // Valida los datos del evento
    if event.Amount <= 0 {
        return errors.New("invalid amount")
    }

    // Guarda el evento en el estado cross-chain
    state, exists := CrossChainStates[event.DestinationChain]
    if !exists {
        state = &CrossChainState{}
        CrossChainStates[event.DestinationChain] = state
    }
    state.Events = append(state.Events, event)

    log.Printf("Cross-chain event emitted: %+v", event)
    return nil
}

func ProcessIncomingCrossChainEvent(event CrossChainEvent) error {
    // Valida el evento recibido
    if event.Amount <= 0 || event.Receiver == "" {
        return errors.New("invalid cross-chain event")
    }

    // Registra el evento en el estado de la blockchain actual
    state, exists := CrossChainStates[event.OriginChain]
    if !exists {
        state = &CrossChainState{}
        CrossChainStates[event.OriginChain] = state
    }
    state.Events = append(state.Events, event)

    log.Printf("Cross-chain event processed: %+v", event)
    return nil
}

func FundCommunity(amount int64) {
    if amount <= 0 || TotalSupply+amount > MaxSupply {
        return
    }
    CommunityFund.Balance += amount
    TotalSupply += amount
    log.Printf("Community fund increased by %d. Total: %d\n", amount, CommunityFund.Balance)
}

func ProcessTransaction(tx *Transaction) {
    // Valida la transacción...
    BurnTransactionFee(tx.Fee)


func BurnTransactionFee(fee int64) {
    TotalSupply -= fee
    if TotalSupply < 0 {
        TotalSupply = 0
    }
    log.Printf("Burned %d tokens. Total supply: %d\n", fee, TotalSupply)
}