package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SourceConfig represents the full YAML structure
type SourceConfig struct {
	SourceID string   `yaml:"source_id"` // Sẽ bỏ qua khi convert to DTO
	Name     string   `yaml:"name"`
	Members  []string `yaml:"members"`
	Metadata Metadata `yaml:"metadata"`
}

type Metadata struct {
	ProgrammingLanguage string `yaml:"programming_language"`
	Framework           string `yaml:"framework"`
	Module              string `yaml:"module"`
}

// GeneratorSourceDto - DTO không chứa source_id
type GeneratorSourceDto struct {
	AppName             string
	ProgrammingLanguage string
	Framework           string
	Module              string
	Members             []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate_source.go <path-to-service-folder>")
		fmt.Println("Example: go run generate_source.go sources-service/sample")
		os.Exit(1)
	}

	servicePath := os.Args[1]
	sourceFile := filepath.Join(servicePath, "source.yml")

	// Kiểm tra file phải là source.yml
	if filepath.Base(sourceFile) != "source.yml" {
		fmt.Printf("❌ Skipped: File must be named 'source.yml', got: %s\n", filepath.Base(sourceFile))
		os.Exit(0)
	}

	// Đọc file
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Printf("❌ Error reading file %s: %v\n", sourceFile, err)
		os.Exit(1)
	}

	// Parse YAML
	var config SourceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("❌ Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Convert to DTO (bỏ qua source_id)
	dto := GeneratorSourceDto{
		AppName:             config.Name,
		ProgrammingLanguage: config.Metadata.ProgrammingLanguage,
		Framework:           config.Metadata.Framework,
		Module:              config.Metadata.Module,
		Members:             config.Members,
	}

	// Print DTO
	fmt.Println("========================================")
	fmt.Println("        GENERATOR SOURCE DTO")
	fmt.Println("========================================")
	fmt.Printf("AppName:             %s\n", dto.AppName)
	fmt.Printf("ProgrammingLanguage: %s\n", dto.ProgrammingLanguage)
	fmt.Printf("Framework:           %s\n", dto.Framework)
	fmt.Printf("Module:              %s\n", dto.Module)
	fmt.Printf("Members:             %v\n", dto.Members)
	fmt.Println("========================================")
}
