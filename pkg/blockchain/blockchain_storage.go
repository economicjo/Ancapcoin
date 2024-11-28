package blockchain

import (
	"bytes"
	"errors"
	"log"

	"github.com/boltdb/bolt"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
	lastHashKey  = "l" // Key to store the last block's hash
)

// Blockchain represents the blockchain stored in a BoltDB database
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type TransactionPool struct {
	Transactions map[string]*Transaction
}

// Errors
var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrBlockNotFound       = errors.New("block not found")
)

// NewBlockchain initializes a new blockchain with a genesis block
func NewBlockchain(address string) *Blockchain {
	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			// Calcular la recompensa para el bloque génesis
			genesisReward := CalculateBlockReward(0) // Asume altura 0 para el bloque génesis
			genesis := NewGenesisBlock(NewCoinbaseTransaction(address, genesisReward, "Genesis Block"))

			// Crear el bucket para los bloques
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}

			// Guardar el bloque génesis
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}

			// Guardar el hash del último bloque
			err = b.Put([]byte(lastHashKey), genesis.Hash)
			if err != nil {
				return err
			}

			tip = genesis.Hash
		} else {
			tip = b.Get([]byte(lastHashKey))
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip: tip, db: db}
}

// AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(transactions []*transaction.Transaction) error {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte(lastHashKey))
		return nil
	})
	if err != nil {
		return err
	}

	newBlock, err := bc.consensus.AddBlock(transactions, lastHash, bc.GetBlockchainHeight()+1)
	if err != nil {
		return err
	}

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte(lastHashKey), newBlock.Hash)
		if err != nil {
			return err
		}
		bc.tip = newBlock.Hash
		return nil
	})
	return err
}

	// Crear y minar el nuevo bloque
	newBlock := NewBlock(transactions, lastHash)

	// Validar el PoW del nuevo bloque
	pow := util.NewProofOfWork(newBlock, difficulty)
	if !pow.Validate() {
		return errors.New("proof of work validation failed")
	}

	// Guardar el bloque en la base de datos
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte(lastHashKey), newBlock.Hash)
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash
		return nil
	})
	return err
}

// GetBlock retrieves a block by its hash
func (bc *Blockchain) GetBlock(hash []byte) (*Block, error) {
	var block *Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(hash)
		if encodedBlock == nil {
			return ErrBlockNotFound
		}

		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return block, nil
}

// Iterator returns a BlockchainIterator to traverse the blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{currentHash: bc.tip, db: bc.db}
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Next returns the next block in the iteration
func (it *BlockchainIterator) Next() (*Block, error) {
	var block *Block

	err := it.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(it.currentHash)
		if encodedBlock == nil {
			return ErrBlockNotFound
		}

		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		return nil, err
	}

	it.currentHash = block.PrevBlockHash
	return block, nil
}

// Close safely closes the blockchain database
func (bc *Blockchain) Close() error {
	return bc.db.Close()
}

// FindTransaction searches for a transaction by its ID in the blockchain
func (bc *Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	it := bc.Iterator()

	for {
		block, err := it.Next()
		if err == ErrBlockNotFound {
			break
		}

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil, ErrTransactionNotFound
}

// Validate ensures the integrity of the entire blockchain
func (bc *Blockchain) Validate() error {
	it := bc.Iterator()
	var prevBlock *Block

	for {
		block, err := it.Next()
		if err == ErrBlockNotFound {
			break
		} else if err != nil {
			return err
		}

		// Validar cada transacción dentro del bloque
		for _, tx := range block.Transactions {
			if !tx.Verify(bc) {
				log.Printf("Invalid transaction in block: %x", block.Hash)
				return errors.New("invalid transaction found in block")
			}
		}

		// Validar el PoW
		pow := util.NewProofOfWork(block)
		if !pow.Validate() {
			log.Printf("Invalid PoW for block: %x", block.Hash)
			return errors.New("invalid proof of work")
		}

		// Validar hash previo
		if prevBlock != nil && !bytes.Equal(block.PrevBlockHash, prevBlock.Hash) {
			log.Printf("Invalid previous hash for block: %x", block.Hash)
			return errors.New("invalid block hash chain")
		}

		prevBlock = block
	}
	return nil
}


func (bc *Blockchain) AddBlock(transactions []*Transaction) error {
	// Validar que todas las transacciones sean válidas
	for _, tx := range transactions {
		if !tx.Verify(bc) { // Verifica las firmas y los UTXOs
			return errors.New("invalid transaction detected")
		}
	}

	// Obtener el hash del bloque anterior
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte(lastHashKey))
		return nil
	})
	if err != nil {
		return err
	}

	// Crear y minar el nuevo bloque
	newBlock := NewBlock(transactions, lastHash)

	// Validar el PoW del nuevo bloque
	pow := util.NewProofOfWork(newBlock, difficulty)
	if !pow.Validate() {
		return errors.New("proof of work validation failed")
	}

	// Guardar el bloque en la base de datos
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte(lastHashKey), newBlock.Hash)
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash
		return nil
	})
	return err
}

func (bc *Blockchain) Validate() error {
	it := bc.Iterator()
	var prevBlock *Block

	for {
		block, err := it.Next()
		if err == ErrBlockNotFound {
			break
		} else if err != nil {
			return err
		}

		// Validar cada transacción dentro del bloque
		for _, tx := range block.Transactions {
			if !tx.Verify(bc) {
				return errors.New("invalid transaction found in block")
			}
		}

		// Validar el PoW
		pow := util.NewProofOfWork(block)
		if !pow.Validate() {
			return errors.New("invalid proof of work")
		}

		// Validar hash previo
		if prevBlock != nil && !bytes.Equal(block.PrevBlockHash, prevBlock.Hash) {
			return errors.New("invalid block hash chain")
		}

		prevBlock = block
	}
	return nil
}

func (bc *Blockchain) GetHeaders(startHeight int) []BlockHeader {
	var headers []BlockHeader
	it := bc.Iterator()

	for {
		block, err := it.Next()
		if err == ErrBlockNotFound {
			break
		}

		if block.Height >= startHeight {
			headers = append(headers, BlockHeader{
				Hash:          block.Hash,
				PrevBlockHash: block.PrevBlockHash,
				Timestamp:     block.Timestamp,
			})
		}
	}

	return headers
}

// AddTransaction adds a transaction to the pool
func (tp *TransactionPool) RemoveTransactions(transactions []*Transaction) {
	for _, tx := range transactions {
		delete(tp.Transactions, string(tx.ID))
	}
}

// GetTransactions retrieves all valid transactions for a new block
func (tp *TransactionPool) GetTransactions() []*Transaction {
	var transactions []*Transaction
	for _, tx := range tp.Transactions {
		transactions = append(transactions, tx)
	}
	return transactions

	func (bc *Blockchain) GetBlockchainHeight() int {
		var height int
		err := bc.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			blockData := b.Get(bc.tip)
			block := DeserializeBlock(blockData)
			height = block.Height
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		return height
	}