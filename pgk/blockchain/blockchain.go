package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

const genesisData = "19/11/2023 the day we reclaimed freedom" // Nueva frase en el bloque génesis.
const subsidy = 100 // La recompensa inicial de bloques en AncapCoin es ahora 100.
const halvingInterval = 210000 // Intervalo de bloques para el halving sigue igual.

// Block representa cada bloque en la blockchain.
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock crea y retorna un bloque vinculado al bloque anterior.
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// SetGenesisBlock crea el bloque génesis.
func SetGenesisBlock() *Block {
	return NewBlock(genesisData, []byte{})
}

// Blockchain mantiene una secuencia de bloques.
type Blockchain struct {
	blocks []*Block
}

// AddBlock guarda el bloque en la blockchain.
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewBlockchain crea una nueva blockchain con un bloque génesis.
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{SetGenesisBlock()}}
}

// CalculateBlockReward calcula la recompensa por bloque basado en el punto de tiempo en la cadena.
func (bc *Blockchain) CalculateBlockReward() int {
	currentHeight := len(bc.blocks)
	reward := subsidy
	halvings := currentHeight / halvingInterval

	// Realiza el halving de la recompensa.
	for i := 0; i < halvings; i++ {
		reward /= 2

		// Detiene el halving si la recompensa llega a 0 o si se alcanzan 42 millones de unidades.
		if reward == 0 || (currentHeight*reward >= 42000000) {
			break
		}
	}

	return reward
}

// Serialize convierte el bloque en bytes.
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock convierte bytes a Block.
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}