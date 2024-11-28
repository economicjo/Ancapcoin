package explorer

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"example.com/ancapcoin/pkg/blockchain"
)

// BlockData represents a block's information to be displayed in the explorer
type BlockData struct {
	Height        int
	Timestamp     string
	Transactions  int
	PrevBlockHash string
	Hash          string
	Nonce         int
}

// Explorer holds the blockchain and provides methods to start the web interface
type Explorer struct {
	Blockchain *blockchain.Blockchain
	templates  *template.Template
	mu         sync.RWMutex
	blockCache map[string]*blockchain.Block // Cache for fast block lookup
}

// NewExplorer initializes the blockchain explorer
func NewExplorer(bc *blockchain.Blockchain) *Explorer {
	return &Explorer{
		Blockchain: bc,
		templates:  loadTemplates(),
		blockCache: buildBlockCache(bc),
	}
}

// loadTemplates loads and parses the required HTML templates
func loadTemplates() *template.Template {
	tmpl, err := template.ParseFiles(
		"templates/index.html",        // List of blocks
		"templates/blockDetails.html", // Block details
	)
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	return tmpl
}

// buildBlockCache creates a map of block hashes for fast lookup
func buildBlockCache(bc *blockchain.Blockchain) map[string]*blockchain.Block {
	cache := make(map[string]*blockchain.Block)
	for _, block := range bc.Blocks {
		cache[string(block.Hash)] = block
	}
	return cache
}

// Start initializes and starts the blockchain explorer on the specified port
func (e *Explorer) Start(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", e.renderBlocks)
	mux.HandleFunc("/block/", e.renderBlockDetails)

	log.Printf("Explorer running at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

// renderBlocks renders the list of blocks in the blockchain
func (e *Explorer) renderBlocks(w http.ResponseWriter, r *http.Request) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	blocks := e.getBlocksData()
	err := e.templates.ExecuteTemplate(w, "index.html", blocks)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

// renderBlockDetails renders the details of a specific block
func (e *Explorer) renderBlockDetails(w http.ResponseWriter, r *http.Request) {
	blockHash := r.URL.Path[len("/block/"):]
	block := e.getBlockByHash(blockHash)
	if block == nil {
		http.NotFound(w, r)
		return
	}

	blockData := BlockData{
		Height:        e.getBlockHeight(block),
		Timestamp:     time.Unix(block.Timestamp, 0).Format(time.RFC3339),
		Transactions:  len(block.Transactions),
		Hash:          string(block.Hash),
		PrevBlockHash: string(block.PrevBlockHash),
		Nonce:         block.Nonce,
	}

	err := e.templates.ExecuteTemplate(w, "blockDetails.html", blockData)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

// getBlocksData converts all blockchain blocks into a display-friendly format
func (e *Explorer) getBlocksData() []BlockData {
	var blocks []BlockData

	for i, block := range e.Blockchain.Blocks {
		blocks = append(blocks, BlockData{
			Height:        i,
			Timestamp:     time.Unix(block.Timestamp, 0).Format(time.RFC3339),
			Transactions:  len(block.Transactions),
			PrevBlockHash: string(block.PrevBlockHash),
			Hash:          string(block.Hash),
			Nonce:         block.Nonce,
		})
	}

	return blocks
}

// getBlockByHash finds a block by its hash in the blockchain using the cache
func (e *Explorer) getBlockByHash(hash string) *blockchain.Block {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.blockCache[hash]
}

// getBlockHeight returns the height of a block in the blockchain
func (e *Explorer) getBlockHeight(block *blockchain.Block) int {
	for i, b := range e.Blockchain.Blocks {
		if b == block {
			return i
		}
	}
	return -1
}

func (e *Explorer) renderStats(w http.ResponseWriter, r *http.Request) {
	stats := struct {
		Height       int
		TotalBlocks  int
		TotalNodes   int
	}{
		Height:      e.Blockchain.GetBlockchainHeight(),
		TotalBlocks: len(e.Blockchain.GetAllBlocks()),
		TotalNodes:  len(e.Blockchain.Peers),
	}
	e.templates.ExecuteTemplate(w, "stats.html", stats)
}

func (explorer *Explorer) GetLatestBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := explorer.Blockchain.GetRecentBlocks(10) // Últimos 10 bloques
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

// Obtener estadísticas detalladas de un bloque
func (explorer *Explorer) GetBlockDetails(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "Block hash is required", http.StatusBadRequest)
		return
	}

	block, err := explorer.Blockchain.GetBlockByHash(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

func (explorer *Explorer) SearchBlock(w http.ResponseWriter, r *http.Request) {
    hash := r.URL.Query().Get("hash")
    block := explorer.Blockchain.GetBlockByHash(hash)
    if block == nil {
        http.NotFound(w, r)
        return
    }
    json.NewEncoder(w).Encode(block)
}