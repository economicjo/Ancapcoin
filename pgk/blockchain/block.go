package blockchain

import (
	"ANCAPCOIN/pkg/transaction"
	"ANCAPCOIN/pkg/util"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"log"
	"time"
)

type MerkleProof struct {
	Proof [][]byte
	Index int
}

// Block represents each block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*transaction.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int // Block height for traceability
}

// NewBlock creates and returns a new block with transactions
func NewBlock(transactions []*transaction.Transaction, prevBlockHash []byte, height int) *Block {
	var validTransactions []*transaction.Transaction
	for _, tx := range transactions {
		if tx.Validate() {
			validTransactions = append(validTransactions, tx)
		} else {
			log.Printf("Invalid transaction excluded: %x", tx.ID)
		}
	}

	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  validTransactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Height:        height,
	}

	pow := util.NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// calculateMerkleRoot generates a Merkle root from a slice of hashes
func calculateMerkleRoot(txHashes [][]byte) []byte {
	if len(txHashes) == 0 {
		return nil
	}

	// Duplicate the last hash if the number of hashes is odd
	if len(txHashes)%2 != 0 {
		txHashes = append(txHashes, txHashes[len(txHashes)-1])
	}

	// Combine hashes iteratively
	for len(txHashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(txHashes); i += 2 {
			combined := append(txHashes[i], txHashes[i+1]...)
			newHash := sha256.Sum256(combined)
			newLevel = append(newLevel, newHash[:])
		}
		txHashes = newLevel
	}

	return txHashes[0]
}

// NewGenesisBlock creates and returns the genesis block with a Coinbase transaction
func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{}, 0)
}

// HashTransactions calculates a Merkle root-like hash of all transactions in the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	// Include transaction hashes
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	// Include SegWit witness data if available
	for _, tx := range b.Transactions {
		if tx.SegWit != nil && len(tx.SegWit.Witness) > 0 {
			txHashes = append(txHashes, tx.SegWit.Witness)
		}
	}

	// Use a Merkle Tree if needed
	if len(txHashes) > 1 {
		return calculateMerkleRoot(txHashes)
	}

	// If there's only one hash, return it directly
	return txHashes[0]
}

// Serialize converts a block to a byte slice
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock converts a byte slice into a Block
func DeserializeBlock(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

// Validate checks if a block is valid
func (b *Block) Validate(prevBlock *Block) error {
	if !bytes.Equal(b.PrevBlockHash, prevBlock.Hash) {
		return fmt.Errorf("invalid previous block hash: got %x, expected %x", b.PrevBlockHash, prevBlock.Hash)
	}

	pow := util.NewProofOfWork(b)
	if !pow.Validate() {
		return errors.New("invalid proof of work")
	}

	for _, tx := range b.Transactions {
		if !tx.Validate() {
			return fmt.Errorf("invalid transaction detected: %x", tx.ID)
		}
	}

	if b.ContainsDuplicateTransactions() {
		return errors.New("block contains duplicate transactions")
	}

	return nil
}
// GetTransactionByID retrieves a transaction by its ID
func (b *Block) GetTransactionByID(txID []byte) (*transaction.Transaction, bool) {
	for _, tx := range b.Transactions {
		if bytes.Equal(tx.ID, txID) {
			return tx, true
		}
	}
	return nil, false
}

// TotalBlockValue calculates the total value of transactions in the block
func (b *Block) TotalBlockValue() int64 {
	var total int64
	for _, tx := range b.Transactions {
		for _, out := range tx.Vout {
			total += out.Value
		}
	}
	return total
}

// ContainsDuplicateTransactions checks for duplicate transactions in a block
func (b *Block) ContainsDuplicateTransactions() bool {
	seen := make(map[string]bool)
	for _, tx := range b.Transactions {
		txID := string(tx.ID)
		if seen[txID] {
			return true
		}
		seen[txID] = true
	}
	return false
}

// LogBlockDetails logs detailed information about the block for auditing purposes
func (b *Block) LogBlockDetails() {
	log.Printf("Block Height: %d", b.Height)
	log.Printf("Timestamp: %d", b.Timestamp)
	log.Printf("Previous Hash: %x", b.PrevBlockHash)
	log.Printf("Hash: %x", b.Hash)
	log.Printf("Number of Transactions: %d", len(b.Transactions))
	log.Printf("Total Value: %d", b.TotalBlockValue())
}

func NewBlockFromPool(pool *TransactionPool, prevBlockHash []byte, height int) *Block {
	transactions := pool.GetTransactions()
	if len(validTransactions) == 0 {
		log.Println("No valid transactions to include in the block")
		return nil
	}
	return NewBlock(transactions, prevBlockHash, height)
}

func (b *Block) HasSegWit() bool {
	for _, tx := range b.Transactions {
		if len(tx.SegWit.Witness) > 0 {
			return true
		}
	}
	return false
}

func (b *Block) GenerateMerkleProof(txID []byte) *MerkleProof {
	var proof [][]byte
	var index int

	txHashes := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		txHashes[i] = tx.ID
		if bytes.Equal(tx.ID, txID) {
			index = i
		}
	}

	for len(txHashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(txHashes); i += 2 {
			combined := append(txHashes[i], txHashes[i+1]...)
			hash := sha256.Sum256(combined)
			newLevel = append(newLevel, hash[:])
			if i == index || i+1 == index {
				proof = append(proof, hash[:])
			}
		}
		txHashes = newLevel
	}
	return &MerkleProof{Proof: proof, Index: index}
}

func (proof *MerkleProof) Validate() bool {
	hash := proof.Proof[0]
	for i := 1; i < len(proof.Proof); i++ {
		hash = sha256.Sum256(append(hash, proof.Proof[i]...))[:]
	}
	return bytes.Equal(hash, proof.Proof[len(proof.Proof)-1])
}