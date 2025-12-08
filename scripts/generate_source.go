package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	printDTO(dto)

	// Process based on programming language
	if err := processService(dto); err != nil {
		fmt.Printf("❌ Error processing service: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Service generated and pushed successfully!")
}

func printDTO(dto GeneratorSourceDto) {
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

func processService(dto GeneratorSourceDto) error {
	switch dto.ProgrammingLanguage {
	case "golang":
		return processGolang(dto)
	case "nodejs":
		return processNodeJS(dto)
	default:
		return fmt.Errorf("unsupported programming language: %s", dto.ProgrammingLanguage)
	}
}

// getUranusBinary tìm uranus binary phù hợp với OS/Arch hiện tại
func getUranusBinary() (string, error) {
	// Tìm thư mục dist (relative to working directory)
	distDir := "dist"

	// Xác định binary name dựa vào OS và Architecture
	goos := runtime.GOOS     // darwin, linux, windows
	goarch := runtime.GOARCH // amd64, arm64

	binaryName := fmt.Sprintf("uranus-%s-%s", goos, goarch)
	binaryPath := filepath.Join(distDir, binaryName)

	// Kiểm tra binary tồn tại
	if _, err := os.Stat(binaryPath); err == nil {
		// Đảm bảo binary có quyền execute
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return "", fmt.Errorf("failed to chmod binary: %w", err)
		}
		fmt.Printf("📍 Found local binary: %s\n", binaryPath)
		return binaryPath, nil
	}

	// Nếu không tìm thấy binary local, fallback to go install
	fmt.Printf("⚠️  Local binary not found for %s-%s, using go install...\n", goos, goarch)
	if err := runCommand("go", "install", "github.com/tqhuy-dev/xgen-uranus@latest"); err != nil {
		return "", fmt.Errorf("failed to install uranus CLI: %w", err)
	}

	// Sau khi install, uranus sẽ nằm trong $GOPATH/bin hoặc $HOME/go/bin
	return "uranus", nil
}

func processGolang(dto GeneratorSourceDto) error {
	fmt.Println("\n🔧 Processing Golang service...")

	// Step 1: Tìm uranus binary
	fmt.Println("📦 Finding uranus CLI...")
	uranusBin, err := getUranusBinary()
	if err != nil {
		return fmt.Errorf("failed to get uranus binary: %w", err)
	}

	// Step 2: Generate app using uranus
	fmt.Printf("🚀 Generating app: %s\n", dto.AppName)
	if err := runCommand(uranusBin, "generate", "app",
		"--name", dto.AppName, "--module" , fmt.Sprintf("github.com/tqhuy-dev/%s" ,dto.AppName), "--skip_init=true"); err != nil {
		return fmt.Errorf("failed to generate app: %w", err)
	}

	// Step 3: Create GitHub repository
	fmt.Printf("📁 Creating GitHub repository: %s\n", dto.AppName)
	if err := createGitHubRepo(dto.AppName); err != nil {
		return fmt.Errorf("failed to create GitHub repo: %w", err)
	}

	// Step 4: Push code to repository
	fmt.Println("📤 Pushing code to repository...")
	if err := pushToRepo(dto.AppName); err != nil {
		return fmt.Errorf("failed to push to repo: %w", err)
	}

	return nil
}

func processNodeJS(dto GeneratorSourceDto) error {
	// TODO: Implement NodeJS processing (NestJS, Express, etc.)
	fmt.Println("⚠️ NodeJS processing not implemented yet")
	return nil
}

func runCommand(name string, args ...string) error {
	fmt.Printf("  → Running: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandInDir(dir string, name string, args ...string) error {
	fmt.Printf("  → Running in %s: %s %s\n", dir, name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createGitHubRepo(repoName string) error {
	// Sử dụng gh CLI để tạo repo (đã có sẵn trên GitHub Actions)
	// GH_TOKEN environment variable cần được set
	err := runCommand("gh", "repo", "create",
		fmt.Sprintf("tqhuy-dev/%s", repoName),
		"--private",
		"--confirm")

	if err != nil {
		// Repo có thể đã tồn tại, không phải lỗi critical
		fmt.Printf("  ⚠️ Note: %v (repo might already exist)\n", err)
	}
	return nil
}

func pushToRepo(appName string) error {
	// Generated code nằm trong folder có tên = appName
	repoDir := appName

	// Kiểm tra folder tồn tại
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("generated folder not found: %s", repoDir)
	}

	// Get GitHub token from environment
	ghToken := os.Getenv("GH_TOKEN")
	if ghToken == "" {
		ghToken = os.Getenv("GITHUB_TOKEN")
	}

	repoOwner := "tqhuy-dev"

	// Build repo URL with token for authentication
	var repoURL string
	if ghToken != "" {
		repoURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", ghToken, repoOwner, appName)
	} else {
		repoURL = fmt.Sprintf("https://github.com/%s/%s.git", repoOwner, appName)
	}

	// Git commands
	commands := []struct {
		name string
		args []string
	}{
		{"git", []string{"init"}},
		{"git", []string{"config", "user.email", "github-actions[bot]@users.noreply.github.com"}},
		{"git", []string{"config", "user.name", "github-actions[bot]"}},
		{"git", []string{"remote", "add", "origin", repoURL}},
		{"git", []string{"add", "-A"}},
		{"git", []string{"commit", "-m", "Initial commit from jupiter-registry"}},
		{"git", []string{"branch", "-M", "main"}},
		{"git", []string{"push", "-u", "origin", "main", "--force"}},
	}

	for _, cmd := range commands {
		if err := runCommandInDir(repoDir, cmd.name, cmd.args...); err != nil {
			return fmt.Errorf("command '%s %s' failed: %w", cmd.name, strings.Join(cmd.args, " "), err)
		}
	}

	return nil
}
