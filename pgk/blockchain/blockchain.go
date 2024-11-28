package blockchain

import (
	"ANCAPCOIN/pkg/transaction"
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
	Subsidy         = 100
	HalvingInterval = 210000
)

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

// CalculateBlockReward computes the mining reward based on the height
func CalculateBlockReward(height int) int {
	reward := Subsidy
	halvings := height / HalvingInterval

	for i := 0; i < halvings; i++ {
		reward /= 2
		if reward == 0 {
			break
		}
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