package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

const (
	repo          = "bang9/burnshot"
	checkCooldown = 1 * time.Hour
)

type release struct {
	TagName string `json:"tag_name"`
}

// CheckAndUpgrade checks for a newer version and self-upgrades if found.
// Skips if checked within the last hour. Non-fatal on any error.
func CheckAndUpgrade(currentVersion string) {
	if currentVersion == "dev" {
		return
	}

	// Cooldown: skip if checked recently
	home, _ := os.UserHomeDir()
	if home != "" {
		checkFile := filepath.Join(home, ".burnshot", "last-check")
		if info, err := os.Stat(checkFile); err == nil {
			if time.Since(info.ModTime()) < checkCooldown {
				return
			}
		}
	}

	latest, err := latestVersion()

	// Touch check file after API call (even on same version / error)
	if home != "" {
		dir := filepath.Join(home, ".burnshot")
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "last-check"), nil, 0644)
	}

	if err != nil || latest == "" || latest == currentVersion {
		return
	}

	fmt.Fprintf(os.Stderr, " ↑ Upgrading burnshot %s → %s...\n", currentVersion, latest)

	if err := downloadAndReplace(latest); err != nil {
		fmt.Fprintf(os.Stderr, " ↑ Upgrade failed: %v (continuing with current version)\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, " ↑ Upgraded to %s — restarting...\n\n", latest)

	// Re-exec with new binary
	exe, _ := os.Executable()
	syscall.Exec(exe, os.Args, os.Environ())
}

func latestVersion() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.TagName, nil
}

func downloadAndReplace(version string) error {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/burnshot-%s", repo, version, platform)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download HTTP %d", resp.StatusCode)
	}

	// Get path of current binary
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	// Write to temp file next to binary
	tmp := exe + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	// Atomic rename
	if err := os.Rename(tmp, exe); err != nil {
		os.Remove(tmp)
		return err
	}

	return nil
}

