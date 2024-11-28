package token

type Token struct {
	Name        string  // Nombre del token
	Symbol      string  // Símbolo del token (e.g., ANC)
	TotalSupply int64   // Suministro total
	Owner       string  // Dirección del creador del token
	Balances    map[string]int64 // Saldo de cada dirección
}

// Crear un nuevo token
func NewToken(name, symbol string, totalSupply int64, owner string) *Token {
	return &Token{
		Name:        name,
		Symbol:      symbol,
		TotalSupply: totalSupply,
		Owner:       owner,
		Balances:    map[string]int64{owner: totalSupply}, // El creador posee inicialmente todos los tokens
	}
}

// Transferir tokens entre direcciones
func (t *Token) Transfer(from, to string, amount int64) error {
	if t.Balances[from] < amount {
		return fmt.Errorf("insufficient balance")
	}
	t.Balances[from] -= amount
	t.Balances[to] += amount
	return nil
}