package transaction

// Transaction representa una transacción de AncapCoin
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// TXInput representa una entrada de transacción
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// TXOutput representa una salida de transacción
type TXOutput struct {
	Value        int64
	ScriptPubKey string
}

// NewTransaction crea una nueva transacción
func NewTransaction(from, to string, amount int) *Transaction {
	// Implementación simplificada
	return &Transaction{}
}