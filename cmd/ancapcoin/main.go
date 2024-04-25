package main

import (
	"ANCAPCOIN/pkg/blockchain"
	"ANCAPCOIN/pkg/wallet"
	"fmt"
)

func main() {
	// Inicialización del monedero
	w := wallet.NewWallet()
	fmt.Printf("Dirección pública del monedero: %s\n", w.PublicKeyHex())

	// Inicialización de la blockchain
	bc := blockchain.NewBlockchain(w.PublicKeyHex())

	// Añadir bloques de prueba a la blockchain (minado)
	bc.AddBlock("Primer bloque")
	bc.AddBlock("Segundo bloque")

	// Imprimir el contenido de la blockchain
	for _, block := range bc.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Datos: %s\n", block.Data)
		fmt.Printf("Hash: %x\n\n", block.Hash)
	}
}