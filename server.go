package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
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

type GenerateRandomNumberParams struct {
	Min          *int   `json:"min,omitempty" jsonschema:"minimum value (default: 1)"`
	Max          *int   `json:"max,omitempty" jsonschema:"maximum value (default: 100)"`
	Distribution string `json:"distribution,omitempty" jsonschema:"probability distribution: 'uniform' (default), 'normal' (Gaussian/bell curve), or 'exponential' (exponential decay)"`
}

func (p GenerateRandomNumberParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Distribution,
			validation.In("", "uniform", "normal", "exponential"),
		),
		validation.Field(&p.Min),
		validation.Field(&p.Max),
		validation.Field(&p.Min, validation.By(func(value interface{}) error {
			if p.Min != nil && p.Max != nil && *p.Min >= *p.Max {
				return errors.New("min must be less than max")
			}
			return nil
		})),
	)
}

type GenerateRandomNumberResult struct {
	Number int `json:"number" jsonschema:"generated random number"`
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

	// 	// Calculator tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "calculate",
		Description: "Perform basic mathematical operations like add, subtract, multiply, and divide",
	}, handleCalculate)

	// Random number generator tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate-random-number",
		Description: "Generate a random number between 1 and 100",
	}, handleGenerateRandomNumber)

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

func handleGenerateRandomNumber(ctx context.Context, req *mcp.CallToolRequest, param GenerateRandomNumberParams) (*mcp.CallToolResult, GenerateRandomNumberResult, error) {
	if err := param.Validate(); err != nil {
		return &mcp.CallToolResult{IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid parameters: %v", err)}}},
			GenerateRandomNumberResult{}, fmt.Errorf("invalid parameters: %v", err)
	}

	distribution := "uniform"
	if param.Distribution != "" {
		distribution = param.Distribution
	}

	min := 1
	max := 100
	if param.Min != nil {
		min = *param.Min
	}
	if param.Max != nil {
		max = *param.Max
	}

	var number int
	switch distribution {
	case "uniform", "":
		number = rand.Intn(max-min+1) + min
	case "normal":
		mean := float64(max+min) / 2.0
		stdDev := float64(max-min) / 6.0 // ~99.7% within range
		val := rand.NormFloat64()*stdDev + mean
		number = clamp(int(val), min, max)
	case "exponential":
		// Scale exponential to fit range, with rate parameter based on range
		lambda := 1.0 / float64(max-min)
		val := rand.ExpFloat64() / lambda
		number = clamp(int(val)+min, min, max)
	default:
		return &mcp.CallToolResult{IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid distribution: %s", distribution)}}},
			GenerateRandomNumberResult{}, fmt.Errorf("invalid distribution: %s", distribution)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Generated random number: %d (distribution: %s, range: [%d, %d])", number, distribution, min, max)}},
	}, GenerateRandomNumberResult{Number: number}, nil
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
