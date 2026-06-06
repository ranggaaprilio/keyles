package integration

import (
	"os"
	"strings"
	"testing"
)

// TestNoHardcodedSecretsInCommittedFiles verifies that committed configuration
// files do not contain literal secret values
func TestNoHardcodedSecretsInCommittedFiles(t *testing.T) {
	files := []string{
		"../../docker-compose.yml",
		"../../docker-compose.prod.yml",
		"../../backend/.env.example",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("Failed to read %s: %v", file, err)
		}

		contentStr := string(content)

		// Check for common hardcoded secret patterns (excluding example placeholders)
		if strings.Contains(contentStr, "password: ") && !strings.Contains(contentStr, "password: ${") {
			t.Errorf("%s may contain hardcoded password", file)
		}
		if strings.Contains(contentStr, "secret: ") && !strings.Contains(contentStr, "secret: ${") {
			t.Errorf("%s may contain hardcoded secret", file)
		}
		if strings.Contains(contentStr, "token: ") && !strings.Contains(contentStr, "token: ${") {
			t.Errorf("%s may contain hardcoded token", file)
		}
	}
}
