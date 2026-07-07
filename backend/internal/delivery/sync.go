package delivery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func resolveScriptPath(name string) (string, error) {
	candidates := []string{
		filepath.Join("..", "scripts", name),
		filepath.Join("scripts", name),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s not found (tried %v)", name, candidates)
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

func resolveReferenceSeedPath() string {
	if path := strings.TrimSpace(os.Getenv("DELIVERY_REFERENCE_SEED_PATH")); path != "" {
		return path
	}
	candidates := []string{
		filepath.Join("..", "scripts", "delivery_reference_seed.sql"),
		filepath.Join("scripts", "delivery_reference_seed.sql"),
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

func ReferenceSeedAvailable() bool {
	_, err := os.Stat(resolveReferenceSeedPath())
	return err == nil
}

func SyncAvailable() bool {
	return MirrorAvailable() || ReferenceSeedAvailable()
}

func RunSync(dbPath string) (string, error) {
	if MirrorAvailable() {
		return runMirrorSync(dbPath)
	}
	if ReferenceSeedAvailable() {
		return runSeedApply(dbPath)
	}
	return "", fmt.Errorf(
		"нет источника Delivery: зеркало (%s) или seed (%s). На рабочей машине — pull зеркала; на VPS — upload delivery_reference_seed.sql",
		resolveMirrorPath(),
		resolveReferenceSeedPath(),
	)
}

func runMirrorSync(dbPath string) (string, error) {
	scriptPath, err := resolveScriptPath("sync_delivery_reference.py")
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
			return "", fmt.Errorf("delivery sync from mirror failed: %w", err)
		}
		return "", fmt.Errorf("delivery sync from mirror failed: %s", message)
	}
	return message, nil
}

func runSeedApply(dbPath string) (string, error) {
	scriptPath, err := resolveScriptPath("apply_delivery_reference.sh")
	if err != nil {
		return "", err
	}

	seedPath := resolveReferenceSeedPath()
	cmd := exec.Command("bash", scriptPath, "--online", seedPath)
	cmd.Env = append(os.Environ(), "DB_PATH="+dbPath)
	output, err := cmd.CombinedOutput()
	message := strings.TrimSpace(string(output))
	if err != nil {
		if message == "" {
			return "", fmt.Errorf("delivery seed apply failed: %w", err)
		}
		return "", fmt.Errorf("delivery seed apply failed: %s", message)
	}
	return message, nil
}
