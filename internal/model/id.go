package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// JobID returns a stable identifier from source path, user, and command.
// Schedule is intentionally excluded so history survives schedule changes.
func JobID(sourcePath, user, command string) string {
	normalized := normalizeCommand(command)
	key := sourcePath + "\x00" + user + "\x00" + normalized
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}

func normalizeCommand(command string) string {
	cmd := strings.TrimSpace(command)
	if idx := strings.Index(cmd, "%"); idx >= 0 {
		cmd = cmd[:idx]
	}
	return strings.Join(strings.Fields(cmd), " ")
}
