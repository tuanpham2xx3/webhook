// Test script for webhook
// Usage: go run test/webhook_test.go <webhook_url> <secret>
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type TestPayload struct {
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
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run test/webhook_test.go <webhook_url> <secret>")
		fmt.Println("Example: go run test/webhook_test.go http://localhost:8300/deploy your_secret_here")
		os.Exit(1)
	}

	webhookURL := os.Args[1]
	secret := os.Args[2]

	// Tạo test payload
	payload := TestPayload{
		Repository: struct {
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			HTMLURL  string `json:"html_url"`
		}{
			Name:     "test-repo",
			FullName: "user/test-repo",
			HTMLURL:  "https://github.com/user/test-repo",
		},
		Pusher: struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}{
			Name:  "Test User",
			Email: "test@example.com",
		},
		HeadCommit: struct {
			ID      string `json:"id"`
			Message string `json:"message"`
			URL     string `json:"url"`
		}{
			ID:      "1234567890abcdef1234567890abcdef12345678",
			Message: "Test commit for webhook deployment",
			URL:     "https://github.com/user/test-repo/commit/1234567890abcdef1234567890abcdef12345678",
		},
		Ref: "refs/heads/main",
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Generate HMAC signature
	signature := generateSignature(jsonData, secret)

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	req.Header.Set("User-Agent", "GitHub-Hookshot/test")

	// Send request
	fmt.Printf("🔄 Sending test webhook to: %s\n", webhookURL)
	fmt.Printf("📝 Payload: %s\n", string(jsonData))
	fmt.Printf("🔐 Signature: sha256=%s\n", signature)
	fmt.Println("")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Print result
	fmt.Printf("📡 Response Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("📋 Response Body: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ Webhook test successful!")
	} else {
		fmt.Println("❌ Webhook test failed!")
		os.Exit(1)
	}
}

func generateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
