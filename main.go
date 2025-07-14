package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Config struct {
	Port           string
	Secret         string
	DiscordWebhook string
}

type WebhookPayload struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	} `json:"head_commit"`
	Ref string `json:"ref"`

	// Support for GitHub Package Events (chuẩn)
	Package struct {
		Name     string `json:"name"`
		Version  string `json:"package_version"`
		Registry struct {
			Name string `json:"name"`
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"registry"`
	} `json:"package"`
	Action string `json:"action"` // "published" for package events

	// Support for Custom Workflow Payload (GitHub Actions)
	Docker struct {
		Registry       string `json:"registry"`
		ImageName      string `json:"image_name"`
		LatestTag      string `json:"latest_tag"`
		VersionedTag   string `json:"versioned_tag"`
		LatestImage    string `json:"latest_image"`
		VersionedImage string `json:"versioned_image"`
		PullCommand    string `json:"pull_command"`
	} `json:"docker"`

	Deployment struct {
		Environment string `json:"environment"`
		Branch      string `json:"branch"`
		Commit      string `json:"commit"`
		Timestamp   string `json:"timestamp"`
	} `json:"deployment"`
}

type DiscordMessage struct {
	Content string                `json:"content,omitempty"`
	Embeds  []DiscordMessageEmbed `json:"embeds,omitempty"`
}

type DiscordMessageEmbed struct {
	Title       string                     `json:"title,omitempty"`
	Description string                     `json:"description,omitempty"`
	Color       int                        `json:"color,omitempty"`
	Fields      []DiscordMessageEmbedField `json:"fields,omitempty"`
	Footer      *DiscordMessageEmbedFooter `json:"footer,omitempty"`
	Timestamp   string                     `json:"timestamp,omitempty"`
}

type DiscordMessageEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordMessageEmbedFooter struct {
	Text string `json:"text"`
}

var config Config

func init() {
	config = Config{
		Port:           getEnv("PORT", "8300"),
		Secret:         getEnv("WEBHOOK_SECRET", "your_secret_here"),
		DiscordWebhook: getEnv("DISCORD_WEBHOOK", "https://discord.com/api/webhooks/1393287834173050990/9Mb6VxMhpB_UOqf9HEXkbV85N0sLRIpeGDZqFHuQGiZwjzx_FQzt_Xh-Vg6ozo0PJcCa"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value) // Trim whitespace để tránh lỗi
	}
	return defaultValue
}

func main() {
	// In ra cấu hình khi start
	log.Printf("=== WEBHOOK CONFIGURATION ===")
	log.Printf("Port: %s", config.Port)
	log.Printf("Secret: %s", config.Secret)
	log.Printf("Discord Webhook: %s", config.DiscordWebhook)
	log.Printf("=============================")

	r := mux.NewRouter()

	// Middleware
	r.Use(loggingMiddleware)
	r.Use(rateLimitMiddleware)

	// Routes
	r.HandleFunc("/deploy", deployHandler).Methods("POST")
	r.HandleFunc("/health", healthHandler).Methods("GET")

	log.Printf("Starting webhook server on port %s", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, r))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("[%s] %s %s from %s",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			r.RemoteAddr)

		next.ServeHTTP(w, r)

		// Log duration
		log.Printf("Request completed in %v", time.Since(start))
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	// Simple rate limiting - can be enhanced with more sophisticated approach
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	// Security checks
	if !isValidRequest(r) {
		log.Printf("Unauthorized request from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	// Verify signature
	if !verifySignature(r, body) {
		log.Printf("Invalid signature from %s", r.RemoteAddr)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Get event type from header
	eventType := r.Header.Get("X-GitHub-Event")

	// Detect payload type and handle accordingly
	var payloadType string
	if payload.Docker.ImageName != "" && payload.Deployment.Environment != "" {
		// Custom Workflow Payload (từ GitHub Actions)
		payloadType = "workflow"
		log.Printf("Received workflow webhook for repository: %s, environment: %s, image: %s",
			payload.Repository.FullName, payload.Deployment.Environment, payload.Docker.LatestImage)
	} else if eventType == "package" && payload.Action == "published" {
		// GitHub Package Events (chuẩn)
		payloadType = "package"
		log.Printf("Received package webhook for repository: %s, package: %s@%s",
			payload.Repository.FullName, payload.Package.Name, payload.Package.Version)
	} else if eventType == "push" || payload.Ref != "" {
		// GitHub Push Events (chuẩn)
		payloadType = "push"
		log.Printf("Received push webhook for repository: %s, ref: %s", payload.Repository.FullName, payload.Ref)
	} else {
		log.Printf("Unknown payload type for repository: %s", payload.Repository.FullName)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ignored",
			"message": "Unknown payload type",
		})
		return
	}

	// Execute deployment (asynchronously)
	go func() {
		deploySuccess := executeDeployment(payload)
		sendDiscordNotification(payload, deploySuccess, payloadType)
	}()

	// Return immediate response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "accepted",
		"message": "Deployment initiated",
		"type":    payloadType,
	})
}

func isValidRequest(r *http.Request) bool {
	return true
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		return strings.Split(xForwardedFor, ",")[0]
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

func verifySignature(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		// Also check GitHub's alternative header
		signature = r.Header.Get("X-GitHub-Signature-256")
	}

	if signature == "" {
		log.Printf("No signature header found")
		return false
	}

	log.Printf("=== SIGNATURE VERIFICATION ===")
	log.Printf("Received signature: %s", signature)
	log.Printf("Using secret: %s", config.Secret)
	log.Printf("Payload length: %d bytes", len(body))
	result := checkSignature(body, signature, config.Secret)
	log.Printf("Verification result: %t", result)
	log.Printf("===============================")

	return result
}

func checkSignature(payload []byte, signature, secret string) bool {
	// Remove "sha256=" prefix if present
	originalSignature := signature
	if strings.HasPrefix(signature, "sha256=") {
		signature = signature[7:]
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	log.Printf("Original signature: %s", originalSignature)
	log.Printf("Cleaned signature: %s", signature)
	log.Printf("Expected signature: sha256=%s", expectedMAC)
	log.Printf("Signatures match: %t", hmac.Equal([]byte(signature), []byte(expectedMAC)))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func executeDeployment(payload WebhookPayload) bool {
	log.Printf("Starting deployment for %s", payload.Repository.FullName)

	// Get deployment commands based on project type and payload
	commands := getDeploymentCommands(payload.Repository.FullName)

	// If it's a workflow payload with Docker info, use Docker pull command
	if payload.Docker.ImageName != "" && payload.Docker.PullCommand != "" && payload.Docker.LatestImage != "" {
		log.Printf("Detected workflow payload with Docker info")
		log.Printf("Docker Image: %s", payload.Docker.LatestImage)
		log.Printf("Environment: %s", payload.Deployment.Environment)

		// Use custom Docker commands for workflow payloads
		dockerCommands := []string{
			payload.Docker.PullCommand,
			fmt.Sprintf("docker stop %s", payload.Docker.ImageName),
			fmt.Sprintf("docker rm %s", payload.Docker.ImageName),
		}

		// Add run command based on environment
		var runCommand string
		if payload.Deployment.Environment == "production" {
			runCommand = fmt.Sprintf("docker run -d --name %s -p 8100:8100 %s",
				payload.Docker.ImageName, payload.Docker.LatestImage)
		} else {
			runCommand = fmt.Sprintf("docker run -d --name %s-staging -p 8101:8100 %s",
				payload.Docker.ImageName, payload.Docker.LatestImage)
		}
		dockerCommands = append(dockerCommands, runCommand)

		// Use Docker commands if available, otherwise fall back to configured commands
		if len(dockerCommands) > 0 {
			commands = dockerCommands
			log.Printf("Using Docker commands from workflow payload")
		}
	} else if payload.Docker.ImageName != "" || payload.Docker.PullCommand != "" {
		log.Printf("Incomplete Docker payload info - ImageName: '%s', PullCommand: '%s', LatestImage: '%s'",
			payload.Docker.ImageName, payload.Docker.PullCommand, payload.Docker.LatestImage)
	}

	if len(commands) == 0 {
		log.Printf("No deployment commands configured for %s", payload.Repository.FullName)
		return false
	}

	// For Docker workflows, we don't need working directories - Docker handles everything
	var workingDir string

	// Check if we're using Docker workflow (from GitHub Actions)
	if payload.Docker.ImageName != "" && payload.Docker.PullCommand != "" && payload.Docker.LatestImage != "" {
		log.Printf("Using Docker workflow - no working directory needed")
	} else {
		// Only use working directory for non-Docker deployments
		workingDir = getWorkingDirectory(payload.Repository.FullName)

		// Verify working directory exists before using it
		if workingDir != "" {
			if _, err := os.Stat(workingDir); os.IsNotExist(err) {
				log.Printf("Warning: Working directory %s does not exist, continuing without changing directory", workingDir)
				workingDir = "" // Reset to empty so we don't use it
			} else {
				log.Printf("Using working directory: %s", workingDir)
			}
		}
	}

	for _, cmd := range commands {
		// Trim whitespace and skip empty commands
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		log.Printf("Executing: %s", cmd)

		parts := strings.Fields(cmd) // Use Fields instead of Split for better whitespace handling
		if len(parts) == 0 {
			log.Printf("Skipping empty command")
			continue
		}

		execCmd := exec.Command(parts[0], parts[1:]...)

		// Set working directory if specified and exists (only for non-Docker workflows)
		if workingDir != "" {
			execCmd.Dir = workingDir
			log.Printf("Running in directory: %s", workingDir)
		}

		output, err := execCmd.CombinedOutput()

		if err != nil {
			// Some Docker commands are expected to fail (like stopping non-existent containers)
			isDockerStopOrRm := strings.Contains(cmd, "docker stop") || strings.Contains(cmd, "docker rm")
			isContainerNotFound := strings.Contains(string(output), "No such container")

			if isDockerStopOrRm && isContainerNotFound {
				log.Printf("Command failed (expected): %s - Container doesn't exist, continuing...", cmd)
			} else {
				log.Printf("Command failed: %s, Error: %v, Output: %s", cmd, err, string(output))
				return false
			}
		}

		log.Printf("Command successful: %s", cmd)
		if len(output) > 0 {
			log.Printf("Output: %s", string(output))
		} else {
			log.Printf("Command completed with no output")
		}
	}

	log.Printf("Deployment completed successfully for %s", payload.Repository.FullName)
	return true
}

func getDeploymentCommands(repoName string) []string {
	// 1. Check for custom commands in environment variables (per repository)
	repoKey := strings.ReplaceAll(strings.ToUpper(repoName), "/", "_")
	repoKey = strings.ReplaceAll(repoKey, "-", "_")

	// Check for repository-specific commands
	if customCommands := os.Getenv("DEPLOY_COMMANDS_" + repoKey); customCommands != "" {
		return strings.Split(customCommands, ";")
	}

	// 2. Check for generic custom commands
	if customCommands := os.Getenv("DEPLOY_COMMANDS"); customCommands != "" {
		return strings.Split(customCommands, ";")
	}

	// 3. Auto-detect based on project type
	return autoDetectDeployCommands(repoName)
}

func autoDetectDeployCommands(repoName string) []string {
	workingDir := getWorkingDirectory(repoName)
	baseCommands := []string{"git pull origin main"}

	// Check for different project types by looking for marker files
	projectTypes := detectProjectType(workingDir)

	var buildCommands []string
	var serviceCommands []string

	for _, projectType := range projectTypes {
		switch projectType {
		case "go":
			buildCommands = append(buildCommands,
				"go mod tidy",
				"go build -o app",
			)
			serviceName := getServiceName(repoName, "app")
			serviceCommands = append(serviceCommands, "sudo systemctl restart "+serviceName)

		case "nodejs":
			buildCommands = append(buildCommands,
				"npm ci",
				"npm run build",
			)
			serviceName := getServiceName(repoName, "node-app")
			serviceCommands = append(serviceCommands, "sudo systemctl restart "+serviceName)

		case "python":
			buildCommands = append(buildCommands,
				"pip install -r requirements.txt",
			)
			serviceName := getServiceName(repoName, "python-app")
			serviceCommands = append(serviceCommands, "sudo systemctl restart "+serviceName)

		case "php":
			buildCommands = append(buildCommands,
				"composer install --no-dev --optimize-autoloader",
			)
			serviceCommands = append(serviceCommands,
				"sudo systemctl restart nginx",
				"sudo systemctl restart php-fpm",
			)

		case "java":
			buildCommands = append(buildCommands,
				"./mvnw clean package -DskipTests",
			)
			serviceName := getServiceName(repoName, "java-app")
			serviceCommands = append(serviceCommands, "sudo systemctl restart "+serviceName)

		case "dotnet":
			buildCommands = append(buildCommands,
				"dotnet restore",
				"dotnet build --configuration Release",
				"dotnet publish --configuration Release --output ./publish",
			)
			serviceName := getServiceName(repoName, "dotnet-app")
			serviceCommands = append(serviceCommands, "sudo systemctl restart "+serviceName)

		case "docker":
			buildCommands = append(buildCommands,
				"docker build -t "+strings.ToLower(repoName)+" .",
			)
			serviceCommands = append(serviceCommands,
				"docker-compose down",
				"docker-compose up -d",
			)
		}
	}

	// Combine all commands
	allCommands := baseCommands
	allCommands = append(allCommands, buildCommands...)
	allCommands = append(allCommands, serviceCommands...)

	return allCommands
}

func detectProjectType(workingDir string) []string {
	var types []string

	// Helper function to check if file exists
	fileExists := func(filename string) bool {
		if workingDir == "" {
			_, err := os.Stat(filename)
			return err == nil
		}
		_, err := os.Stat(workingDir + "/" + filename)
		return err == nil
	}

	// Check for different project types
	if fileExists("go.mod") || fileExists("main.go") {
		types = append(types, "go")
	}

	if fileExists("package.json") {
		types = append(types, "nodejs")
	}

	if fileExists("requirements.txt") || fileExists("setup.py") || fileExists("pyproject.toml") {
		types = append(types, "python")
	}

	if fileExists("composer.json") || fileExists("index.php") {
		types = append(types, "php")
	}

	if fileExists("pom.xml") || fileExists("build.gradle") {
		types = append(types, "java")
	}

	if fileExists("*.csproj") || fileExists("*.sln") {
		types = append(types, "dotnet")
	}

	if fileExists("Dockerfile") || fileExists("docker-compose.yml") {
		types = append(types, "docker")
	}

	// Default to go if no specific type detected
	if len(types) == 0 {
		types = []string{"go"}
	}

	return types
}

func getWorkingDirectory(repoName string) string {
	// Check for repository-specific working directory
	repoKey := strings.ReplaceAll(strings.ToUpper(repoName), "/", "_")
	repoKey = strings.ReplaceAll(repoKey, "-", "_")

	if workDir := os.Getenv("WORK_DIR_" + repoKey); workDir != "" {
		return workDir
	}

	// Check for generic working directory
	if workDir := os.Getenv("WORK_DIR"); workDir != "" {
		return workDir
	}

	return ""
}

func getServiceName(repoName, defaultName string) string {
	// Extract service name from repository name
	// e.g., "user/my-api" -> "my-api"
	parts := strings.Split(repoName, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return defaultName
}

func sendDiscordNotification(payload WebhookPayload, success bool, payloadType string) {
	log.Printf("Sending Discord notification...")

	color := 0x00ff00 // Green for success
	status := "✅ Deployment Successful"
	if !success {
		color = 0xff0000 // Red for failure
		status = "❌ Deployment Failed"
	}

	var fields []DiscordMessageEmbedField
	var title string

	// Handle different payload types
	if payloadType == "package" {
		// GitHub Package Events
		title = fmt.Sprintf("%s - Package Deployment", status)
		fields = []DiscordMessageEmbedField{
			{
				Name:   "Package",
				Value:  payload.Package.Name,
				Inline: true,
			},
			{
				Name:   "Version",
				Value:  payload.Package.Version,
				Inline: true,
			},
			{
				Name:   "Registry",
				Value:  payload.Package.Registry.Type,
				Inline: true,
			},
		}
	} else if payloadType == "workflow" {
		// Custom Workflow Payload (GitHub Actions)
		title = fmt.Sprintf("%s - Workflow Deployment", status)
		fields = []DiscordMessageEmbedField{
			{
				Name:   "Environment",
				Value:  payload.Deployment.Environment,
				Inline: true,
			},
			{
				Name:   "Branch",
				Value:  payload.Deployment.Branch,
				Inline: true,
			},
			{
				Name:   "Commit",
				Value:  payload.Deployment.Commit[:7],
				Inline: true,
			},
			{
				Name:   "Docker Image",
				Value:  payload.Docker.LatestImage,
				Inline: false,
			},
			{
				Name:   "Registry",
				Value:  payload.Docker.Registry,
				Inline: true,
			},
			{
				Name:   "Tags",
				Value:  fmt.Sprintf("latest: %s\nversioned: %s", payload.Docker.LatestTag, payload.Docker.VersionedTag),
				Inline: true,
			},
		}
	} else {
		// GitHub Push Events
		title = fmt.Sprintf("%s - Code Deployment", status)
		fields = []DiscordMessageEmbedField{
			{
				Name:   "Branch",
				Value:  strings.Replace(payload.Ref, "refs/heads/", "", 1),
				Inline: true,
			},
			{
				Name:   "Commit",
				Value:  fmt.Sprintf("[%s](%s)", payload.HeadCommit.ID[:7], payload.HeadCommit.URL),
				Inline: true,
			},
			{
				Name:   "Author",
				Value:  payload.Pusher.Name,
				Inline: true,
			},
			{
				Name:   "Message",
				Value:  payload.HeadCommit.Message,
				Inline: false,
			},
		}
	}

	embed := DiscordMessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("Repository: **%s**", payload.Repository.FullName),
		Color:       color,
		Fields:      fields,
		Footer: &DiscordMessageEmbedFooter{
			Text: "Auto Deploy Webhook",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	message := DiscordMessage{
		Embeds: []DiscordMessageEmbed{embed},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling Discord message: %v", err)
		return
	}

	resp, err := http.Post(config.DiscordWebhook, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error sending Discord notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("Discord webhook returned status: %d", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Discord response: %s", string(body))
	} else {
		log.Printf("Discord notification sent successfully")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
