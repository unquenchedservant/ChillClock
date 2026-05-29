package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func envFilePath() (string, error) {
	dir, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".env"), nil
}

func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func LoadAPIKey() (string, error) {
	path, err := envFilePath()
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "CHILLCLOCK_API_KEY=") {
			return strings.TrimPrefix(line, "CHILLCLOCK_API_KEY="), nil
		}
	}
	return "", scanner.Err()
}

func SaveAPIKey(key string) error {
	path, err := envFilePath()
	if err != nil {
		return err
	}
	var lines []string
	if f, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(strings.TrimSpace(line), "CHILLCLOCK_API_KEY") {
				lines = append(lines, line)
			}
		}
		f.Close()
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	lines = append(lines, fmt.Sprintf("CHILLCLOCK_API_KEY=%s", key))
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}
