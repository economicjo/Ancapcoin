package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"sync"
	"example.com/ancapcoin/pkg/blockchain"
	"example.com/ancapcoin/pkg/wallet"
	"golang.org/x/time/rate"
)

var (
    proposalMutex sync.Mutex
    Proposals     = make(map[string]*blockchain.Proposal)
)


// API represents the REST API for the blockchain
type API struct {
	Blockchain *blockchain.Blockchain
	RateLimiter *rate.Limiter
}

// NewAPI initializes a new API instance
func NewAPI(bc *blockchain.Blockchain) *API {
	return &API{
		Blockchain:  bc,
		RateLimiter: rate.NewLimiter(rate.Every(time.Second), 5), // Max 5 requests per second
	}
}

// Response structure for API responses
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Helper function to write JSON responses
func (api *API) writeJSONResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := APIResponse{
		Status:  http.StatusText(status),
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// --- Middleware Features ---

// Logger middleware logs each incoming request
func (api *API) Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}

// Auth middleware checks for a valid API key in headers
func (api *API) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey == "" || apiKey != "your-secure-api-key" {
			api.writeJSONResponse(w, http.StatusUnauthorized, "Unauthorized: Missing or invalid API key", nil)
			return
		}
		next(w, r)
	}
}

// CORS middleware handles cross-origin resource sharing
func (api *API) CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-KEY")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// RateLimit middleware limits the rate of requests
func (api *API) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.RateLimiter.Allow() {
			api.writeJSONResponse(w, http.StatusTooManyRequests, "Too many requests", nil)
			return
		}
		next(w, r)
	}
}

// --- API Handlers ---

// GetBlocks returns all blocks in the blockchain
func (api *API) GetBlocks(w http.ResponseWriter, r *http.Request) {
	var blocks []blockchain.Block
	iter := api.Blockchain.Iterator()

	for {
		block := iter.Next()
		blocks = append(blocks, *block)
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	api.writeJSONResponse(w, http.StatusOK, "Success", blocks)
}

// GetBlockByID retrieves a specific block by its hash or index
func (api *API) GetBlockByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		api.writeJSONResponse(w, http.StatusBadRequest, "Missing block ID or index", nil)
		return
	}

	block, err := api.Blockchain.GetBlockByID(id)
	if err != nil {
		api.writeJSONResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	api.writeJSONResponse(w, http.StatusOK, "Success", block)
}

