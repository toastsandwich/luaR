package luar

import (
	"os"
	"strings"
	"testing"
)

type TestConfig struct {
	AppName  string          `lua:"app_name"`
	Debug    bool            `lua:"debug"`
	Port     int             `lua:"port"`
	Database TestDatabaseCfg `lua:"database"`
	RabbitMQ string          `lua:"rabbitmq"`
}

type TestDatabaseCfg struct {
	Host     string `lua:"host"`
	Port     int    `lua:"port"`
	User     string `lua:"user"`
	Password string `lua:"password"`
}

type SimpleConfig struct {
	Name string `lua:"name"`
	Port int    `lua:"port"`
}

type NestedConfig struct {
	Outer InnerConfig `lua:"outer"`
}

type InnerConfig struct {
	Value string `lua:"value"`
}

func TestUnmarshal_BasicTypes(t *testing.T) {
	data := []byte(`
name = "MyApp"
port = 8080
enabled = true
`)
	var config SimpleConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "MyApp" {
		t.Errorf("Name: expected 'MyApp', got '%s'", config.Name)
	}
	if config.Port != 8080 {
		t.Errorf("Port: expected 8080, got %d", config.Port)
	}
}

func TestUnmarshal_Table(t *testing.T) {
	data := []byte(`
app_name = "MyApp"
debug = true
port = 8080

database = {
    host = "localhost",
    port = 5432,
    user = "admin",
    password = "secret"
}

rabbitmq = "mttq://localhost:8090"
`)
	var config TestConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	tests := []struct {
		got  interface{}
		want interface{}
		desc string
	}{
		{config.AppName, "MyApp", "AppName"},
		{config.Debug, true, "Debug"},
		{config.Port, 8080, "Port"},
		{config.Database.Host, "localhost", "Database.Host"},
		{config.Database.Port, 5432, "Database.Port"},
		{config.Database.User, "admin", "Database.User"},
		{config.Database.Password, "secret", "Database.Password"},
		{config.RabbitMQ, "mttq://localhost:8090", "RabbitMQ"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.want, tt.got)
		}
	}
}

func TestUnmarshal_NestedStruct(t *testing.T) {
	data := []byte(`
outer = {
    value = "test"
}
`)
	var config NestedConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Outer.Value != "test" {
		t.Errorf("Outer.Value: expected 'test', got '%s'", config.Outer.Value)
	}
}

func TestUnmarshal_File(t *testing.T) {
	data, err := os.ReadFile("examples/config.lua")
	if err != nil {
		t.Skip("examples/config.lua not found")
	}

	var config TestConfig
	err = Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.AppName != "MyApp" {
		t.Errorf("AppName: expected 'MyApp', got '%s'", config.AppName)
	}
	if config.Database.Host != "localhost" {
		t.Errorf("Database.Host: expected 'localhost', got '%s'", config.Database.Host)
	}
}

func TestDecoder(t *testing.T) {
	data := `name = "TestApp"
port = 3000`

	reader := strings.NewReader(data)
	var config SimpleConfig
	err := NewDecoder(reader).Decode(&config)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if config.Name != "TestApp" {
		t.Errorf("Name: expected 'TestApp', got '%s'", config.Name)
	}
	if config.Port != 3000 {
		t.Errorf("Port: expected 3000, got %d", config.Port)
	}
}

func TestMarshal_SimpleStruct(t *testing.T) {
	config := SimpleConfig{
		Name: "MyApp",
		Port: 8080,
	}

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "name = \"MyApp\"") {
		t.Errorf("expected 'name = \"MyApp\"' in output, got: %s", output)
	}
	if !strings.Contains(output, "port = 8080") {
		t.Errorf("expected 'port = 8080' in output, got: %s", output)
	}
}

func TestMarshal_NestedStruct(t *testing.T) {
	config := TestConfig{
		AppName:  "MyApp",
		Debug:    true,
		Port:     8080,
		RabbitMQ: "amqp://localhost",
		Database: TestDatabaseCfg{
			Host:     "localhost",
			Port:     5432,
			User:     "admin",
			Password: "secret",
		},
	}

	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	output := string(data)
	tests := []string{
		"app_name = \"MyApp\"",
		"debug = true",
		"port = 8080",
		"host = \"localhost\"",
		"port = 5432",
	}

	for _, want := range tests {
		if !strings.Contains(output, want) {
			t.Errorf("expected '%s' in output, got: %s", want, output)
		}
	}
}

func TestEncoder(t *testing.T) {
	config := SimpleConfig{
		Name: "Test",
		Port: 9000,
	}

	var buf strings.Builder
	err := NewEncoder(&buf).Encode(config)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name = \"Test\"") {
		t.Errorf("expected 'name = \"Test\"' in output, got: %s", output)
	}
}

func TestUnmarshal_EmptyTable(t *testing.T) {
	data := []byte(`empty = {}`)
	type WithEmpty struct {
		Empty map[string]interface{} `lua:"empty"`
	}
	var config WithEmpty
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestUnmarshal_MissingField(t *testing.T) {
	data := []byte(`name = "test"`)
	var config SimpleConfig
	err := Unmarshal(data, &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if config.Name != "test" {
		t.Errorf("Name: expected 'test', got '%s'", config.Name)
	}
	if config.Port != 0 {
		t.Errorf("Port: expected 0 (default), got %d", config.Port)
	}
}

func TestMarshal_Bool(t *testing.T) {
	type BoolConfig struct {
		Enabled bool `lua:"enabled"`
		Active  bool `lua:"active"`
	}

	config := BoolConfig{Enabled: true, Active: false}
	data, err := Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "enabled = true") {
		t.Errorf("expected 'enabled = true' in output")
	}
	if !strings.Contains(output, "active = false") {
		t.Errorf("expected 'active = false' in output")
	}
}

func TestRoundTrip(t *testing.T) {
	original := SimpleConfig{
		Name: "RoundTrip",
		Port: 9999,
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded SimpleConfig
	err = Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: expected '%s', got '%s'", original.Name, decoded.Name)
	}
	if decoded.Port != original.Port {
		t.Errorf("Port: expected %d, got %d", original.Port, decoded.Port)
	}
}
