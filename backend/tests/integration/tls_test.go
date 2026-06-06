package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getRepoRoot() string {
	wd, _ := os.Getwd()
	// From backend/tests/integration, go up 3 levels to repo root
	return filepath.Join(wd, "..", "..", "..")
}

func TestNoHardcodedSecretsInDockerCompose(t *testing.T) {
	repoRoot := getRepoRoot()
	files := []string{
		filepath.Join(repoRoot, "docker-compose.yml"),
		filepath.Join(repoRoot, "docker-compose.prod.yml"),
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

		if strings.Contains(contentStr, "POSTGRES_PASSWORD: ") && !strings.Contains(contentStr, "POSTGRES_PASSWORD: ${") {
			t.Errorf("%s contains hardcoded POSTGRES_PASSWORD", file)
		}
		if strings.Contains(contentStr, "DB_PASSWORD: ") && !strings.Contains(contentStr, "DB_PASSWORD: ${") {
			t.Errorf("%s contains hardcoded DB_PASSWORD", file)
		}
		if strings.Contains(contentStr, "JWT_SECRET: ") && !strings.Contains(contentStr, "JWT_SECRET: ${") {
			t.Errorf("%s contains hardcoded JWT_SECRET", file)
		}
		if strings.Contains(contentStr, "REDIS_PASSWORD: ") && !strings.Contains(contentStr, "REDIS_PASSWORD: ${") {
			t.Errorf("%s contains hardcoded REDIS_PASSWORD", file)
		}
	}
}

func TestCaddyfileExists(t *testing.T) {
	repoRoot := getRepoRoot()
	content, err := os.ReadFile(filepath.Join(repoRoot, "Caddyfile"))
	if err != nil {
		t.Fatalf("Failed to read Caddyfile: %v", err)
	}

	contentStr := string(content)
	assert.Contains(t, contentStr, "reverse_proxy", "Caddyfile should contain reverse_proxy directive")
	assert.Contains(t, contentStr, "Content-Security-Policy", "Caddyfile should contain CSP header")
	assert.Contains(t, contentStr, "Strict-Transport-Security", "Caddyfile should contain HSTS header")
}

func TestDevCertsDirectoryExists(t *testing.T) {
	repoRoot := getRepoRoot()
	_, err := os.Stat(filepath.Join(repoRoot, "backend", "infrastructure", "certs", "dev-certs"))
	assert.NoError(t, err, "dev-certs directory should exist")
}

func TestGenerateDevCertsScriptExists(t *testing.T) {
	repoRoot := getRepoRoot()
	info, err := os.Stat(filepath.Join(repoRoot, "backend", "infrastructure", "certs", "generate-dev-certs.sh"))
	assert.NoError(t, err, "generate-dev-certs.sh should exist")
	if err == nil {
		assert.NotZero(t, info.Mode()&0111, "generate-dev-certs.sh should be executable")
	}
}
