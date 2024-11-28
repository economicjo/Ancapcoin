package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"log"

	"ANCAPCOIN/pkg/util"
)

// Transaction represents a transaction in AncapCoin
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type ContractTransaction struct {
	ContractAddress string        
	Function        string        
	Args            []interface{} 
}

// TXInput represents an input in a transaction
type TXInput struct {
	Txid      []byte // Transaction ID where the UTXO is located
	Vout      int    // Index of the output in the transaction
	Signature []byte // Digital signature for validation
	PubKey    []byte // Public key of the sender
}

// TXOutput represents an output in a transaction
type TXOutput struct {
	Value        int64  // Amount of the output
	ScriptPubKey string // Locking script for the output
}

type SegWit struct {
	Witness []byte // Separate signature data
}

// Modified Transaction to include SegWit
type Transaction struct {
	ID        []byte
	Vin       []TXInput
	Vout      []TXOutput
	SegWit    SegWit
}

type SegWitData struct {
    Witness []byte // Datos segregados para las firmas
}

type Transaction struct {
    ID       []byte
    Vin      []TXInput
    Vout     []TXOutput
    SegWit   *SegWitData // Agregar soporte para SegWit
}

// NewCoinbaseTransaction creates a new Coinbase transaction
func NewCoinbaseTransaction(to, data string) *Transaction {
	if data == "" {
		data = "Coinbase reward"
	}

	txin := TXInput{Txid: []byte{}, Vout: -1, Signature: nil, PubKey: []byte(data)}
	txout := TXOutput{Value: 50, ScriptPubKey: to}
	tx := Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}
	tx.ID = tx.Hash()
	return &tx
}

// Hash calculates the hash of the transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(buffer.Bytes())
	return hash[:]
}

// Sign signs each input of a transaction using the sender's private key
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]*Transaction) error {
	if tx.IsCoinbase() {
		return nil // Coinbase transactions don't need to be signed
	}

	for _, vin := range tx.Vin {
		if prevTXs[string(vin.Txid)] == nil {
			return errors.New("previous transaction is missing")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inIdx, vin := range txCopy.Vin {
		prevTX := prevTXs[string(vin.Txid)]
		txCopy.Vin[inIdx].Signature = nil
		txCopy.Vin[inIdx].PubKey = prevTX.Vout[vin.Vout].ScriptPubKey

		dataToSign := txCopy.Hash()
		signature, err := util.SignData(dataToSign, privateKey)
		if err != nil {
			return err
		}

		tx.Vin[inIdx].Signature = signature
		txCopy.Vin[inIdx].PubKey = nil
	}

	return nil
}

// Verify verifies the signatures of transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]*Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[string(vin.Txid)] == nil {
			log.Panic("previous transaction is missing")
		}
	}

	txCopy := tx.TrimmedCopy()
	for inIdx, vin := range tx.Vin {
		prevTX := prevTXs[string(vin.Txid)]
		txCopy.Vin[inIdx].Signature = nil
		txCopy.Vin[inIdx].PubKey = prevTX.Vout[vin.Vout].ScriptPubKey

		dataToVerify := txCopy.Hash()
		if !util.VerifySignature(dataToVerify, vin.Signature, vin.PubKey) {
			return false
		}
		txCopy.Vin[inIdx].PubKey = nil
	}

	return true
}

// TrimmedCopy creates a trimmed version of the transaction for signing
func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{Txid: vin.Txid, Vout: vin.Vout, Signature: nil, PubKey: nil})
	}

	return &Transaction{ID: tx.ID, Vin: inputs, Vout: tx.Vout}
}

// Serialize serializes a transaction into a byte slice
func (tx *Transaction) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

// DeserializeTransaction deserializes a byte slice into a transaction
func DeserializeTransaction(data []byte) *Transaction {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&tx)
	if err != nil {
		log.Panic(err)
	}

	return &tx
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount int64, utxoSet map[string][]TXOutput, privateKey *ecdsa.PrivateKey) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	// Gather UTXOs to cover the amount
	accumulated, validOutputs := FindSpendableOutputs(from, amount, utxoSet)

	if accumulated < amount {
		return nil, errors.New("insufficient funds")
	}

	// Create inputs
	for txid, outs := range validOutputs {
		txIDBytes := []byte(txid)
		for _, outIdx := range outs {
			input := TXInput{Txid: txIDBytes, Vout: outIdx, Signature: nil, PubKey: []byte(from)}
			inputs = append(inputs, input)
		}
	}

	// Create outputs
	outputs = append(outputs, TXOutput{Value: amount, ScriptPubKey: to})
	if accumulated > amount {
		outputs = append(outputs, TXOutput{Value: accumulated - amount, ScriptPubKey: from}) // Change
	}

	tx := Transaction{ID: nil, Vin: inputs, Vout: outputs}
	tx.ID = tx.Hash()

	// Sign the transaction
	if err := tx.Sign(privateKey, map[string]*Transaction{}); err != nil {
		return nil, err
	}

	return &tx, nil
}

func FindSpendableOutputs(address string, amount int64, utxoSet map[string][]TXOutput) (int64, map[string][]int) {
	var accumulated int64
	validOutputs := make(map[string][]int)

	for txid, outputs := range utxoSet {
		for idx, out := range outputs {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				validOutputs[txid] = append(validOutputs[txid], idx)

				if accumulated >= amount {
					break
				}
			}
		}
	}

	return accumulated, validOutputs
}

// CanBeUnlockedWith verifica si el ScriptPubKey puede ser desbloqueado con los datos proporcionados
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}