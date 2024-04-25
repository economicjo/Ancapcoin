package wallet

import (
	"ANCAPCOIN/pkg/util"
)

// Wallet almacena las claves privadas y públicas
type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

// NewWallet crea y retorna un nuevo monedero
func NewWallet() *Wallet {
	private, public := util.NewKeyPair()
	return &Wallet{private, public}
}

// PublicKeyHex retorna la clave pública en formato hexadecimal
func (w *Wallet) PublicKeyHex() string {
	return util.PubKeyHex(w.PublicKey)
}