// AddBlock adds a new block to the blockchain
func (api *API) AddBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.writeJSONResponse(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
		return
	}

	var req struct {
		Transactions []string `json:"transactions"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Transactions) == 0 {
		api.writeJSONResponse(w, http.StatusBadRequest, "Invalid input data", nil)
		return
	}

	api.Blockchain.AddBlock(req.Transactions)
	api.writeJSONResponse(w, http.StatusOK, "Block added successfully", nil)
}

// GetBalance retrieves the balance of a wallet address
func (api *API) GetBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if strings.TrimSpace(address) == "" {
		api.writeJSONResponse(w, http.StatusBadRequest, "Missing wallet address", nil)
		return
	}

	balance := api.Blockchain.CalculateBalance(address)
	api.writeJSONResponse(w, http.StatusOK, "Success", map[string]interface{}{
		"address": address,
		"balance": balance,
	})
}

// --- API Initialization ---

func (api *API) Start() {
	http.HandleFunc("/blocks", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetBlocks)))))
	http.HandleFunc("/block", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetBlockByID)))))
	http.HandleFunc("/add-block", api.Logger(api.CORS(api.Auth(api.RateLimit(api.AddBlock)))))
	http.HandleFunc("/balance", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetBalance)))))
	http.HandleFunc("/network-stats", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetNetworkStats)))))
	http.HandleFunc("/wallet-history", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetWalletHistory)))))
	http.HandleFunc("/proposals/create", api.Logger(api.CORS(api.Auth(api.RateLimit(api.CreateProposal)))))
	http.HandleFunc("/proposals/vote", api.Logger(api.CORS(api.Auth(api.RateLimit(api.VoteForProposal)))))
	http.HandleFunc("/proposals/close", api.Logger(api.CORS(api.Auth(api.RateLimit(api.CloseProposal)))))
	http.HandleFunc("/fees", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetDynamicFees)))))
	http.HandleFunc("/cross-chain/event", api.Logger(api.CORS(api.Auth(api.RateLimit(api.HandleCrossChainEvent)))))
	http.HandleFunc("/audit/mempool", api.Logger(api.CORS(api.Auth(api.RateLimit(api.AnalyzeMempool)))))
	http.HandleFunc("/cross-chain/state", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetCrossChainState)))))
	http.HandleFunc("/economic-state", api.Logger(api.CORS(api.Auth(api.RateLimit(api.GetEconomicState)))))

	port := ":8080"
	fmt.Printf("API running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (api *API) GetNetworkStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"totalBlocks":    len(api.Blockchain.GetAllBlocks()),
		"totalTxs":       api.Blockchain.GetTotalTransactions(),
		"activeContracts": api.Blockchain.GetActiveContracts(),
		"pendingTxs":     len(api.Blockchain.GetMempool()),
	}
	api.writeJSONResponse(w, http.StatusOK, "Success", stats)
}

// Endpoint para obtener el historial de saldos de una billetera
func (api *API) GetWalletHistory(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if strings.TrimSpace(address) == "" {
		api.writeJSONResponse(w, http.StatusBadRequest, "Missing wallet address", nil)
		return
	}

	history := api.Blockchain.GetTransactionHistory(address)
	api.writeJSONResponse(w, http.StatusOK, "Success", history)
}

func (api *API) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
    balance := api.Blockchain.CalculateBalance(api.Wallet.GetAddress())
    json.NewEncoder(w).Encode(map[string]interface{}{
        "balance": balance,
    })
}

func (api *API) CreateProposal(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        api.writeJSONResponse(w, http.StatusMethodNotAllowed, "Invalid request method", nil)
        return
    }

    var req struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil || req.Title == "" || req.Description == "" {
        api.writeJSONResponse(w, http.StatusBadRequest, "Invalid input data", nil)
        return
    }

    proposal := blockchain.NewProposal(req.Title, req.Description)
    proposalMutex.Lock()
    Proposals[proposal.ID] = proposal
    proposalMutex.Unlock()

    api.writeJSONResponse(w, http.StatusOK, "Proposal created successfully", proposal)
}

// Vote for a proposal
func (api *API) VoteForProposal(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ProposalID string `json:"proposal_id"`
        Vote       string `json:"vote"` // "yes" or "no"
    }

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil || req.ProposalID == "" || (req.Vote != "yes" && req.Vote != "no") {
        api.writeJSONResponse(w, http.StatusBadRequest, "Invalid input data", nil)
        return
    }

    proposalMutex.Lock()
    defer proposalMutex.Unlock()

    proposal, exists := Proposals[req.ProposalID]
    if !exists {
        api.writeJSONResponse(w, http.StatusNotFound, "Proposal not found", nil)
        return
    }

    proposal.AddVote(req.Vote)
    api.writeJSONResponse(w, http.StatusOK, "Vote added successfully", proposal)
}

// Close a proposal
func (api *API) CloseProposal(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ProposalID string `json:"proposal_id"`
    }

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil || req.ProposalID == "" {
        api.writeJSONResponse(w, http.StatusBadRequest, "Invalid input data", nil)
        return
    }

    proposalMutex.Lock()
    defer proposalMutex.Unlock()

    proposal, exists := Proposals[req.ProposalID]
    if !exists {
        api.writeJSONResponse(w, http.StatusNotFound, "Proposal not found", nil)
        return
    }

    if proposal.Status != "open" {
        api.writeJSONResponse(w, http.StatusConflict, "Proposal already closed", nil)
        return
    }

    proposal.Close()
    api.writeJSONResponse(w, http.StatusOK, "Proposal closed successfully", proposal)
}


func (api *API) GetDynamicFees(w http.ResponseWriter, r *http.Request) {
    mempoolSize := len(api.Blockchain.GetMempool())
    fee := AdjustFees(mempoolSize)

    api.writeJSONResponse(w, http.StatusOK, "Success", map[string]int64{
        "current_fee": fee,
    })
}

// Agregar a la función Start

func (api *API) GetDocumentation(w http.ResponseWriter, r *http.Request) {
    documentation := map[string]interface{}{
        "architecture": map[string]string{
            "description": "AncapCoin is a decentralized cryptocurrency based on blockchain technology...",
            "consensus":   "Supports PoW, PoS, DPoS, and PoA",
            "mempool":     "Prioritizes transactions dynamically based on fees and congestion",
        },
        "api": []map[string]string{
            {"endpoint": "/blocks", "method": "GET", "description": "Retrieve all blocks in the blockchain"},
            {"endpoint": "/balance", "method": "GET", "description": "Get wallet balance by address"},
            {"endpoint": "/proposals/create", "method": "POST", "description": "Create a new governance proposal"},
            {"endpoint": "/proposals/vote", "method": "POST", "description": "Vote for a proposal"},
        },
        "protocol": map[string]string{
            "type": "P2P",
            "encryption": "TLS for secure communication",
            "transactions": "Validated based on UTXO and fees",
        },
    }

    api.writeJSONResponse(w, http.StatusOK, "API Documentation", documentation)
}

func (api *API) VerifyBlockchain(w http.ResponseWriter, r *http.Request) {
    err := api.Blockchain.Validate()
    if err != nil {
        api.writeJSONResponse(w, http.StatusInternalServerError, "Blockchain integrity check failed", err.Error())
        return
    }
    api.writeJSONResponse(w, http.StatusOK, "Blockchain integrity check passed", nil)
}



func (api *API) HandleCrossChainEvent(w http.ResponseWriter, r *http.Request) {
    var event blockchain.CrossChainEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        api.writeJSONResponse(w, http.StatusBadRequest, "Invalid cross-chain event data", nil)
        return
    }

    // Procesa el evento recibido
    if err := blockchain.ProcessIncomingCrossChainEvent(event); err != nil {
        api.writeJSONResponse(w, http.StatusInternalServerError, "Failed to process cross-chain event", err.Error())
        return
    }

    api.writeJSONResponse(w, http.StatusOK, "Cross-chain event processed successfully", nil)
}



func (api *API) GetCrossChainState(w http.ResponseWriter, r *http.Request) {
    chain := r.URL.Query().Get("chain")
    state, exists := blockchain.CrossChainStates[chain]
    if !exists {
        api.writeJSONResponse(w, http.StatusNotFound, "Cross-chain state not found", nil)
        return
    }
    api.writeJSONResponse(w, http.StatusOK, "Success", state)
}


func (api *API) GetEconomicState(w http.ResponseWriter, r *http.Request) {
    state := map[string]interface{}{
        "TotalSupply":    blockchain.TotalSupply,
        "MaxSupply":      blockchain.MaxSupply,
        "CommunityFund":  blockchain.CommunityFund.Balance,
        "BlockReward":    blockchain.CalculateBlockReward(len(api.Blockchain.Blocks)),
        "MempoolSize":    len(api.Blockchain.GetMempool()),
        "BurnRate":       "Variable", // Según las tarifas
    }
    api.writeJSONResponse(w, http.StatusOK, "Success", state)
}