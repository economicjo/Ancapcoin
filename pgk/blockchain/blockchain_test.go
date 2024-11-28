package blockchain

import (
	"bytes"
	"testing"
)

// TestBlockchainInitialization validates the initialization of a new blockchain.
func TestBlockchainInitialization(t *testing.T) {
	bc := NewBlockchain("test-address")
	defer bc.Close()

	if bc == nil {
		t.Fatal("Failed to initialize blockchain")
	}

	// Verify that the blockchain contains the genesis block only.
	iter := bc.Iterator()
	block, err := iter.Next()
	if err != nil {
		t.Fatalf("Error iterating blockchain: %v", err)
	}

	if len(block.PrevBlockHash) != 0 {
		t.Fatal("Genesis block should not have a previous hash")
	}

	if block.Transactions == nil || len(block.Transactions) == 0 {
		t.Fatal("Genesis block should contain at least one transaction")
	}

	if !block.Transactions[0].IsCoinbase() {
		t.Fatal("Genesis block should contain a coinbase transaction")
	}
}

// TestAddingBlocks validates that new blocks are added correctly to the blockchain.
func TestAddingBlocks(t *testing.T) {
	bc := NewBlockchain("test-address")
	defer bc.Close()

	// Add a new block with a transaction.
	testTransaction := []*Transaction{NewCoinbaseTransaction("test-reward", "Test Block")}
	err := bc.AddBlock(testTransaction)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	// Verify the block was added correctly.
	iter := bc.Iterator()
	newBlock, err := iter.Next() // Most recent block
	if err != nil {
		t.Fatalf("Error iterating blockchain: %v", err)
	}

	genesisBlock, err := iter.Next()
	if err != nil {
		t.Fatalf("Error iterating blockchain: %v", err)
	}

	if !bytes.Equal(newBlock.PrevBlockHash, genesisBlock.Hash) {
		t.Fatal("New block's previous hash does not match the genesis block's hash")
	}

	if len(newBlock.Transactions) != 1 {
		t.Fatalf("Expected 1 transaction in new block, found %d", len(newBlock.Transactions))
	}

	if !newBlock.Transactions[0].IsCoinbase() {
		t.Fatal("New block should contain a coinbase transaction")
	}
}

// TestBlockSerialization validates block serialization and deserialization.
func TestBlockSerialization(t *testing.T) {
	transactions := []*Transaction{NewCoinbaseTransaction("test-address", "Test Serialization")}
	block := NewBlock(transactions, []byte("prev-hash"))

	serialized := block.Serialize()
	deserialized := DeserializeBlock(serialized)

	if !bytes.Equal(block.Hash, deserialized.Hash) {
		t.Fatal("Deserialized block hash does not match the original hash")
	}

	if !bytes.Equal(block.PrevBlockHash, deserialized.PrevBlockHash) {
		t.Fatal("Deserialized block's previous hash does not match the original")
	}

	if block.Timestamp != deserialized.Timestamp {
		t.Fatal("Deserialized block's timestamp does not match the original")
	}
}

// TestBlockchainIteration validates the blockchain iterator functionality.
func TestBlockchainIteration(t *testing.T) {
	bc := NewBlockchain("test-address")
	defer bc.Close()

	// Add several blocks.
	for i := 0; i < 3; i++ {
		tx := []*Transaction{NewCoinbaseTransaction("test-reward", "Block " + string(i))}
		err := bc.AddBlock(tx)
		if err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	iter := bc.Iterator()
	count := 0
	for {
		block, err := iter.Next()
		if err != nil {
			break
		}
		count++
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	if count != 4 { // 3 added blocks + 1 genesis block
		t.Fatalf("Blockchain should have 4 blocks, but found %d", count)
	}
}

// TestGenesisBlock validates the correct creation of the genesis block.
func TestGenesisBlock(t *testing.T) {
	genesisTx := NewCoinbaseTransaction("test-address", "Genesis Test")
	genesisBlock := NewGenesisBlock(genesisTx)

	if len(genesisBlock.PrevBlockHash) != 0 {
		t.Fatal("Genesis block should not have a previous hash")
	}

	if len(genesisBlock.Transactions) != 1 {
		t.Fatal("Genesis block should contain exactly one transaction")
	}

	if !genesisBlock.Transactions[0].IsCoinbase() {
		t.Fatal("Genesis block's transaction should be a coinbase transaction")
	}
}

// TestFindTransaction ensures transactions can be retrieved by ID.
func TestFindTransaction(t *testing.T) {
	bc := NewBlockchain("test-address")
	defer bc.Close()

	// Add a block containing a transaction.
	tx := NewCoinbaseTransaction("recipient", "Test Find Transaction")
	err := bc.AddBlock([]*Transaction{tx})
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	// Attempt to find the transaction by ID.
	foundTx, err := bc.FindTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Transaction not found: %v", err)
	}

	if !bytes.Equal(tx.ID, foundTx.ID) {
		t.Fatal("Found transaction does not match the original")
	}
}

// TestBlockchainValidation validates the integrity of the blockchain.
func TestBlockchainValidation(t *testing.T) {
	bc := NewBlockchain("test-address")
	defer bc.Close()

	// Add valid blocks.
	for i := 0; i < 2; i++ {
		tx := []*Transaction{NewCoinbaseTransaction("test-reward", "Block Validation " + string(i))}
		err := bc.AddBlock(tx)
		if err != nil {
			t.Fatalf("Failed to add block %d: %v", i, err)
		}
	}

	// Validate the blockchain.
	err := bc.Validate()
	if err != nil {
		t.Fatalf("Blockchain validation failed: %v", err)
	}
}
