package delivery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CredentialsConfigured() bool {
	return strings.TrimSpace(os.Getenv("DELIVERY_SAPIENS_DB_PASSWORD")) != ""
}

func resolveScriptPath() (string, error) {
	candidates := []string{
		filepath.Join("..", "scripts", "sync_delivery_reference.py"),
		filepath.Join("scripts", "sync_delivery_reference.py"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("sync script not found (tried %v)", candidates)
}

func RunSync(dbPath string) (string, error) {
	if !CredentialsConfigured() {
		return "", fmt.Errorf("DELIVERY_SAPIENS_DB_PASSWORD is not set")
	}

	scriptPath, err := resolveScriptPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("python", scriptPath)
	cmd.Env = append(os.Environ(), "DB_PATH="+dbPath)
	output, err := cmd.CombinedOutput()
	message := strings.TrimSpace(string(output))
	if err != nil {
		if message == "" {
			return "", fmt.Errorf("delivery sync failed: %w", err)
		}
		return "", fmt.Errorf("delivery sync failed: %s", message)
	}
	return message, nil
}
