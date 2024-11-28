package wallet

import (
	"example.com/ancapcoin/pkg/util"
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"os"

	"golang.org/x/crypto/ripemd160"
)

// Wallet stores the private and public keys
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

type WalletManager struct {
	Wallets map[string]*Wallet // Multiple wallets by address
}

// CreateWallet creates a new wallet and returns the address
func (wm *WalletManager) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.GetAddress()
	wm.Wallets[address] = wallet
	return address
}

// NewWallet creates and returns a new wallet
func NewWallet() (*Wallet, error) {
	privateKey, publicKey, err := util.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	return &Wallet{PrivateKey: privateKey, PublicKey: publicKey}, nil
}

// PublicKeyHex returns the public key in hexadecimal format
func (w *Wallet) PublicKeyHex() string {
	return util.PublicKeyHex(w.PublicKey)
}

// GetAddress generates and returns the wallet address
func (w *Wallet) GetAddress() (string, error) {
	pubKeyHash := hashPublicKey(w.PublicKey)
	version := byte(0x00) // Address version (0x00 for mainnet)
	payload := append([]byte{version}, pubKeyHash...)
	checksum := calculateChecksum(payload)
	fullPayload := append(payload, checksum...)
	return hex.EncodeToString(fullPayload), nil
}

// ValidateAddress verifies if an address is valid
func ValidateAddress(address string) (bool, error) {
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return false, err
	}
	if len(addressBytes) < 25 {
		return false, errors.New("invalid address length")
	}

	payload := addressBytes[:len(addressBytes)-4]
	checksum := addressBytes[len(addressBytes)-4:]
	expectedChecksum := calculateChecksum(payload)

	return bytes.Equal(checksum, expectedChecksum), nil
}

// BackupWallet saves the wallet to a file
func (w *Wallet) BackupWallet(filename string) error {
	data, err := w.Serialize()
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0600)
	if err != nil {
		return errors.New("failed to backup wallet")
	}

	return nil
}

// RestoreWallet loads a wallet from a file
func RestoreWallet(filename string) (*Wallet, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.New("failed to restore wallet")
	}

	wallet, err := Deserialize(data)
	if err != nil {
		return nil, errors.New("failed to deserialize wallet")
	}

	return wallet, nil
}

// hashPublicKey performs SHA256 followed by RIPEMD160 on the public key
func hashPublicKey(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)
	ripeHasher := ripemd160.New()
	_, err := ripeHasher.Write(pubHash[:])
	if err != nil {
		panic(err) // Consider better error handling here
	}
	return ripeHasher.Sum(nil)
}

// calculateChecksum computes the last 4 bytes of the double SHA256 hash
func calculateChecksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:4]
}

// Serialize converts a wallet to bytes for storage
func (w *Wallet) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(w)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Deserialize converts bytes into a wallet
func Deserialize(data []byte) (*Wallet, error) {
	var wallet Wallet
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// SignMessage signs a message using the wallet's private key
func (w *Wallet) SignMessage(message []byte) ([]byte, error) {
	return util.SignData(message, w.PrivateKey)
}

// VerifySignature verifies the signature of a given message
func VerifySignature(message, signature, publicKey []byte) (bool, error) {
	return util.VerifySignature(message, signature, publicKey)
}

// ExportPublicKey exports the public key in a format suitable for sharing
func (w *Wallet) ExportPublicKey() string {
	return util.PublicKeyHex(w.PublicKey)
}

// ExportPrivateKey exports the private key in hexadecimal format (optional, secure only for backups)
func (w *Wallet) ExportPrivateKey() (string, error) {
	if w.PrivateKey == nil {
		return "", errors.New("private key not available")
	}

	privateKeyBytes := w.PrivateKey.D.Bytes()
	return hex.EncodeToString(privateKeyBytes), nil
}
