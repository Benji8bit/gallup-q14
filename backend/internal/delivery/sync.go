package delivery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func resolveMirrorPath() string {
	if path := strings.TrimSpace(os.Getenv("DELIVERY_MIRROR_PATH")); path != "" {
		return path
	}
	candidates := []string{
		filepath.Join("..", "data", "delivery_mirror.db"),
		filepath.Join("data", "delivery_mirror.db"),
		filepath.Join("..", "backend", "data", "delivery_mirror.db"),
		filepath.Join("backend", "data", "delivery_mirror.db"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return candidates[0]
}

func MirrorAvailable() bool {
	_, err := os.Stat(resolveMirrorPath())
	return err == nil
}

func RunSync(dbPath string) (string, error) {
	if !MirrorAvailable() {
		return "", fmt.Errorf(
			"локальная копия Delivery не найдена (%s). Загрузите delivery_mirror.db на сервер",
			resolveMirrorPath(),
		)
	}

	scriptPath, err := resolveScriptPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("python3", scriptPath)
	mirrorPath := resolveMirrorPath()
	cmd.Env = append(os.Environ(),
		"DB_PATH="+dbPath,
		"DELIVERY_MIRROR_PATH="+mirrorPath,
	)
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
