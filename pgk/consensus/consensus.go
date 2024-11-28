package blockchain

import (
	"errors"
	"math/big"
	"sync"

	"ANCAPCOIN/pkg/transaction"
	"ANCAPCOIN/pkg/util"
)

type ProofOfAuthority struct {
	Authorities []string
}

// Consensus defines the interface for consensus mechanisms
type Consensus interface {
	AddBlock(transactions []*transaction.Transaction, prevHash []byte, height int) (*Block, error)
	ValidateBlock(block *Block, prevBlock *Block) error
}

// ProofOfWork implements the Proof of Work consensus mechanism
type ProofOfWork struct {
	Difficulty int
}

func NewProofOfWork(difficulty int) *ProofOfWork {
	return &ProofOfWork{Difficulty: difficulty}
}

func (pow *ProofOfWork) AddBlock(transactions []*transaction.Transaction, prevHash []byte, height int) (*Block, error) {
	block := NewBlock(transactions, prevHash, height)
	powAlgo := util.NewProofOfWork(block, pow.Difficulty)
	nonce, hash := powAlgo.Run()
	block.Nonce = nonce
	block.Hash = hash
	return block, nil
}

func (pow *ProofOfWork) ValidateBlock(block *Block, prevBlock *Block) error {
	if !bytes.Equal(block.PrevBlockHash, prevBlock.Hash) {
		return errors.New("invalid previous block hash")
	}
	powAlgo := util.NewProofOfWork(block, pow.Difficulty)
	if !powAlgo.Validate() {
		return errors.New("invalid proof of work")
	}
	return nil
}

// ProofOfStake implements the Proof of Stake consensus mechanism
type ProofOfStake struct {
	StakeMap map[string]int64
	mu       sync.RWMutex
}

func NewProofOfStake() *ProofOfStake {
	return &ProofOfStake{
		StakeMap: make(map[string]int64),
	}
}

func (pos *ProofOfStake) AddBlock(transactions []*transaction.Transaction, prevHash []byte, height int) (*Block, error) {
	// Simplified staking logic: select the highest stake holder
	stakeholder := pos.selectValidator()
	block := NewBlock(transactions, prevHash, height)
	block.Hash = util.HashSHA256([]byte(stakeholder))
	return block, nil
}

func (pos *ProofOfStake) ValidateBlock(block *Block, prevBlock *Block) error {
	if !bytes.Equal(block.PrevBlockHash, prevBlock.Hash) {
		return errors.New("invalid previous block hash")
	}
	return nil
}

func (pos *ProofOfStake) selectValidator() string {
	pos.mu.RLock()
	defer pos.mu.RUnlock()

	var maxStake int64
	var stakeholder string
	for addr, stake := range pos.StakeMap {
		if stake > maxStake {
			maxStake = stake
			stakeholder = addr
		}
	}
	return stakeholder
}

func NewProofOfAuthority(authorities []string) *ProofOfAuthority {
	return &ProofOfAuthority{Authorities: authorities}
}

func (poa *ProofOfAuthority) AddBlock(transactions []*transaction.Transaction, prevHash []byte, height int) (*Block, error) {
	// Seleccionar una autoridad basada en un criterio, como la altura del bloque
	selectedAuthority := poa.Authorities[height%len(poa.Authorities)]
	block := NewBlock(transactions, prevHash, height)
	block.Hash = util.HashSHA256([]byte(selectedAuthority))
	return block, nil
}

func (poa *ProofOfAuthority) ValidateBlock(block *Block, prevBlock *Block) error {
	if !bytes.Equal(block.PrevBlockHash, prevBlock.Hash) {
		return errors.New("invalid previous block hash")
	}
	return nil
}

func (pos *ProofOfStake) RegisterValidator(address string, stake int64) {
	pos.mu.Lock()
	defer pos.mu.Unlock()
	pos.StakeMap[address] += stake
}

func (pos *ProofOfStake) selectValidator() string {
	pos.mu.RLock()
	defer pos.mu.RUnlock()
	var maxStake int64
	var selected string
	for addr, stake := range pos.StakeMap {
		if stake > maxStake {
			maxStake = stake
			selected = addr
		}
	}
	return selected
}