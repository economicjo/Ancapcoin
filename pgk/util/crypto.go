package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"golang.org/x/crypto/sha3"
	"golang.org/x/crypto/sha3"
)

// GenerateKeyPair generates a public and private key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Combine X and Y coordinates of the public key
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return private, pubKey, nil
}

// ValidatePublicKey ensures the validity of a public key
func ValidatePublicKey(pubKey []byte) error {
	if len(pubKey) != 64 {
		return errors.New("invalid public key length")
	}

	x := new(big.Int).SetBytes(pubKey[:len(pubKey)/2])
	y := new(big.Int).SetBytes(pubKey[len(pubKey)/2:])
	curve := elliptic.P256()

	if !curve.IsOnCurve(x, y) {
		return errors.New("invalid public key: not on curve")
	}
	return nil
}

// PublicKeyHex converts a public key to a hexadecimal string
func PublicKeyHex(pubKey []byte) string {
	return hex.EncodeToString(pubKey)
}

// SignData signs the data using the private key
func SignData(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// Serialize the signature as R||S
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature verifies the signature of the data
func VerifySignature(data, signature, pubKey []byte) (bool, error) {
	if err := ValidatePublicKey(pubKey); err != nil {
		return false, err
	}

	if len(signature) != 64 {
		return false, errors.New("invalid signature length")
	}

	hash := sha256.Sum256(data)
	curve := elliptic.P256()

	// Extract X, Y coordinates from the public key
	x := new(big.Int).SetBytes(pubKey[:len(pubKey)/2])
	y := new(big.Int).SetBytes(pubKey[len(pubKey)/2:])
	publicKey := ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	// Extract R, S from the signature
	r := new(big.Int).SetBytes(signature[:len(signature)/2])
	s := new(big.Int).SetBytes(signature[len(signature)/2:])

	// Verify the signature
	return ecdsa.Verify(&publicKey, hash[:], r, s), nil
}

// ProofOfWork represents the Proof of Work algorithm
type ProofOfWork struct {
    block    *Block
    target   *big.Int
    hashAlgo string // "sha256", "sha3", "blake3"
}

func (pow *ProofOfWork) Run(prepareDataFunc func(int) []byte) (int, []byte) {
	var hashInt big.Int
	var hash []byte
	nonce := 0

	for nonce < (1 << 31) {
		data := prepareDataFunc(nonce)

		// Select hash algorithm
		switch pow.hashAlgo {
		case "sha256":
			hash = HashSHA256(data)
		case "sha3":
			hash = HashSHA3(data)
		case "blake3":
			hash = HashBLAKE3(data)
		default:
			panic("invalid hash algorithm")
		}

		hashInt.SetBytes(hash)
		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}
	return nonce, hash
}

// NewProofOfWork creates a new Proof of Work instance
func NewProofOfWork(block interface{}, difficulty int, hashAlgo string) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty)) // Adjust difficulty

	// Validate hash algorithm
	validHashAlgos := map[string]bool{"sha256": true, "sha3": true, "blake3": true}
	if !validHashAlgos[hashAlgo] {
		panic("invalid hash algorithm")
	}

	return &ProofOfWork{block: block, target: target, hashAlgo: hashAlgo}
}

// Run executes the Proof of Work algorithm
func (pow *ProofOfWork) Run(prepareDataFunc func(int) []byte) (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for nonce < (1 << 31) {
		data := prepareDataFunc(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}
	return nonce, hash[:]
}

// Validate checks the validity of a Proof of Work
func (pow *ProofOfWork) Validate() bool {
	data := pow.PrepareData()
	hash := sha256.Sum256(data)
	var hashInt big.Int
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}

// PrepareData prepares the data for the Proof of Work algorithm
func PrepareData(block interface{}, nonce int, customFields [][]byte) []byte {
	return bytes.Join(
		append(customFields, IntToHex(int64(nonce))),
		[]byte{},
	)
}

// IntToHex converts an integer to a hexadecimal representation
func IntToHex(num int64) []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

// HashData creates a SHA256 hash from a byte slice
func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (pow *ProofOfWork) Validate(prepareDataFunc func(int) []byte, nonce int) bool {
	data := prepareDataFunc(nonce)
	var hash []byte

	// Select hash algorithm
	switch pow.hashAlgo {
	case "sha256":
		hash = HashSHA256(data)
	case "sha3":
		hash = HashSHA3(data)
	case "blake3":
		hash = HashBLAKE3(data)
	default:
		panic("invalid hash algorithm")
	}

	var hashInt big.Int
	hashInt.SetBytes(hash)

	return hashInt.Cmp(pow.target) == -1
}

func HashSHA3(data []byte) []byte {
    hash := sha3.Sum256(data)
    return hash[:]
}

func HashBLAKE3(data []byte) []byte {
	hash := blake3.Sum256(data)
	return hash[:]
}