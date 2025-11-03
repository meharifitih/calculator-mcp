package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "calculator-mcp-server"
	serverVersion = "1.0.0"
)

// CalculateParams defines the parameters for the calculate tool.
type CalculateParams struct {
	Operation string  `json:"operation" jsonschema:"operation to be performed on the numbers"`
	Num1      float32 `json:"num1" jsonschema:"first number"`
	Num2      float32 `json:"num2" jsonschema:"second number"`
}

func (p CalculateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Operation,
			validation.Required,
			validation.In("add", "subtract", "multiply", "divide"),
		),
		validation.Field(&p.Num1, validation.Required),
		validation.Field(&p.Num2,
			validation.Required,
			validation.By(func(value interface{}) error {
				if p.Operation == "divide" && p.Num2 == 0 {
					return errors.New("cannot divide by zero")
				}
				return nil
			}),
		),
	)
}

// CalculateResult defines the result for the calculate tool.
type CalculateResult struct {
	Result float32 `json:"result" jsonschema:"result of the operation"`
}

func createMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion},
		nil)
	log.Printf("Initializing MCP server: %s v%s", serverName, serverVersion)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calculate",
		Description: "Perform basic mathematical operations like add, subtract, multiply, and divide",
	}, handleCalculate)

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
			log.Fatalf("Stdio server error: %v", err)
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

func handleCalculate(ctx context.Context, req *mcp.CallToolRequest, param CalculateParams) (*mcp.CallToolResult, CalculateResult, error) {
	if err := param.Validate(); err != nil {
		return &mcp.CallToolResult{IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)}}},
			CalculateResult{}, fmt.Errorf("invalid parameters: %v", err)
	}

	var result float32
	switch param.Operation {
	case "add":
		result = param.Num1 + param.Num2
	case "subtract":
		result = param.Num1 - param.Num2
	case "multiply":
		result = param.Num1 * param.Num2
	case "divide":
		result = param.Num1 / param.Num2
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Result: %f", result)}},
	}, CalculateResult{Result: result}, nil
}
