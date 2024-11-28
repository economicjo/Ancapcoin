package lightning

import (
	"sync"
)

type LightningChannel struct {
	Participants [2]string  // Direcciones de los participantes
	Balance      [2]int64   // Saldos de los participantes
	mu           sync.Mutex // Mutex para concurrencia
}

// NewLightningChannel crea un canal de pago entre dos participantes
func NewLightningChannel(participant1, participant2 string, balance1, balance2 int64) *LightningChannel {
	return &LightningChannel{
		Participants: [2]string{participant1, participant2},
		Balance:      [2]int64{balance1, balance2},
	}
}

// SendPayment realiza un pago de un participante a otro
func (lc *LightningChannel) SendPayment(from string, amount int64) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var sender, receiver int
	if lc.Participants[0] == from {
		sender = 0
		receiver = 1
	} else if lc.Participants[1] == from {
		sender = 1
		receiver = 0
	} else {
		return fmt.Errorf("sender not found in the channel")
	}

	if lc.Balance[sender] < amount {
		return fmt.Errorf("insufficient balance")
	}

	lc.Balance[sender] -= amount
	lc.Balance[receiver] += amount
	return nil
}
