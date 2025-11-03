package main

mport (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type InPut struct {
	Name string `json:"name" jsonschema:"The name of the person to greet"`
}

type OutPut struct {
	Greating string `json:"greeting" jsonschema:"the greeting to tell the personn"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input InPut) (*mcp.CallToolResult, OutPut, error) {
	return nil, OutPut{Greating: "HI " + input.Name}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "Greeter", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
