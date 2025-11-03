package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "calculator-mcp-server"
	serverVersion = "1.0.0"
)

func createMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion},
		nil)
	log.Printf("Initializing MCP server: %s v%s", serverName, serverVersion)

	return server
}

func main() {

	transport := os.Getenv("TRANSPORT")
	if transport == "" {
		transport = "streamable-http"
	}

	server := createMCPServer()

	switch transport {
	case "stdio":
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal("Stdio server error: %v", err)
		}
	case "streamable-http":
		startHTTPServer(server)
	}
}

func startHTTPServer(s *mcp.Server) {
	port := "8080"
	if portStr := os.Getenv("PORT"); portStr != "" {
		port = portStr
	}

	log.Printf("starting server with streamable-http transport on port %s", port)

	// Create the streamable HTTP handler.
	handdler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s
	}, nil)

	// Create HTTP mux for additional endpoints
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Health check write error: %v", err)
		}
	})

	mux.Handle("/mcp", handdler)

	// Start HTTP server
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("HTTP server error %v", err)
	}

}
