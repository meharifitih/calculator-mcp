package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

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

	log.Println("Loaded tools: calculate, random_number")

	// Math constants resource
	server.AddResource(&mcp.Resource{
		URI:         "math://constants",
		Name:        "math-constants",
		Description: "Mathematical constants",
	}, handleMathConstants)

	log.Println("Loaded resources: math-constants")

	// Calculation explanation prompt
	server.AddPrompt(&mcp.Prompt{
		Name:        "calculation-explanation",
		Description: "Explain how a mathematical calculation works",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "operation",
				Description: "The mathematical operation: add, subtract, multiply, or divide",
				Required:    true,
			},
			{
				Name:        "num1",
				Description: "The first number",
				Required:    true,
			},
			{
				Name:        "num2",
				Description: "The second number",
				Required:    true,
			},
		},
	}, handleCalculationExplanation)

	log.Println("Loaded prompts: calculation-explanation")

	// Random number generation prompt
	server.AddPrompt(&mcp.Prompt{
		Name:        "generate-random-number-prompt",
		Description: "Generate and explain a random number",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "min",
				Description: "Minimum value (default: 1)",
				Required:    false,
			},
			{
				Name:        "max",
				Description: "Maximum value (default: 100)",
				Required:    false,
			},
			{
				Name:        "distribution",
				Description: "Probability distribution: 'uniform' (default), 'normal' (Gaussian/bell curve), or 'exponential' (exponential decay)",
				Required:    false,
			},
		},
	}, handleGenerateRandomNumberPrompt)

	log.Println("Loaded prompts: calculation-explanation, generate-random-number-prompt")

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

func handleMathConstants(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	constants := map[string]float64{
		"pi":           3.141592653589793,
		"e":            2.718281828459045,
		"golden_ratio": 1.618033988749895,
		"sqrt2":        1.4142135623730951,
		"sqrt3":        1.7320508075688772,
		"ln2":          0.6931471805599453,
		"ln10":         2.302585092994046,
		"euler":        0.5772156649015329,
	}

	uri := req.Params.URI
	constantName := ""
	if uri == "math://constants" {
		// Return all constants as JSON
		jsonData := fmt.Sprintf(`{
  "pi": %f,
  "e": %f,
  "golden_ratio": %f,
  "sqrt2": %f,
  "sqrt3": %f,
  "ln2": %f,
  "ln10": %f,
  "euler": %f
}`, constants["pi"], constants["e"], constants["golden_ratio"],
			constants["sqrt2"], constants["sqrt3"], constants["ln2"],
			constants["ln10"], constants["euler"])
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, Text: jsonData, MIMEType: "application/json"},
			},
		}, nil
	}

	// Try to extract constant name from URI like "math://constants/pi"
	if len(uri) > len("math://constants/") && uri[:len("math://constants/")] == "math://constants/" {
		constantName = uri[len("math://constants/"):]
	}

	if constantName != "" {
		value, ok := constants[constantName]
		if !ok {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{URI: uri, Text: fmt.Sprintf("%f", value), MIMEType: "text/plain"},
			},
		}, nil
	}

	return nil, mcp.ResourceNotFoundError(uri)
}

func handleCalculationExplanation(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := req.Params.Arguments
	operation := args["operation"]
	num1Str := args["num1"]
	num2Str := args["num2"]

	if operation == "" || num1Str == "" || num2Str == "" {
		return &mcp.GetPromptResult{
			Description: "Calculation explanation prompt",
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "Please provide operation, num1, and num2 arguments"},
				},
			},
		}, nil
	}

	var num1, num2 float64
	var err error
	if num1, err = parseFloat(num1Str); err != nil {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Invalid number for num1: %s", num1Str)},
				},
			},
		}, nil
	}
	if num2, err = parseFloat(num2Str); err != nil {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Invalid number for num2: %s", num2Str)},
				},
			},
		}, nil
	}

	var result float64
	var explanation string
	switch operation {
	case "add":
		result = num1 + num2
		explanation = fmt.Sprintf("To add %g and %g, you simply combine the two numbers: %g + %g = %g", num1, num2, num1, num2, result)
	case "subtract":
		result = num1 - num2
		explanation = fmt.Sprintf("To subtract %g from %g, you take away the second number from the first: %g - %g = %g", num2, num1, num1, num2, result)
	case "multiply":
		result = num1 * num2
		explanation = fmt.Sprintf("To multiply %g by %g, you calculate the product: %g × %g = %g", num1, num2, num1, num2, result)
	case "divide":
		if num2 == 0 {
			return &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{
						Role:    "user",
						Content: &mcp.TextContent{Text: "Cannot divide by zero. Division by zero is undefined in mathematics."},
					},
				},
			}, nil
		}
		result = num1 / num2
		explanation = fmt.Sprintf("To divide %g by %g, you calculate the quotient: %g ÷ %g = %g", num1, num2, num1, num2, result)
	default:
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Invalid operation: %s. Valid operations are: add, subtract, multiply, divide", operation)},
				},
			},
		}, nil
	}

	message := fmt.Sprintf("%s\n\nResult: %g", explanation, result)

	return &mcp.GetPromptResult{
		Description: "Calculation explanation",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: message},
			},
		},
	}, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func handleGenerateRandomNumberPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := req.Params.Arguments

	// Parse arguments with defaults
	min := 1
	max := 100
	distribution := "uniform"

	if minStr := args["min"]; minStr != "" {
		if parsed, err := strconv.Atoi(minStr); err == nil {
			min = parsed
		}
	}
	if maxStr := args["max"]; maxStr != "" {
		if parsed, err := strconv.Atoi(maxStr); err == nil {
			max = parsed
		}
	}
	if distStr := args["distribution"]; distStr != "" {
		distribution = distStr
	}

	// Validate
	if min >= max {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Invalid range: min (%d) must be less than max (%d)", min, max)},
				},
			},
		}, nil
	}

	if distribution != "" && distribution != "uniform" && distribution != "normal" && distribution != "exponential" {
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Invalid distribution: %s. Valid distributions are: uniform, normal, exponential", distribution)},
				},
			},
		}, nil
	}

	// Generate random number
	var number int
	var explanation string

	switch distribution {
	case "uniform", "":
		number = rand.Intn(max-min+1) + min
		explanation = fmt.Sprintf("Using uniform distribution, each number between %d and %d has an equal probability of being selected.", min, max)
	case "normal":
		mean := float64(max+min) / 2.0
		stdDev := float64(max-min) / 6.0
		val := rand.NormFloat64()*stdDev + mean
		number = clamp(int(val), min, max)
		explanation = fmt.Sprintf("Using normal (Gaussian) distribution with mean %.2f and standard deviation %.2f. Values near the center (%d-%d) are more likely.", mean, stdDev, (min+max)/2-5, (min+max)/2+5)
	case "exponential":
		lambda := 1.0 / float64(max-min)
		val := rand.ExpFloat64() / lambda
		number = clamp(int(val)+min, min, max)
		explanation = fmt.Sprintf("Using exponential distribution. Lower values in the range (%d-%d) are more likely than higher values.", min, (min+max)/2)
	}

	message := fmt.Sprintf("%s\n\nGenerated random number: %d\nRange: [%d, %d]\nDistribution: %s", explanation, number, min, max, distribution)

	return &mcp.GetPromptResult{
		Description: "Random number generation",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: message},
			},
		},
	}, nil
}
