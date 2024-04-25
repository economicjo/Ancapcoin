package blockchain

import (
	"ANCAPCOIN/pkg/util"
	"time"
)

// Block representa cada elemento de la cadena de bloques
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock crea y retorna un nuevo bloque
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := util.NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock crea y retorna el bloque génesis
func NewGenesisBlock(address string) *Block {
	return NewBlock("Bloque Génesis", []byte{})
}