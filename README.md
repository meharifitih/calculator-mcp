# Calculator MCP Server

A comprehensive Model Context Protocol (MCP) server implementation providing mathematical calculation tools, random number generation, mathematical constants, and educational prompts. Built with Go and the MCP SDK.

## Features

### 🧮 Tools

1. **Calculate Tool** - Perform basic arithmetic operations
   - Addition, subtraction, multiplication, and division
   - Input validation with ozzo-validation
   - Division by zero protection

2. **Generate Random Number Tool** - Generate random numbers with various distributions
   - Uniform distribution (default)
   - Normal (Gaussian) distribution
   - Exponential distribution
   - Customizable min/max range (default: 1-100)

### 📚 Resources

- **Math Constants Resource** - Access mathematical constants
  - Available constants: π (pi), e, golden ratio, √2, √3, ln(2), ln(10), Euler's constant
  - Access via URI: `math://constants`
  - Returns JSON format with all constants

### 💡 Prompts

1. **Calculation Explanation Prompt** - Educational explanations of calculations
   - Explains how operations work step-by-step
   - Provides results with context

2. **Random Number Generation Prompt** - Explains random number generation
   - Describes the distribution used
   - Provides generated number with context

## Installation

### Prerequisites

- Go 1.24.0 or later
- Git

### Build Instructions

1. Clone or navigate to the project directory:
```bash
cd /Users/chipchip/go/src/calulator-mpc-server
```

2. Install dependencies:
```bash
go mod download
```

3. Build the server:
```bash
go build -o calculator-mcp-server server.go
```

4. Build the client (optional):
```bash
cd client
go build -o client client.go
```

## Usage

### Running the Server

The server supports two transport modes:

#### Stdio Mode (for Cursor and other stdio clients)

```bash
TRANSPORT=stdio ./calculator-mcp-server
```

#### Streamable HTTP Mode

```bash
TRANSPORT=streamable-http ./calculator-mcp-server
```

Or with custom port:
```bash
TRANSPORT=streamable-http PORT=3000 ./calculator-mcp-server
```

The HTTP server will be available at:
- MCP endpoint: `http://localhost:8080/mcp` (or custom port)
- Health check: `http://localhost:8080/health`

### Using the Client

The client can connect via stdio or HTTP:

#### Stdio Mode (default)
```bash
cd client
TRANSPORT=stdio go run .
```

#### HTTP Mode
First, start the server in HTTP mode:
```bash
TRANSPORT=streamable-http ./calculator-mcp-server
```

Then in another terminal:
```bash
cd client
TRANSPORT=streamable-http go run .
```

Or with custom server URL:
```bash
TRANSPORT=http SERVER_URL=http://localhost:3000/mcp go run .
```

## Configuration

### Cursor IDE Integration

To use this server with Cursor IDE, add the following to your `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "Calculator MCP Server": {
      "command": "/path/to/calculator-mcp-server",
      "args": [],
      "env": {
        "TRANSPORT": "stdio"
      },
      "disabled": false
    }
  }
}
```

**Important Notes:**
- Cursor uses stdio transport, so `TRANSPORT` must be set to `"stdio"`
- The server automatically detects when called via stdio and uses stdio mode regardless of the `TRANSPORT` env var
- Restart Cursor after adding the configuration

## API Documentation

### Tools

#### `calculate`

Performs basic mathematical operations.

**Parameters:**
- `operation` (string, required): One of `"add"`, `"subtract"`, `"multiply"`, `"divide"`
- `num1` (float32, required): First number
- `num2` (float32, required): Second number

**Example:**
```json
{
  "name": "calculate",
  "arguments": {
    "operation": "multiply",
    "num1": 7,
    "num2": 8
  }
}
```

**Response:**
```json
{
  "content": [{
    "type": "text",
    "text": "Result: 56.000000"
  }],
  "isError": false
}
```

#### `generate-random-number`

Generates a random number with optional distribution.

**Parameters:**
- `min` (int, optional): Minimum value (default: 1)
- `max` (int, optional): Maximum value (default: 100)
- `distribution` (string, optional): One of `"uniform"`, `"normal"`, `"exponential"` (default: `"uniform"`)

**Example:**
```json
{
  "name": "generate-random-number",
  "arguments": {
    "min": 10,
    "max": 50,
    "distribution": "normal"
  }
}
```

**Response:**
```json
{
  "content": [{
    "type": "text",
    "text": "Generated random number: 28 (distribution: normal, range: [10, 50])"
  }],
  "isError": false
}
```

### Resources

#### `math://constants`

Returns all mathematical constants in JSON format.

**Example Request:**
```
URI: math://constants
```

**Response:**
```json
{
  "pi": 3.141592653589793,
  "e": 2.718281828459045,
  "golden_ratio": 1.618033988749895,
  "sqrt2": 1.4142135623730951,
  "sqrt3": 1.7320508075688772,
  "ln2": 0.6931471805599453,
  "ln10": 2.302585092994046,
  "euler": 0.5772156649015329
}
```

### Prompts

#### `calculation-explanation`

Explains how a mathematical calculation works.

**Parameters:**
- `operation` (string, required): Operation to explain
- `num1` (string, required): First number
- `num2` (string, required): Second number

