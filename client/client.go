package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "1.0.0"}, nil)

	var session *mcp.ClientSession
	var err error

	// Check if using HTTP transport
	transportType := os.Getenv("TRANSPORT")
	if transportType == "streamable-http" || transportType == "http" {
		// Connect via HTTP
		endpoint := os.Getenv("SERVER_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8080/mcp"
		}

		log.Printf("Connecting to server via HTTP: %s", endpoint)

		transport := &mcp.StreamableClientTransport{
			Endpoint:   endpoint,
			HTTPClient: &http.Client{},
		}

		session, err = client.Connect(ctx, transport, nil)
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	} else {
		// Connect via stdio (default)
		serverPath := findServerBinary()
		if serverPath == "" {
			log.Fatal("Server binary not found. Please build the server first: go build -o server server.go")
		}

		log.Printf("Connecting to server via stdio: %s", serverPath)

		transport := mcp.CommandTransport{Command: exec.Command(serverPath)}

		session, err = client.Connect(ctx, &transport, nil)
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	}

	defer session.Close()

	log.Println("Connected to server successfully!")

	// Test calculate tool
	log.Println("=== Testing Calculate Tool ===")
	testCalculateTool(ctx, session)

	// Test generate-random-number tool
	log.Println("\n=== Testing Generate Random Number Tool ===")
	testGenerateRandomNumber(ctx, session)

	// Test math constants resource
	log.Println("\n=== Testing Math Constants Resource ===")
	testMathConstants(ctx, session)

	// Test prompts
	log.Println("\n=== Testing Prompts ===")
	testPrompts(ctx, session)
}

func findServerBinary() string {
	// Try multiple possible locations, prioritizing calculator-mcp-server
	possiblePaths := []string{
		"../calculator-mcp-server",
		"../server",
		"./calculator-mcp-server",
		"./server",
		"../myserver",
		"./myserver",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err == nil {
				return absPath
			}
		}
	}

	return ""
}

func testCalculateTool(ctx context.Context, session *mcp.ClientSession) {
	tests := []struct {
		operation string
		num1      float32
		num2      float32
	}{
		{"add", 10, 5},
		{"subtract", 10, 5},
		{"multiply", 10, 5},
		{"divide", 10, 5},
	}

	for _, test := range tests {
		param := mcp.CallToolParams{
			Name: "calculate",
			Arguments: map[string]any{
				"operation": test.operation,
				"num1":      test.num1,
				"num2":      test.num2,
			},
		}

		res, err := session.CallTool(ctx, &param)
		if err != nil {
			log.Printf("  Error calling calculate (%s): %v", test.operation, err)
			continue
		}

		if res.IsError {
			log.Printf("  Calculate (%s) returned error", test.operation)
			for _, c := range res.Content {
				log.Printf("    Error: %s", c.(*mcp.TextContent).Text)
			}
			continue
		}

		for _, c := range res.Content {
			log.Printf("  %s: %s", test.operation, c.(*mcp.TextContent).Text)
		}
	}
}

func testGenerateRandomNumber(ctx context.Context, session *mcp.ClientSession) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "default (uniform 1-100)",
			args: map[string]any{},
		},
		{
			name: "custom range",
			args: map[string]any{
				"min": 10,
				"max": 50,
			},
		},
		{
			name: "normal distribution",
			args: map[string]any{
				"min":          1,
				"max":          100,
				"distribution": "normal",
			},
		},
		{
			name: "exponential distribution",
			args: map[string]any{
				"min":          1,
				"max":          100,
				"distribution": "exponential",
			},
		},
	}

	for _, test := range tests {
		param := mcp.CallToolParams{
			Name:      "generate-random-number",
			Arguments: test.args,
		}

		res, err := session.CallTool(ctx, &param)
		if err != nil {
			log.Printf("  Error (%s): %v", test.name, err)
			continue
		}

		if res.IsError {
			log.Printf("  %s returned error", test.name)
			continue
		}

		for _, c := range res.Content {
			log.Printf("  %s: %s", test.name, c.(*mcp.TextContent).Text)
		}
	}
}

func testMathConstants(ctx context.Context, session *mcp.ClientSession) {
	// Test reading all constants
	res, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "math://constants",
	})
	if err != nil {
		log.Printf("  Error reading math constants: %v", err)
		return
	}

	if len(res.Contents) > 0 {
		log.Printf("  All constants: %s", res.Contents[0].Text)
	}

	// Test reading a specific constant
	constants := []string{"pi", "e", "golden_ratio"}
	for _, constant := range constants {
		res, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
			URI: fmt.Sprintf("math://constants/%s", constant),
		})
		if err != nil {
			log.Printf("  Error reading %s: %v", constant, err)
			continue
		}

		if len(res.Contents) > 0 {
			log.Printf("  %s = %s", constant, res.Contents[0].Text)
		}
	}
}

func testPrompts(ctx context.Context, session *mcp.ClientSession) {
	// Test calculation explanation prompt
	log.Println("  Testing calculation-explanation prompt:")
	res, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "calculation-explanation",
		Arguments: map[string]string{
			"operation": "multiply",
			"num1":      "7",
			"num2":      "8",
		},
	})
	if err != nil {
		log.Printf("    Error: %v", err)
	} else {
		for _, msg := range res.Messages {
			log.Printf("    %s: %s", msg.Role, msg.Content.(*mcp.TextContent).Text)
		}
	}

	// Test generate-random-number-prompt
	log.Println("\n  Testing generate-random-number-prompt:")
	res, err = session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "generate-random-number-prompt",
		Arguments: map[string]string{
			"min":          "1",
			"max":          "50",
			"distribution": "normal",
		},
	})
	if err != nil {
		log.Printf("    Error: %v", err)
	} else {
		for _, msg := range res.Messages {
			log.Printf("    %s: %s", msg.Role, msg.Content.(*mcp.TextContent).Text)
		}
	}
}
