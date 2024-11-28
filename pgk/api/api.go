package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ANCAPCOIN/pkg/blockchain"
	"ANCAPCOIN/pkg/wallet"
	"golang.org/x/time/rate"
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

	port := ":8080"
	fmt.Printf("API running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