**Example:**
```json
{
  "name": "calculation-explanation",
  "arguments": {
    "operation": "divide",
    "num1": "20",
    "num2": "4"
  }
}
```

#### `generate-random-number-prompt`

Generates and explains a random number.

**Parameters:**
- `min` (string, optional): Minimum value
- `max` (string, optional): Maximum value
- `distribution` (string, optional): Distribution type

## Architecture

### Project Structure

```
calulator-mpc-server/
├── server.go              # Main server implementation
├── client/
│   └── client.go          # Test client
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
└── README.md             # This file
```

### Server Components

1. **Server Initialization** (`createMCPServer`)
   - Creates MCP server instance
   - Registers all tools, resources, and prompts
   - Sets up validation

2. **Transport Handling**
   - Auto-detects stdio vs HTTP based on stdin
   - Supports explicit `TRANSPORT` environment variable
   - Defaults to `streamable-http` when run interactively

3. **Validation**
   - Uses ozzo-validation for input validation
   - Custom validation rules for business logic
   - Division by zero protection

### Dependencies

- `github.com/modelcontextprotocol/go-sdk` - MCP SDK
- `github.com/go-ozzo/ozzo-validation/v4` - Input validation
- Standard Go libraries for math, HTTP, and JSON

## Testing

### Testing with Postman (Streamable HTTP)

You can test the streamable-http server using Postman or any HTTP client. First, start the server in HTTP mode:

```bash
TRANSPORT=streamable-http ./calculator-mcp-server
```

Then use Postman to send requests to `http://localhost:8080/mcp`.

#### Example: Calculate Tool

**Request:**
- Method: `POST`
- URL: `http://localhost:8080/mcp`
- Headers:
  - `Content-Type: application/json`
- Body (raw JSON):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "calculate",
    "arguments": {
      "operation": "multiply",
      "num1": 15,
      "num2": 23
    }
  }
}
```

#### Example: Generate Random Number

**Request:**
- Method: `POST`
- URL: `http://localhost:8080/mcp`
- Headers:
  - `Content-Type: application/json`
- Body (raw JSON):
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "generate-random-number",
    "arguments": {
      "min": 10,
      "max": 50,
      "distribution": "normal"
    }
  }
}
```

#### Example: Read Math Constants Resource

**Request:**
- Method: `POST`
- URL: `http://localhost:8080/mcp`
- Headers:
  - `Content-Type: application/json`
- Body (raw JSON):
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "resources/read",
  "params": {
    "uri": "math://constants"
  }
}
```

#### Example: Get Prompt

**Request:**
- Method: `POST`
- URL: `http://localhost:8080/mcp`
- Headers:
  - `Content-Type: application/json`
- Body (raw JSON):
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "prompts/get",
  "params": {
    "name": "calculation-explanation",
    "arguments": {
      "operation": "divide",
      "num1": "20",
      "num2": "4"
    }
  }
}
```

**Note:** Stdio mode cannot be tested with Postman as it requires stdin/stdout pipe communication. Use the provided test client or Cursor IDE for stdio testing.

### Testing with curl (Streamable HTTP)

```bash
# Calculate 15 × 23
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "calculate",
      "arguments": {
        "operation": "multiply",
        "num1": 15,
        "num2": 23
      }
    }
  }'
```

### Testing with the Test Client

#### Stdio Mode
```bash
cd client
TRANSPORT=stdio go run .
```

#### HTTP Mode
First, start the server:
```bash
TRANSPORT=streamable-http ./calculator-mcp-server
```

Then in another terminal:
```bash
cd client
TRANSPORT=streamable-http go run .
```

## Examples

### Example 1: Calculate 15 × 23

Using curl (as shown above) or Postman with the request body from the Testing section.

### Example 2: Generate Random Number with Normal Distribution

Using the test client or Postman with the request examples from the Testing section.

### Example 3: Access Math Constants

The constants are available as a resource that can be read by MCP clients. In Cursor, you can reference `math://constants` in your prompts. You can also test this via Postman using the resource read request example above.

## Validation Rules

### Calculate Tool
- Operation must be one of: `add`, `subtract`, `multiply`, `divide`
- Both numbers are required
- Division by zero is prevented

### Generate Random Number Tool
- Min must be less than max (if both provided)
- Distribution must be: `uniform`, `normal`, or `exponential`
- Range validation ensures min < max

## Error Handling

The server provides clear error messages:
- Invalid parameters return validation errors
- Division by zero returns specific error message
- Unknown tools/resources return appropriate MCP protocol errors

## Development

### Testing

Run the test client:
```bash
cd client
TRANSPORT=stdio go run .
```

This will test all tools, resources, and prompts.

### Adding New Features

1. **New Tool**: Add handler function and register in `createMCPServer()`
2. **New Resource**: Add resource handler and register with `server.AddResource()`
3. **New Prompt**: Add prompt handler and register with `server.AddPrompt()`

All handlers must follow the MCP SDK patterns shown in the existing code.

## License

This project is provided as-is for educational and development purposes.

## Contributing

Feel free to extend this server with additional mathematical functions, constants, or educational features.

## Version

Current version: **1.0.0**

---

For more information about the Model Context Protocol, visit: https://modelcontextprotocol.io

