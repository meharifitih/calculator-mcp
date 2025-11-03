package main

import (
	"context"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-clinet", Version: "1.0.0"}, nil)

	// Connect to a server over stdin/stdout.
	transport := mcp.CommandTransport{Command: exec.Command("./myserver")}

	session, err := client.Connect(ctx, &transport, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	param := mcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "mehari"},
	}

	res, err := session.CallTool(ctx, &param)
	if err != nil {
		log.Fatalf("Error calling tool: %v", err)
	}

	if res.IsError {
		log.Fatal("tool failed")
	}

	for _, c := range res.Content {
		log.Print(c.(*mcp.TextContent).Text)
	}

}
