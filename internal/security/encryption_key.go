// ==============================================================================
// MasterDnsVPN
// Author: MasterkinG32
// Github: https://github.com/masterking32
// Year: 2026
// ==============================================================================

package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"masterdnsvpn-go/internal/config"
)

// InsecureMethodWarning is non-nil when the configured method offers no real security.
var InsecureMethodWarning = map[int]string{
	0: "⚠ SECURITY WARNING: Encryption method NONE — all VPN traffic is plaintext!",
	1: "⚠ SECURITY WARNING: XOR cipher is NOT real encryption and is trivially breakable.",
}

// EncryptionKeyInfo holds the result of key loading or generation.
type EncryptionKeyInfo struct {
	MethodID   int
	MethodName string
	Key        string
	Path       string
	Loaded     bool
	Generated  bool
}

func EnsureServerEncryptionKey(cfg config.ServerConfig) (EncryptionKeyInfo, error) {
	info := EncryptionKeyInfo{
		MethodID:   cfg.DataEncryptionMethod,
		MethodName: EncryptionMethodName(cfg.DataEncryptionMethod),
		Path:       cfg.EncryptionKeyPath(),
	}

	// Warn about insecure methods.
	if warn, bad := InsecureMethodWarning[cfg.DataEncryptionMethod]; bad {
		fmt.Fprintln(os.Stderr, warn)
	}

	// Check parent directory permissions.
	if dir := filepath.Dir(info.Path); dir != "" {
		if st, err := os.Stat(dir); err == nil {
			if st.Mode().Perm()&0o022 != 0 {
				fmt.Fprintf(os.Stderr,
					"⚠ SECURITY WARNING: key directory %q is world/group writable (mode %04o). "+
						"Other users on this system may read or replace the encryption key.\n",
					dir, st.Mode().Perm())
			}
		}
	}

	requiredLength := requiredKeyLength(cfg.DataEncryptionMethod)
	raw, err := os.ReadFile(info.Path)
	if err == nil {
		key := strings.TrimSpace(string(raw))
		if len(key) == requiredLength {
			info.Key = key
			info.Loaded = true
			return info, nil
		}
	}

	key, err := generateHexText(requiredLength)
	if err != nil {
		return info, fmt.Errorf("generate encryption key: %w", err)
	}
	if err := os.WriteFile(info.Path, []byte(key), 0o600); err != nil {
		return info, fmt.Errorf("write encryption key file %s: %w", info.Path, err)
	}

	info.Key = key
	info.Generated = true
	return info, nil
}

func EncryptionMethodName(methodID int) string {
	switch methodID {
	case 0:
		return "NONE"
	case 1:
		return "XOR"
	case 2:
		return "ChaCha20"
	case 3:
		return "AES-128-GCM"
	case 4:
		return "AES-192-GCM"
	case 5:
		return "AES-256-GCM"
	default:
		return "UNKNOWN"
	}
}

func requiredKeyLength(methodID int) int {
	switch methodID {
	case 3:
		return 16
	case 4:
		return 24
	default:
		return 32
	}
}

func generateHexText(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}
	buf := make([]byte, (length+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf)[:length], nil
}
