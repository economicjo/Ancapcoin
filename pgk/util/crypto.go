package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

// NewKeyPair genera un par de claves pública y privada
func NewKeyPair() ([]byte, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return private.D.Bytes(), pubKey
}

// PubKeyHex convierte una clave pública a formato hexadecimal
func PubKeyHex(pubKey []byte) string {
	return hex.EncodeToString(pubKey)
}

// NewProofOfWork simula la creación de un "Proof of Work"
func NewProofOfWork(block *Block) *ProofOfWork {
	return &ProofOfWork{block, 256}
}

// ProofOfWork representa la estructura del algoritmo de prueba de trabajo
type ProofOfWork struct {
	block  *Block
	target int
}

// Run simula la ejecución del algoritmo de prueba de trabajo
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for nonce < (1<<31) {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(big.NewInt(int64(pow.target))) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce, hash[:]
}

// prepareData prepara los datos para el hash
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(pow.target)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}