package p2p

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

// Message represents a structured message between nodes
type Message struct {
	Type    string `json:"type"`    // Message type (e.g., "block", "transaction", "peer_request", "peer_response")
	Payload string `json:"payload"` // Message payload
}



// Node represents a P2P node
type Node struct {
	Address      string
	Peers        map[string]net.Conn
	mu           sync.Mutex
	messagePool  map[string]struct{}
	tlsConfig    *tls.Config // Configuración TLS para conexiones cifradas
	discoveryURL string      // URL para descubrir nodos
}

// NewNode initializes a new Node instance
func NewNode(address string) *Node {
	return &Node{
		Address:     address,
		Peers:       make(map[string]net.Conn),
		messagePool: make(map[string]struct{}),
	}
}

// Start starts the P2P node and listens for incoming connections
func (n *Node) Start() {
	ln, err := net.Listen("tcp", n.Address)
	if err != nil {
		log.Fatalf("Failed to start node on %s: %v", n.Address, err)
	}
	defer ln.Close()

	log.Printf("Node listening on %s\n", n.Address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Connection error: %v\n", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

// handleConnection handles an individual peer connection
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	n.addPeer(remoteAddr, conn)
	defer n.removePeer(remoteAddr)

	log.Printf("Connected to %s\n", remoteAddr)

	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Error decoding message from %s: %v\n", remoteAddr, err)
			return
		}
		n.handleMessage(msg, remoteAddr)
	}
}

// handleMessage processes received messages based on their type
func (n *Node) handleMessage(msg Message, remoteAddr string) {
	switch msg.Type {
	case "block":
		log.Printf("Received block: %s\n", msg.Payload)
	case "transaction":
		log.Printf("Received transaction: %s\n", msg.Payload)
	case "peer_request":
		n.sendPeers(remoteAddr)
	case "peer_response":
		var peers []string
		json.Unmarshal([]byte(msg.Payload), &peers)
		for _, peer := range peers {
			if _, exists := n.Peers[peer]; !exists && peer != n.Address {
				n.ConnectToPeer(peer)
			}
		}
	default:
		log.Printf("Unknown message type from %s: %s\n", remoteAddr, msg.Type)
	}
}

// Broadcast sends a message to all connected peers
func (n *Node) Broadcast(message Message) {
	hash := sha256.Sum256([]byte(message.Payload))
	msgHash := hex.EncodeToString(hash[:])

	n.mu.Lock()
	if _, exists := n.messagePool[msgHash]; exists {
		n.mu.Unlock()
		return // Message already broadcasted
	}
	n.messagePool[msgHash] = struct{}{}
	n.mu.Unlock()

	log.Printf("Broadcasting message: %s\n", message.Payload)
	for addr, conn := range n.Peers {
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(message); err != nil {
			log.Printf("Failed to send message to %s: %v\n", addr, err)
			n.removePeer(addr)
		}
	}
}

// addPeer safely adds a new peer to the node
func (n *Node) addPeer(address string, conn net.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Peers[address] = conn
}

// removePeer safely removes a peer from the node
func (n *Node) removePeer(address string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if conn, exists := n.Peers[address]; exists {
		conn.Close()
		delete(n.Peers, address)
		log.Printf("Disconnected from %s\n", address)
	}
}

// retryConnection attempts to reconnect to a peer after a delay
func (n *Node) retryConnection(address string, delay time.Duration) {
	for {
		time.Sleep(delay)
		if _, exists := n.Peers[address]; exists {
			return // Peer already connected
		}

		log.Printf("Retrying connection to %s\n", address)
		n.ConnectToPeer(address)
	}
}

// ConnectToPeer establishes a connection to a remote peer
func (n *Node) ConnectToPeer(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to connect to peer %s: %v\n", address, err)
		go n.retryConnection(address, 5*time.Second)
		return
	}

	n.addPeer(address, conn)
	go n.handleConnection(conn)
	log.Printf("Successfully connected to peer %s\n", address)
}

// sendPeers sends the list of connected peers to a remote node
func (n *Node) sendPeers(remoteAddr string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	var peerList []string
	for addr := range n.Peers {
		peerList = append(peerList, addr)
	}

	payload, _ := json.Marshal(peerList)
	message := Message{Type: "peer_response", Payload: string(payload)}

	if conn, exists := n.Peers[remoteAddr]; exists {
		encoder := json.NewEncoder(conn)
		encoder.Encode(message)
	}
}

// StartMetricsServer runs a metrics server to monitor the node
func (n *Node) StartMetricsServer(port string) {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		n.mu.Lock()
		defer n.mu.Unlock()

		stats := struct {
			Address    string `json:"address"`
			PeersCount int    `json:"peers_count"`
		}{
			Address:    n.Address,
			PeersCount: len(n.Peers),
		}

		json.NewEncoder(w).Encode(stats)
	})

	log.Printf("Metrics server running on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func (n *Node) SyncBlockchain(peerAddr string) {
	message := Message{Type: "blockchain_request"}
	n.SendMessageToPeer(peerAddr, message)
}

func (n *Node) HandleBlockchainResponse(payload string) {
	var chain []*blockchain.Block
	json.Unmarshal([]byte(payload), &chain)
	// Lógica para comparar cadenas y actualizar si es necesario
	n.Blockchain.ReplaceChain(chain)
}

func (n *Node) ValidateMessage(message Message, signature []byte, pubKey []byte) bool {
	hash := sha256.Sum256([]byte(message.Payload))
	return util.VerifySignature(hash[:], signature, pubKey)
}

func (n *Node) StartTLS(certFile, keyFile string) {
	tlsConfig, err := config.LoadTLSConfig(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to start TLS: %v", err)
	}
	n.tlsConfig = tlsConfig

	ln, err := tls.Listen("tcp", n.Address, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to start node with TLS on %s: %v", n.Address, err)
	}
	defer ln.Close()

	log.Printf("Secure node listening on %s\n", n.Address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Connection error: %v\n", err)
			continue
		}
		go n.handleConnection(conn)
	}
}

func (n *Node) DiscoverPeers() {
	resp, err := http.Get(n.discoveryURL)
	if err != nil {
		log.Printf("Failed to discover peers: %v", err)
		return
	}
	defer resp.Body.Close()

	var peers []string
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		log.Printf("Failed to decode peers: %v", err)
		return
	}

	for _, peer := range peers {
		if _, exists := n.Peers[peer]; !exists && peer != n.Address {
			go n.ConnectToPeer(peer)
		}
	}
}

func (n *Node) SyncSPV(peerAddr string) {
	message := Message{Type: "spv_request"}
	n.SendMessageToPeer(peerAddr, message)
}

func (n *Node) HandleSPVResponse(payload string) {
	var proof blockchain.MerkleProof
	json.Unmarshal([]byte(payload), &proof)

	// Lógica para validar el proof y agregar bloques si son válidos
	if proof.Validate() {
		log.Println("SPV proof validated successfully!")
	} else {
		log.Println("Invalid SPV proof received.")
	}
}