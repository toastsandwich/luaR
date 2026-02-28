# luaR

A Lua configuration reader and writer for Go with struct tag support.

## Installation

```bash
go get github.com/toastsandwich/luaR
```

## Usage

### Decoding Lua Config

Define a struct with `lua` tags (similar to `json` tags):

```go
package main

import (
    "fmt"
    "github.com/toastsandwich/luaR"
)

type Config struct {
    AppName   string         `lua:"app_name"`
    Debug     bool           `lua:"debug"`
    Port      int            `lua:"port"`
    Database  DatabaseConfig `lua:"database"`
    RabbitMQ  string         `lua:"rabbitmq"`
}

type DatabaseConfig struct {
    Host     string `lua:"host"`
    Port     int    `lua:"port"`
    User     string `lua:"user"`
    Password string `lua:"password"`
}

func main() {
    luaData := []byte(`
        app_name = "MyApp"
        debug = true
        port = 8080

        database = {
            host = "localhost",
            port = 5432,
            user = "admin",
            password = "secret"
        }

        rabbitmq = "amqp://localhost"
    `)

    var config Config
    if err := luar.Unmarshal(luaData, &config); err != nil {
        panic(err)
    }

    fmt.Printf("%+v\n", config)
    // Output: {AppName:MyApp Debug:true Port:8080 Database:{Host:localhost Port:5432 User:admin Password:secret} RabbitMQ:amqp://localhost}
}
```

### Using Decoder/Encoder with io.Reader/io.Writer

```go
// Decode from any io.Reader
reader := strings.NewReader(`name = "test"`)
var config Config
if err := luar.NewDecoder(reader).Decode(&config); err != nil {
    panic(err)
}

// Encode to any io.Writer
writer := os.Stdout
if err := luar.NewEncoder(writer).Encode(config); err != nil {
    panic(err)
}
```

### Encoding to Lua

```go
config := Config{
    AppName:  "MyApp",
    Debug:    true,
    Port:     8080,
    Database: DatabaseConfig{
        Host: "localhost",
        Port: 5432,
    },
}

data, err := luar.Marshal(config)
if err != nil {
    panic(err)
}

fmt.Println(string(data))
// Output:
// app_name = "MyApp"
// debug = true
// port = 8080
// database = {host = "localhost", port = 5432}
```

## API

### Unmarshal

```go
func Unmarshal(data []byte, v interface{}) error
```

Parses Lua data and populates the Go struct pointed to by `v`.

### NewDecoder

```go
func NewDecoder(r io.Reader) *Decoder
```

Creates a Decoder from an `io.Reader`.

### Decoder.Decode

```go
func (d *Decoder) Decode(v interface{}) error
```

Decodes Lua data into the provided Go value.

### Marshal

```go
func Marshal(v interface{}) ([]byte, error)
```

Encodes a Go struct to Lua format.

### NewEncoder

```go
func NewEncoder(w io.Writer) *Encoder
```

Creates an Encoder that writes to an `io.Writer`.

### Encoder.Encode

```go
func (e *Encoder) Encode(v interface{}) error
```

Encodes a Go value to Lua format.

## Struct Tags

The decoder supports `lua` struct tags:

```go
type Config struct {
    AppName string `lua:"app_name"`  // Maps to "app_name" in Lua
    Name    string                   // Maps to "name" (lowercase field name)
}
```

If no `lua` tag is specified, the field name is converted to lowercase.

## Supported Types

- `string`, `int`, `int8`, `int16`, `int32`, `int64`
- `float32`, `float64`
- `bool`
- `nil`
- Nested structs
- Maps (`map[string]interface{}`)
- Slices

## Development

### Running Tests

```bash
go test -v
```

### Project Structure

```
luaR/
├── lexer.go       # Lua tokenizer
├── lexer_test.go  # Lexer tests
├── tokens.go      # Token type definitions
├── ast.go         # AST node types
├── ast_test.go    # AST tests
├── parser.go      # Lua parser
├── parser_test.go # Parser tests
├── luar.go        # Decoder/Encoder implementation
├── luar_test.go   # Decoder/Encoder tests
└── examples/
    └── config.lua # Example Lua config file
```

## License

MIT
